package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
	"cchoice/cmd/web/models"
	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/httputil"
	"cchoice/internal/logs"
	"cchoice/internal/server/forms"
	"cchoice/internal/services"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

func (s *Server) adminCollectionReceiptsTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Table Handler]"
	ctx := r.Context()

	var q forms.AdminReceiptsTableQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	listPage := max(q.Page, 1)
	perPage := constants.DefaultAdminTablePageSize

	receipts, totalCount, err := s.services.collectionReceipt.GetCollectionReceiptsPaginated(ctx, listPage, perPage)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	listPage = models.ClampPage(listPage, totalCount, perPage)
	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       perPage,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/collection-receipts/table"),
		ContentTarget: "#collection-receipts-table-content",
	}

	if err := compadmin.AdminCollectionReceiptsTableContent(receipts, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
	}
}

func (s *Server) adminCollectionReceiptsGenerateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Generate Modal Handler]"
	ctx := r.Context()

	var q forms.AdminReceiptGenerateQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	var prefill *compadmin.CollectionReceiptPrefill
	if invoiceID := strings.TrimSpace(q.InvoiceID); invoiceID != "" {
		invoice, _, err := s.services.invoice.GetInvoice(ctx, invoiceID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
			return
		}
		p := buildCollectionReceiptPrefill(invoice)
		prefill = &p
	}

	if err := compadmin.CollectionReceiptGenerateModal(config, prefill).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminCollectionReceiptsCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Create Handler]"
	ctx := r.Context()

	var f forms.AdminCollectionReceiptCreateForm
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}

	in := services.CreateCollectionReceiptInput{
		InvoiceID:      strings.TrimSpace(f.InvoiceID),
		ReceivedFrom:   strings.TrimSpace(f.ReceivedFrom),
		RecipientEmail: strings.TrimSpace(f.RecipientEmail),
		TIN:            strings.TrimSpace(f.TIN),
		Address:        strings.TrimSpace(f.Address),
		ReceiptDate:    strings.TrimSpace(f.ReceiptDate),
		Amount:         strings.TrimSpace(f.Amount),
		PaymentFor:     strings.TrimSpace(f.PaymentFor),
		PaymentForm:    strings.TrimSpace(f.PaymentForm),
		SCCitizenTIN:   strings.TrimSpace(f.SCCitizenTIN),
		OSCAPWDIDNo:    strings.TrimSpace(f.OSCAPWDIDNo),
	}
	for _, st := range f.Settlements {
		in.Settlements = append(in.Settlements, services.CollectionReceiptSettlementInput{
			InvoiceID:     strings.TrimSpace(st.InvoiceID),
			InvoiceNumber: strings.TrimSpace(st.InvoiceNumber),
			Amount:        strings.TrimSpace(st.Amount),
		})
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	receipt, _, err := s.services.collectionReceipt.CreateCollectionReceipt(ctx, staffID, in)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	receiptDBID, err := s.services.collectionReceipt.GetCollectionReceiptDBID(ctx, receipt.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pdfStatus := "queued"
	if err := s.queueCollectionReceiptPDF(ctx, receiptDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		pdfStatus = "failed"
	}

	resp := map[string]any{
		"id":              receipt.ID,
		"receipt_number":  receipt.ReceiptNumber,
		"preview_url":     utils.URL(fmt.Sprintf("/admin/collection-receipts/%s/preview", receipt.ID)),
		"pdf_url":         utils.URL(fmt.Sprintf("/admin/collection-receipts/%s/pdf", receipt.ID)),
		"pdf_status":      pdfStatus,
	}
	if f.Action == "preview" {
		resp["open_preview"] = true
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) adminCollectionReceiptsViewHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts View Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	receipt, settlements, err := s.services.collectionReceipt.GetCollectionReceipt(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := s.services.collectionReceipt.BuildRenderData(config, receipt, settlements)
	if err := compadmin.AdminCollectionReceiptViewPage(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrRenderFailed.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminCollectionReceiptsPreviewModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Preview Modal Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	receipt, settlements, err := s.services.collectionReceipt.GetCollectionReceipt(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}
	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	data := s.services.collectionReceipt.BuildRenderData(config, receipt, settlements)
	if err := compadmin.CollectionReceiptPreviewModal(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminCollectionReceiptsPDFHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts PDF Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	receipt, settlements, err := s.services.collectionReceipt.GetCollectionReceipt(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if receipt.Status == enums.RECEIPT_STATUS_PROCESSING && strings.TrimSpace(receipt.PDFPath) == "" {
		http.Error(w, "PDF is still being generated", http.StatusConflict)
		return
	}

	receiptDBID, err := s.services.collectionReceipt.GetCollectionReceiptDBID(ctx, idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pdfBytes, err := s.services.collectionReceipt.ReadStoredPDF(receiptDBID, receipt.PDFPath)
	if err != nil {
		if s.collectionReceiptJobRunner != nil {
			if qerr := s.collectionReceiptJobRunner.QueueGeneratePDF(ctx, receiptDBID); qerr != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(qerr))
			}
			http.Error(w, "PDF is being generated", http.StatusConflict)
			return
		}
		config, cfgErr := s.services.invoice.GetConfig(ctx)
		if cfgErr != nil {
			http.Error(w, cfgErr.Error(), http.StatusInternalServerError)
			return
		}
		pdfBytes, err = services.RenderCollectionReceiptPDF(config, receipt, settlements)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, errs.ErrCollectionReceiptPDFFailed.Error(), http.StatusInternalServerError)
			return
		}
	}

	filename := receipt.ReceiptNumber
	if strings.TrimSpace(filename) == "" {
		filename = "collection_receipt"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, filename))
	if _, err := w.Write(pdfBytes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) adminCollectionReceiptsJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Job Status Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}
	receiptDBID, err := s.services.collectionReceipt.GetCollectionReceiptDBID(ctx, p.ID)
	if err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	status, err := s.services.collectionReceipt.GetJobStatus(ctx, receiptDBID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	receipt, _, err := s.services.collectionReceipt.GetCollectionReceipt(ctx, p.ID)
	if err != nil {
		writeInvoiceJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"pdf_status":   status.PDFStatus,
		"email_status": status.EmailStatus,
		"pdf_error":    status.PDFError,
		"email_error":  status.EmailError,
		"pdf_ready":    strings.TrimSpace(receipt.PDFPath) != "",
	})
}

func (s *Server) adminCollectionReceiptsEmailHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Collection Receipts Email Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	receiptDBID, err := s.services.collectionReceipt.GetCollectionReceiptDBID(ctx, idStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := s.queueCollectionReceiptEmail(ctx, staffID, receiptDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		respondInvoiceHX(w, r, adminInvoicesPage, "", err.Error())
		return
	}

	respondInvoiceHX(w, r, adminInvoicesPage, "Collection receipt email queued", "")
}

func (s *Server) queueCollectionReceiptPDF(ctx context.Context, receiptID int64) error {
	if s.collectionReceiptJobRunner != nil {
		return s.collectionReceiptJobRunner.QueueGeneratePDF(ctx, receiptID)
	}
	return s.services.collectionReceipt.GenerateAndStorePDF(ctx, receiptID)
}

func (s *Server) queueCollectionReceiptEmail(ctx context.Context, staffID string, receiptID int64) error {
	if s.collectionReceiptJobRunner != nil {
		return s.collectionReceiptJobRunner.QueueSendEmail(ctx, staffID, receiptID)
	}
	return s.services.collectionReceipt.SendCollectionReceiptEmailByID(ctx, staffID, receiptID)
}

func buildCollectionReceiptPrefill(invoice services.Invoice) compadmin.CollectionReceiptPrefill {
	receivedFrom := strings.TrimSpace(invoice.RecipientRegisteredName)
	if receivedFrom == "" {
		receivedFrom = invoice.RecipientName
	}
	currency := invoice.Currency
	if strings.TrimSpace(currency) == "" {
		currency = constants.PHP
	}
	amountDisplay, _ := utils.SchemaPrice(invoice.TotalSalesVATInclusiveRaw, currency)
	return compadmin.CollectionReceiptPrefill{
		InvoiceID: invoice.ID,
		InvoiceNumber: services.InvoiceSearchLabel(services.InvoiceListItem{
			InvoiceNumber:  invoice.InvoiceNumber,
			RecipientEmail: invoice.RecipientEmail,
			Total:          invoice.Total,
		}),
		ReceivedFrom:   receivedFrom,
		RecipientEmail: invoice.RecipientEmail,
		TIN:            invoice.RecipientTIN,
		Address:        invoice.RecipientAddress,
		Amount:         amountDisplay,
		Settlements: []services.CollectionReceiptSettlementInput{{
			InvoiceID:     invoice.ID,
			InvoiceNumber: invoice.InvoiceNumber,
			Amount:        amountDisplay,
		}},
	}
}

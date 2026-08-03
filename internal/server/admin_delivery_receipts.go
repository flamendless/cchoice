package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func (s *Server) adminDeliveryReceiptsTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Table Handler]"
	ctx := r.Context()

	var q forms.AdminReceiptsTableQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	listPage := max(q.Page, 1)
	perPage := constants.DefaultAdminTablePageSize

	receipts, totalCount, err := s.services.deliveryReceipt.GetDeliveryReceiptsPaginated(ctx, listPage, perPage)
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
		TableURL:      utils.URL("/admin/delivery-receipts/table"),
		ContentTarget: "#delivery-receipts-table-content",
	}

	if err := compadmin.AdminDeliveryReceiptsTableContent(receipts, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
	}
}

func (s *Server) adminDeliveryReceiptsGenerateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Generate Modal Handler]"
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

	var prefill *compadmin.DeliveryReceiptPrefill
	if invoiceID := strings.TrimSpace(q.InvoiceID); invoiceID != "" {
		invoice, lines, err := s.services.invoice.GetInvoice(ctx, invoiceID)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
			return
		}
		p := buildDeliveryReceiptPrefill(invoice, lines)
		prefill = &p
	}

	if err := compadmin.DeliveryReceiptGenerateModal(config, prefill).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminDeliveryReceiptsCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Create Handler]"
	ctx := r.Context()

	var f forms.AdminDeliveryReceiptCreateForm
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}

	in := services.CreateDeliveryReceiptInput{
		InvoiceID:      strings.TrimSpace(f.InvoiceID),
		DeliveredTo:    strings.TrimSpace(f.DeliveredTo),
		RecipientEmail: strings.TrimSpace(f.RecipientEmail),
		TIN:            strings.TrimSpace(f.TIN),
		Address:        strings.TrimSpace(f.Address),
		ReceiptDate:    strings.TrimSpace(f.ReceiptDate),
		Terms:          strings.TrimSpace(f.Terms),
		PONumber:       strings.TrimSpace(f.PONumber),
	}
	for _, l := range f.Lines {
		in.Lines = append(in.Lines, services.DeliveryReceiptLineInput{
			Quantity:    l.Quantity,
			Unit:        l.Unit,
			Description: l.Description,
		})
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	receipt, err := s.services.deliveryReceipt.CreateDeliveryReceipt(ctx, staffID, in)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	receiptDBID, err := s.services.deliveryReceipt.GetDeliveryReceiptDBID(ctx, receipt.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pdfStatus := "queued"
	if err := s.queueDeliveryReceiptPDF(ctx, staffID, receiptDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		pdfStatus = "failed"
	}

	resp := map[string]any{
		"id":              receipt.ID,
		"receipt_number":  receipt.ReceiptNumber,
		"preview_url":     utils.URL(fmt.Sprintf("/admin/delivery-receipts/%s/preview", receipt.ID)),
		"pdf_url":         utils.URL(fmt.Sprintf("/admin/delivery-receipts/%s/pdf", receipt.ID)),
		"pdf_status":      pdfStatus,
	}
	if f.Action == "preview" {
		resp["open_preview"] = true
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) adminDeliveryReceiptsViewHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts View Handler]"
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

	receipt, lines, err := s.services.deliveryReceipt.GetDeliveryReceipt(ctx, idStr)
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

	data := s.services.deliveryReceipt.BuildRenderData(config, receipt, lines)
	if err := compadmin.AdminDeliveryReceiptViewPage(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrRenderFailed.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminDeliveryReceiptsPreviewModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Preview Modal Handler]"
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

	receipt, lines, err := s.services.deliveryReceipt.GetDeliveryReceipt(ctx, idStr)
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

	data := s.services.deliveryReceipt.BuildRenderData(config, receipt, lines)
	if err := compadmin.DeliveryReceiptPreviewModal(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminDeliveryReceiptsPDFHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts PDF Handler]"
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

	receipt, lines, err := s.services.deliveryReceipt.GetDeliveryReceipt(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if receipt.Status == enums.RECEIPT_STATUS_PROCESSING && strings.TrimSpace(receipt.PDFPath) == "" {
		http.Error(w, "PDF is still being generated", http.StatusConflict)
		return
	}

	receiptDBID, err := s.services.deliveryReceipt.GetDeliveryReceiptDBID(ctx, idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pdfBytes, err := s.services.deliveryReceipt.ReadStoredPDF(receiptDBID, receipt.PDFPath)
	if err != nil {
		staffID := s.sessionManager.GetString(ctx, SessionStaffID)
		if s.deliveryReceiptJobRunner != nil {
			if qerr := s.deliveryReceiptJobRunner.QueueGeneratePDF(ctx, receiptDBID); qerr != nil {
				logs.LogCtx(ctx).Error(logtag, zap.Error(qerr))
			} else {
				s.services.deliveryReceipt.LogDeliveryReceiptPDFQueued(ctx, staffID, idStr)
			}
			http.Error(w, "PDF is being generated", http.StatusConflict)
			return
		}
		config, cfgErr := s.services.invoice.GetConfig(ctx)
		if cfgErr != nil {
			http.Error(w, cfgErr.Error(), http.StatusInternalServerError)
			return
		}
		pdfBytes, err = services.RenderDeliveryReceiptPDF(config, receipt, lines)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, errs.ErrDeliveryReceiptPDFFailed.Error(), http.StatusInternalServerError)
			return
		}
	}

	filename := receipt.ReceiptNumber
	if strings.TrimSpace(filename) == "" {
		filename = "delivery_receipt"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, filename))
	if _, err := w.Write(pdfBytes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		return
	}

	s.services.deliveryReceipt.LogDeliveryReceiptPDFDownload(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr)
}

func (s *Server) adminDeliveryReceiptsJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Job Status Handler]"
	ctx := r.Context()

	var p forms.AdminReceiptPath
	if err := httputil.BindPath(r, &p); err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}
	receiptDBID, err := s.services.deliveryReceipt.GetDeliveryReceiptDBID(ctx, p.ID)
	if err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	status, err := s.services.deliveryReceipt.GetJobStatus(ctx, receiptDBID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	receipt, _, err := s.services.deliveryReceipt.GetDeliveryReceipt(ctx, p.ID)
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

func (s *Server) adminDeliveryReceiptsEmailHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Delivery Receipts Email Handler]"
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

	receiptDBID, err := s.services.deliveryReceipt.GetDeliveryReceiptDBID(ctx, idStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := s.queueDeliveryReceiptEmail(ctx, staffID, receiptDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		respondInvoiceHX(w, r, adminInvoicesPage, "", err.Error())
		return
	}

	respondInvoiceHX(w, r, adminInvoicesPage, "Delivery receipt email queued", "")
}

func (s *Server) queueDeliveryReceiptPDF(ctx context.Context, staffID string, receiptID int64) error {
	if s.deliveryReceiptJobRunner != nil {
		if err := s.deliveryReceiptJobRunner.QueueGeneratePDF(ctx, receiptID); err != nil {
			return err
		}
		s.services.deliveryReceipt.LogDeliveryReceiptPDFQueued(ctx, staffID, s.encoder.Encode(receiptID))
		return nil
	}
	return s.services.deliveryReceipt.GenerateAndStorePDF(ctx, staffID, receiptID)
}

func (s *Server) queueDeliveryReceiptEmail(ctx context.Context, staffID string, receiptID int64) error {
	if s.deliveryReceiptJobRunner != nil {
		return s.deliveryReceiptJobRunner.QueueSendEmail(ctx, staffID, receiptID)
	}
	return s.services.deliveryReceipt.SendDeliveryReceiptEmailByID(ctx, staffID, receiptID)
}

func buildDeliveryReceiptPrefill(invoice services.Invoice, lines []services.InvoiceLine) compadmin.DeliveryReceiptPrefill {
	deliveredTo := strings.TrimSpace(invoice.RecipientRegisteredName)
	if deliveredTo == "" {
		deliveredTo = invoice.RecipientName
	}
	receiptDate := strings.TrimSpace(invoice.DeliveryDate)
	if receiptDate == "" {
		receiptDate = invoice.IssueDate
	}
	prefill := compadmin.DeliveryReceiptPrefill{
		InvoiceID:      invoice.ID,
		InvoiceNumber: services.InvoiceSearchLabel(services.InvoiceListItem{
			InvoiceNumber:  invoice.InvoiceNumber,
			RecipientEmail: invoice.RecipientEmail,
			Total:          invoice.Total,
		}),
		DeliveredTo:    deliveredTo,
		RecipientEmail: invoice.RecipientEmail,
		TIN:            invoice.RecipientTIN,
		Address:        invoice.RecipientAddress,
		ReceiptDate:    receiptDate,
	}
	if invoice.PaymentTermsValue > 0 && strings.TrimSpace(invoice.PaymentTermsUnit) != "" {
		u := enums.ParsePaymentTermsUnitToEnum(invoice.PaymentTermsUnit)
		if u != enums.PAYMENT_TERMS_UNDEFINED {
			prefill.Terms = strconv.FormatInt(invoice.PaymentTermsValue, 10) + " " + u.String()
		}
	}
	for _, ln := range lines {
		prefill.Lines = append(prefill.Lines, services.DeliveryReceiptLineInput{
			Quantity:    strconv.FormatInt(ln.Quantity, 10),
			Unit:        "pc",
			Description: ln.Description,
		})
	}
	return prefill
}

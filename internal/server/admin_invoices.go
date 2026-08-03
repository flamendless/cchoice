package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
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

const adminInvoicesPage = "/admin/invoices"

func (s *Server) adminInvoicesListPageHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices List Page Handler]"
	ctx := r.Context()

	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	access := compadmin.InvoicePageAccess{
		CanManageInvoices:           s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_INVOICES),
		CanCreateInvoice:            s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_INVOICES) || s.HasRole(ctx, enums.STAFF_ROLE_CREATE_INVOICE),
		CanAccessDeliveryReceipts:   s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_DELIVERY_RECEIPTS) || s.HasRole(ctx, enums.STAFF_ROLE_CREATE_DELIVERY_RECEIPT),
		CanCreateDeliveryReceipt:    s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_DELIVERY_RECEIPTS) || s.HasRole(ctx, enums.STAFF_ROLE_CREATE_DELIVERY_RECEIPT),
		CanAccessCollectionReceipts: s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_COLLECTION_RECEIPTS) || s.HasRole(ctx, enums.STAFF_ROLE_CREATE_COLLECTION_RECEIPT),
		CanCreateCollectionReceipt:  s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_COLLECTION_RECEIPTS) || s.HasRole(ctx, enums.STAFF_ROLE_CREATE_COLLECTION_RECEIPT),
	}
	if err := compadmin.AdminInvoicesListPage(config, access).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.String("path", r.URL.Path), zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}
}

func (s *Server) adminInvoicesConfigUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Config Update Handler]"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, httputil.ErrorMessage(errs.ErrInvalidParams)))
		return
	}

	existing, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	in := services.InvoiceConfigInput{
		BusinessName:        strings.TrimSpace(r.FormValue("business_name")),
		Address:             strings.TrimSpace(r.FormValue("address")),
		TIN:                 strings.TrimSpace(r.FormValue("tin")),
		VATRegistration:     strings.TrimSpace(r.FormValue("vat_registration")),
		Email:               strings.TrimSpace(r.FormValue("email")),
		ContactNumber:       strings.TrimSpace(r.FormValue("contact_number")),
		Website:             strings.TrimSpace(r.FormValue("website")),
		FooterNotes:         strings.TrimSpace(r.FormValue("footer_notes")),
		Currency:            strings.TrimSpace(r.FormValue("currency")),
		VATPercentage:       strings.TrimSpace(r.FormValue("vat_percentage")),
		ProprietorName:      strings.TrimSpace(r.FormValue("proprietor_name")),
		BIRBookletsInfo:     strings.TrimSpace(r.FormValue("bir_booklets_info")),
		BIRAuthorityToPrint: strings.TrimSpace(r.FormValue("bir_authority_to_print")),
		BIRDateIssued:       strings.TrimSpace(r.FormValue("bir_date_issued")),
		LogoURL:             existing.LogoURL,
		LogoPath:            existing.LogoPath,
	}

	if file, header, ferr := r.FormFile("logo"); ferr == nil {
		defer file.Close()
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, file); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrFileRead.Error()))
			return
		}
		contentType := header.Header.Get("Content-Type")
		url, localPath, err := s.services.image.UploadInvoiceLogo(ctx, filepath.Ext(header.Filename), &buf, contentType)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvoiceLogoInvalid.Error()))
			return
		}
		in.LogoURL = url
		in.LogoPath = localPath
	}

	if err := s.services.invoice.UpdateConfig(ctx, s.sessionManager.GetString(ctx, SessionStaffID), in); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminInvoicesPage, "Invoice configuration saved"))
}

func (s *Server) adminInvoicesConfigModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Config Modal Handler]"
	ctx := r.Context()

	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	if err := compadmin.InvoiceConfigModal(config).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminInvoicesRecipientsTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipients Table Handler]"
	ctx := r.Context()

	var q forms.AdminInvoiceRecipientsQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	listPage := max(q.Page, 1)
	perPage := constants.DefaultAdminTablePageSize

	recipients, totalCount, err := s.services.invoice.GetRecipientsPaginated(ctx, q.Search, listPage, perPage)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	listPage = models.ClampPage(listPage, totalCount, perPage)
	canManage := s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_INVOICES)
	pagination := models.TablePagination{
		Page:          listPage,
		PerPage:       perPage,
		TotalCount:    totalCount,
		TableURL:      utils.URL("/admin/invoices/recipients/table"),
		Include:       "[name='search']",
		ContentTarget: "#invoice-recipients-table-content",
	}

	if err := compadmin.AdminInvoiceRecipientsTableContent(recipients, q.Search, canManage, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
	}
}

func (s *Server) adminInvoicesRecipientCreateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipient Create Modal Handler]"
	ctx := r.Context()

	if err := compadmin.InvoiceRecipientCreateModal().Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
	}
}

func (s *Server) adminInvoicesRecipientCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipient Create Handler]"
	ctx := r.Context()

	var f forms.AdminInvoiceRecipientForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, httputil.ErrorMessage(err)))
		return
	}
	f.Normalize()

	if _, err := s.services.invoice.CreateRecipient(ctx, s.sessionManager.GetString(ctx, SessionStaffID), toRecipientInput(f)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminInvoicesPage, "Recipient created successfully"))
}

func (s *Server) adminInvoicesRecipientEditModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipient Edit Modal Handler]"
	ctx := r.Context()

	var p forms.AdminInvoiceRecipientPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	recipient, err := s.services.invoice.GetRecipientByID(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	if err := compadmin.InvoiceRecipientEditModal(recipient).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminInvoicesRecipientUpdateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipient Update Handler]"
	ctx := r.Context()

	var p forms.AdminInvoiceRecipientPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	var f forms.AdminInvoiceRecipientForm
	if err := httputil.BindForm(r, &f); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, httputil.ErrorMessage(err)))
		return
	}
	f.Normalize()

	if err := s.services.invoice.UpdateRecipient(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr, toRecipientInput(f)); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminInvoicesPage, "Recipient updated successfully"))
}

func (s *Server) adminInvoicesRecipientDeleteHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipient Delete Handler]"
	ctx := r.Context()

	var p forms.AdminInvoiceRecipientPath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	if err := s.services.invoice.DeleteRecipient(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminInvoicesPage, "Recipient deleted successfully"))
}

func (s *Server) adminInvoicesTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Table Handler]"
	ctx := r.Context()

	var q forms.AdminInvoicesTableQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	listPage := max(q.Page, 1)
	perPage := constants.DefaultAdminTablePageSize

	invoices, totalCount, err := s.services.invoice.GetInvoicesPaginated(ctx, listPage, perPage)
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
		TableURL:      utils.URL("/admin/invoices/table"),
		ContentTarget: "#invoices-table-content",
	}

	if err := compadmin.AdminInvoicesTableContent(invoices, pagination).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
	}
}

func (s *Server) adminInvoicesGenerateModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Generate Modal Handler]"
	ctx := r.Context()

	config, err := s.services.invoice.GetConfig(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	recipients, err := s.services.invoice.GetRecipients(ctx, "")
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	canManage := s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_INVOICES)
	if err := compadmin.InvoiceGenerateModal(config, recipients, canManage).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminInvoicesCreateHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Create Handler]"
	ctx := r.Context()

	var f forms.AdminInvoiceCreateForm
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}

	canManage := s.HasRole(ctx, enums.STAFF_ROLE_MANAGE_INVOICES)
	in := services.CreateInvoiceInput{
		RecipientID:             strings.TrimSpace(f.RecipientID),
		TransactionType:         strings.TrimSpace(f.TransactionType),
		RecipientRegisteredName: strings.TrimSpace(f.RecipientRegisteredName),
		Notes:                   strings.TrimSpace(f.Notes),
		IssueDate:               strings.TrimSpace(f.IssueDate),
		DeliveryDate:            strings.TrimSpace(f.DeliveryDate),
		DueDate:                 strings.TrimSpace(f.DueDate),
		PaymentTermsValue:       f.PaymentTermsValue,
		PaymentTermsUnit:        strings.TrimSpace(f.PaymentTermsUnit),
	}
	if canManage {
		in.WithholdingTax = strings.TrimSpace(f.WithholdingTax)
		in.SCPWDDiscount = strings.TrimSpace(f.SCPWDDiscount)
		in.AddVAT = strings.TrimSpace(f.AddVAT)
	}
	if nr := buildInvoiceNewRecipient(f); nr != nil {
		in.NewRecipient = nr
	}
	for _, l := range f.Lines {
		in.Lines = append(in.Lines, services.InvoiceLineInput{
			ProductID:   strings.TrimSpace(l.ProductID),
			Description: l.Description,
			UnitPrice:   l.UnitPrice,
			Quantity:    l.Quantity,
			TaxType:     strings.TrimSpace(l.TaxType),
		})
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	invoice, err := s.services.invoice.CreateInvoice(ctx, staffID, in)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	invoiceDBID, err := s.services.invoice.GetInvoiceDBID(ctx, invoice.ID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pdfStatus := "queued"
	if err := s.queueInvoicePDF(ctx, invoiceDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		pdfStatus = "failed"
	}

	resp := map[string]any{
		"id":                 invoice.ID,
		"invoice_number":     invoice.InvoiceNumber,
		"preview_url":        utils.URL(fmt.Sprintf("/admin/invoices/%s/preview", invoice.ID)),
		"pdf_url":            utils.URL(fmt.Sprintf("/admin/invoices/%s/pdf", invoice.ID)),
		"pdf_status":         pdfStatus,
		"created_recipient":  in.NewRecipient != nil,
	}
	if f.Action == "preview" {
		resp["open_preview"] = true
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) adminInvoicesViewHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices View Handler]"
	ctx := r.Context()

	var p forms.AdminInvoicePath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	invoice, lines, err := s.services.invoice.GetInvoice(ctx, idStr)
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

	data := s.services.invoice.BuildRenderData(config, invoice, lines)
	if err := compadmin.AdminInvoiceViewPage(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrRenderFailed.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminInvoicesPreviewModalHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Preview Modal Handler]"
	ctx := r.Context()

	var p forms.AdminInvoicePath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	invoice, lines, err := s.services.invoice.GetInvoice(ctx, idStr)
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

	data := s.services.invoice.BuildRenderData(config, invoice, lines)
	if err := compadmin.InvoicePreviewModal(data).Render(ctx, w); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrRenderFailed.Error()))
	}
}

func (s *Server) adminInvoicesPDFHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices PDF Handler]"
	ctx := r.Context()

	var p forms.AdminInvoicePath
	if err := httputil.BindPath(r, &p); err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		http.Error(w, errs.ErrInvalidParams.Error(), http.StatusBadRequest)
		return
	}

	invoice, _, err := s.services.invoice.GetInvoice(ctx, idStr)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if invoice.Status == enums.INVOICE_STATUS_PROCESSING && strings.TrimSpace(invoice.PDFPath) == "" {
		http.Error(w, "PDF is still being generated", http.StatusConflict)
		return
	}

	invoiceDBID, err := s.services.invoice.GetInvoiceDBID(ctx, idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pdfBytes, err := s.services.invoice.ReadStoredPDF(invoiceDBID, invoice.PDFPath)
	if err != nil {
		if s.invoiceJobRunner != nil {
			if qerr := s.invoiceJobRunner.QueueGeneratePDF(ctx, invoiceDBID); qerr != nil {
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
		_, lines, getErr := s.services.invoice.GetInvoice(ctx, idStr)
		if getErr != nil {
			http.Error(w, getErr.Error(), http.StatusInternalServerError)
			return
		}
		pdfBytes, err = services.RenderInvoicePDF(config, invoice, lines)
		if err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			http.Error(w, errs.ErrInvoicePDFFailed.Error(), http.StatusInternalServerError)
			return
		}
	}

	filename := invoice.InvoiceNumber
	if strings.TrimSpace(filename) == "" {
		filename = "invoice"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, filename))
	if _, err := w.Write(pdfBytes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
}

func (s *Server) adminInvoicesJobStatusHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Job Status Handler]"
	ctx := r.Context()

	var p forms.AdminInvoicePath
	if err := httputil.BindPath(r, &p); err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}
	invoiceDBID, err := s.services.invoice.GetInvoiceDBID(ctx, p.ID)
	if err != nil {
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	status, err := s.services.invoice.GetJobStatus(ctx, invoiceDBID)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	invoice, _, err := s.services.invoice.GetInvoice(ctx, p.ID)
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
		"pdf_ready":    strings.TrimSpace(invoice.PDFPath) != "",
	})
}

func (s *Server) adminInvoicesEmailHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Email Handler]"
	ctx := r.Context()

	var p forms.AdminInvoicePath
	if err := httputil.BindPath(r, &p); err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}
	idStr, err := httputil.RequireEncodedID(s.encoder, p.ID)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, errs.ErrInvalidParams.Error()))
		return
	}

	invoiceDBID, err := s.services.invoice.GetInvoiceDBID(ctx, idStr)
	if err != nil {
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	if err := s.queueInvoiceEmail(ctx, staffID, invoiceDBID); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		respondInvoiceHX(w, r, adminInvoicesPage, "", err.Error())
		return
	}

	respondInvoiceHX(w, r, adminInvoicesPage, "Invoice email queued", "")
}

func (s *Server) queueInvoicePDF(ctx context.Context, invoiceID int64) error {
	if s.invoiceJobRunner != nil {
		return s.invoiceJobRunner.QueueGeneratePDF(ctx, invoiceID)
	}
	return s.services.invoice.GenerateAndStorePDF(ctx, invoiceID)
}

func (s *Server) queueInvoiceEmail(ctx context.Context, staffID string, invoiceID int64) error {
	if s.invoiceJobRunner != nil {
		return s.invoiceJobRunner.QueueSendEmail(ctx, staffID, invoiceID)
	}
	return s.services.invoice.SendInvoiceEmailByID(ctx, staffID, invoiceID)
}

func toRecipientInput(f forms.AdminInvoiceRecipientForm) services.InvoiceRecipientInput {
	return services.InvoiceRecipientInput{
		Name:           f.Name,
		Email:          f.Email,
		ContactNumber:  f.ContactNumber,
		Address:        f.Address,
		TIN:            f.TIN,
		RegisteredName: f.RegisteredName,
		Notes:          f.Notes,
	}
}

func buildInvoiceNewRecipient(f forms.AdminInvoiceCreateForm) *services.InvoiceRecipientInput {
	if strings.TrimSpace(f.RecipientID) != "" {
		return nil
	}

	var nr forms.AdminInvoiceNewRecipientInput
	if f.NewRecipient != nil {
		nr = *f.NewRecipient
	}

	name := strings.TrimSpace(nr.Name)
	regName := strings.TrimSpace(f.RecipientRegisteredName)
	if regName == "" {
		regName = strings.TrimSpace(nr.RegisteredName)
	}
	if name == "" {
		name = regName
	}
	if regName == "" {
		regName = name
	}
	if name == "" {
		return nil
	}

	nr.Name = name
	nr.RegisteredName = regName
	nr.Normalize()

	return &services.InvoiceRecipientInput{
		Name:           name,
		Email:          strings.TrimSpace(nr.Email),
		ContactNumber:  nr.ContactNumber,
		Address:        strings.TrimSpace(nr.Address),
		TIN:            strings.TrimSpace(nr.TIN),
		RegisteredName: regName,
		Notes:          strings.TrimSpace(nr.Notes),
	}
}

func (s *Server) adminInvoicesProductsSearchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Products Search Handler]"
	ctx := r.Context()

	var q forms.AdminInvoiceEntitySearchQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}

	products, err := s.services.invoice.SearchProductsForLineItems(ctx, q.Q, 20)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]map[string]any, 0, len(products))
	for _, p := range products {
		items = append(items, map[string]any{
			"id":             p.ID,
			"label":          services.ProductOptionLabel(p),
			"name":           p.Name,
			"serial":         p.Serial,
			"price_centavos": p.UnitPrice,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func (s *Server) adminInvoicesSearchHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Search Handler]"
	ctx := r.Context()

	var q forms.AdminInvoiceEntitySearchQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, errs.ErrInvalidParams.Error())
		return
	}

	invoices, err := s.services.invoice.SearchInvoices(ctx, q.Q, 20)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]map[string]any, 0, len(invoices))
	for _, inv := range invoices {
		items = append(items, map[string]any{
			"id":             inv.ID,
			"label":          services.InvoiceSearchLabel(inv),
			"invoice_number": inv.InvoiceNumber,
			"total":          inv.Total,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
}

func writeInvoiceJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondInvoiceHX(w http.ResponseWriter, r *http.Request, page, successMsg, errorMsg string) {
	if isHTMX(r) {
		if errorMsg != "" {
			w.Header().Set("X-Error-Message", errorMsg)
			http.Error(w, errorMsg, http.StatusBadRequest)
			return
		}
		if successMsg != "" {
			w.Header().Set("X-Success-Message", successMsg)
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if errorMsg != "" {
		redirectHX(w, r, utils.URLWithError(page, errorMsg))
		return
	}
	redirectHX(w, r, utils.URLWithSuccess(page, successMsg))
}

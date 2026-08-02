package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	compadmin "cchoice/cmd/web/components/admin"
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

	if err := compadmin.AdminInvoicesListPage(config).Render(ctx, w); err != nil {
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
		BusinessName:    strings.TrimSpace(r.FormValue("business_name")),
		Address:         strings.TrimSpace(r.FormValue("address")),
		TIN:             strings.TrimSpace(r.FormValue("tin")),
		VATRegistration: strings.TrimSpace(r.FormValue("vat_registration")),
		Email:           strings.TrimSpace(r.FormValue("email")),
		ContactNumber:   strings.TrimSpace(r.FormValue("contact_number")),
		Website:         strings.TrimSpace(r.FormValue("website")),
		FooterNotes:     strings.TrimSpace(r.FormValue("footer_notes")),
		Currency:        strings.TrimSpace(r.FormValue("currency")),
		VATPercentage:   strings.TrimSpace(r.FormValue("vat_percentage")),
		LogoURL:         existing.LogoURL,
		LogoPath:        existing.LogoPath,
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

func (s *Server) adminInvoicesRecipientsTableHandler(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Admin Invoices Recipients Table Handler]"
	ctx := r.Context()

	var q forms.AdminInvoiceRecipientsQuery
	if err := httputil.BindQuery(r, &q); err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
	}

	recipients, err := s.services.invoice.GetRecipients(ctx, q.Search)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	if err := compadmin.AdminInvoiceRecipientsTable(recipients, q.Search).Render(ctx, w); err != nil {
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

	invoices, err := s.services.invoice.GetAllInvoices(ctx)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	if err := compadmin.AdminInvoicesTable(invoices).Render(ctx, w); err != nil {
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

	products, err := s.services.invoice.ListProductsForLineItems(ctx)
	if err != nil {
		logs.LogCtx(ctx).Warn(logtag, zap.Error(err))
		products = nil
	}

	if err := compadmin.InvoiceGenerateModal(config, recipients, products).Render(ctx, w); err != nil {
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

	in := services.CreateInvoiceInput{
		RecipientID: strings.TrimSpace(f.RecipientID),
		Notes:       strings.TrimSpace(f.Notes),
		DueDate:     strings.TrimSpace(f.DueDate),
	}
	if f.NewRecipient != nil && strings.TrimSpace(f.NewRecipient.Name) != "" {
		in.NewRecipient = &services.InvoiceRecipientInput{
			Name:          f.NewRecipient.Name,
			Email:         f.NewRecipient.Email,
			ContactNumber: f.NewRecipient.ContactNumber,
			Address:       f.NewRecipient.Address,
			TIN:           f.NewRecipient.TIN,
			Notes:         f.NewRecipient.Notes,
		}
	}
	for _, l := range f.Lines {
		in.Lines = append(in.Lines, services.InvoiceLineInput{
			ProductID:   strings.TrimSpace(l.ProductID),
			Description: l.Description,
			UnitPrice:   l.UnitPrice,
			Quantity:    l.Quantity,
		})
	}

	staffID := s.sessionManager.GetString(ctx, SessionStaffID)
	invoice, err := s.services.invoice.CreateInvoice(ctx, staffID, in)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		writeInvoiceJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	emailed := false
	var emailErr string
	if f.Action == "email" {
		if err := s.services.invoice.SendInvoiceEmail(ctx, staffID, invoice.ID); err != nil {
			logs.LogCtx(ctx).Error(logtag, zap.Error(err))
			emailErr = err.Error()
		} else {
			emailed = true
		}
	}

	resp := map[string]any{
		"id":             invoice.ID,
		"invoice_number": invoice.InvoiceNumber,
		"pdf_url":        utils.URL(fmt.Sprintf("/admin/invoices/%s/pdf", invoice.ID)),
		"emailed":        emailed,
	}
	if emailErr != "" {
		resp["email_error"] = emailErr
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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

	pdfBytes, err := services.RenderInvoicePDF(config, invoice, lines)
	if err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		http.Error(w, errs.ErrInvoicePDFFailed.Error(), http.StatusInternalServerError)
		return
	}

	filename := invoice.InvoiceNumber
	if strings.TrimSpace(filename) == "" {
		filename = "invoice"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", filename))
	if _, err := w.Write(pdfBytes); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
	}
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

	if err := s.services.invoice.SendInvoiceEmail(ctx, s.sessionManager.GetString(ctx, SessionStaffID), idStr); err != nil {
		logs.LogCtx(ctx).Error(logtag, zap.Error(err))
		redirectHX(w, r, utils.URLWithError(adminInvoicesPage, err.Error()))
		return
	}

	redirectHX(w, r, utils.URLWithSuccess(adminInvoicesPage, "Invoice emailed successfully"))
}

func toRecipientInput(f forms.AdminInvoiceRecipientForm) services.InvoiceRecipientInput {
	return services.InvoiceRecipientInput{
		Name:          f.Name,
		Email:         f.Email,
		ContactNumber: f.ContactNumber,
		Address:       f.Address,
		TIN:           f.TIN,
		Notes:         f.Notes,
	}
}

func writeInvoiceJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

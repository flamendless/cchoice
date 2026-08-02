package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"cchoice/internal/constants"
	"cchoice/internal/database"
	"cchoice/internal/database/queries"
	"cchoice/internal/encode"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
	"cchoice/internal/logs"
	"cchoice/internal/mail"
	"cchoice/internal/utils"

	"go.uber.org/zap"
)

type InvoiceConfig struct {
	BusinessName    string
	Address         string
	TIN             string
	VATRegistration string
	Email           string
	ContactNumber   string
	Website         string
	FooterNotes     string
	LogoURL         string
	LogoPath        string
	Currency        string
	VATPercentage   string
}

type InvoiceConfigInput struct {
	BusinessName    string
	Address         string
	TIN             string
	VATRegistration string
	Email           string
	ContactNumber   string
	Website         string
	FooterNotes     string
	Currency        string
	VATPercentage   string
	LogoURL         string
	LogoPath        string
}

type InvoiceRecipient struct {
	ID            string
	Name          string
	Email         string
	ContactNumber string
	Address       string
	TIN           string
	Notes         string
}

type InvoiceRecipientInput struct {
	Name          string
	Email         string
	ContactNumber string
	Address       string
	TIN           string
	Notes         string
}

type InvoiceProductOption struct {
	ID        string
	Name      string
	Serial    string
	BrandName string
	UnitPrice int64
	Currency  string
	Price     string
}

// InvoiceLineInput describes one requested line. When ProductID is set and
// UnitPrice is empty, the product's price and name are resolved from the DB.
type InvoiceLineInput struct {
	ProductID   string
	Description string
	UnitPrice   string
	Quantity    int64
}

type CreateInvoiceInput struct {
	RecipientID  string
	NewRecipient *InvoiceRecipientInput
	Notes        string
	DueDate      string
	Lines        []InvoiceLineInput
}

type InvoiceLine struct {
	Description  string
	UnitPrice    string
	LineTotal    string
	ProductID    sql.NullInt64
	UnitPriceRaw int64
	LineTotalRaw int64
	Quantity     int64
}

type Invoice struct {
	ID                     string
	InvoiceNumber          string
	Status                 enums.InvoiceStatus
	RecipientName          string
	RecipientEmail         string
	RecipientContactNumber string
	RecipientAddress       string
	RecipientTIN           string
	IssueDate              string
	DueDate                string
	Notes                  string
	Currency               string
	VATPercentage          string
	Subtotal               string
	VATAmount              string
	Total                  string
	EmailedAt              string
	CreatedAt              string
	SubtotalRaw            int64
	VATAmountRaw           int64
	TotalRaw               int64
}

type InvoiceListItem struct {
	ID            string
	InvoiceNumber string
	RecipientName string
	Status        enums.InvoiceStatus
	IssueDate     string
	Total         string
	Emailed       bool
	CreatedAt     string
}

type InvoiceService struct {
	encoder     encode.IEncode
	dbRO        database.IService
	dbRW        database.IService
	staffLog    *StaffLogsService
	mailService mail.IMailService
}

func NewInvoiceService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
	mailService mail.IMailService,
) *InvoiceService {
	return &InvoiceService{
		encoder:     encoder,
		dbRO:        dbRO,
		dbRW:        dbRW,
		staffLog:    staffLog,
		mailService: mailService,
	}
}

func (s *InvoiceService) GetConfig(ctx context.Context) (InvoiceConfig, error) {
	row, err := s.dbRO.GetQueries().GetInvoiceConfig(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvoiceConfig{Currency: constants.PHP, VATPercentage: "12"}, nil
		}
		return InvoiceConfig{}, errors.Join(errs.ErrInvoice, err)
	}
	c := row.TblInvoiceConfig
	return InvoiceConfig{
		BusinessName:    c.BusinessName,
		Address:         c.Address,
		TIN:             c.Tin,
		VATRegistration: c.VatRegistration,
		Email:           c.Email,
		ContactNumber:   c.ContactNumber,
		Website:         c.Website,
		FooterNotes:     c.FooterNotes,
		LogoURL:         c.LogoUrl,
		LogoPath:        c.LogoPath,
		Currency:        cmpOr(c.Currency, constants.PHP),
		VATPercentage:   cmpOr(c.VatPercentage, "0"),
	}, nil
}

func (s *InvoiceService) UpdateConfig(ctx context.Context, staffID string, in InvoiceConfigInput) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionUpdate, constants.ModuleInvoiceConfig, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] update config log", zap.Error(err))
		}
	}()

	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = constants.PHP
	}
	vat := strings.TrimSpace(in.VATPercentage)
	if vat == "" {
		vat = "0"
	}

	if err := s.dbRW.GetQueries().UpsertInvoiceConfig(ctx, queries.UpsertInvoiceConfigParams{
		BusinessName:    in.BusinessName,
		Address:         in.Address,
		Tin:             in.TIN,
		VatRegistration: in.VATRegistration,
		Email:           in.Email,
		ContactNumber:   in.ContactNumber,
		Website:         in.Website,
		FooterNotes:     in.FooterNotes,
		LogoUrl:         in.LogoURL,
		LogoPath:        in.LogoPath,
		Currency:        currency,
		VatPercentage:   vat,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoice, err)
	}
	return nil
}

func (s *InvoiceService) GetRecipients(ctx context.Context, search string) ([]InvoiceRecipient, error) {
	rows, err := s.dbRO.GetQueries().GetAllInvoiceRecipients(ctx, search)
	if err != nil {
		return nil, errors.Join(errs.ErrInvoice, err)
	}
	result := make([]InvoiceRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.toRecipient(row.TblInvoiceRecipient))
	}
	return result, nil
}

func (s *InvoiceService) GetRecipientByID(ctx context.Context, id string) (InvoiceRecipient, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return InvoiceRecipient{}, errs.ErrDecode
	}
	row, err := s.dbRO.GetQueries().GetInvoiceRecipientByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvoiceRecipient{}, errs.ErrInvoiceRecipientNotFound
		}
		return InvoiceRecipient{}, errors.Join(errs.ErrInvoice, err)
	}
	return s.toRecipient(row.TblInvoiceRecipient), nil
}

func (s *InvoiceService) toRecipient(r queries.TblInvoiceRecipient) InvoiceRecipient {
	return InvoiceRecipient{
		ID:            s.encoder.Encode(r.ID),
		Name:          r.Name,
		Email:         r.Email,
		ContactNumber: r.ContactNumber,
		Address:       r.Address,
		TIN:           r.Tin,
		Notes:         r.Notes,
	}
}

func (s *InvoiceService) CreateRecipient(ctx context.Context, staffID string, in InvoiceRecipientInput) (string, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionCreate, constants.ModuleInvoiceRecipients, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] create recipient log", zap.Error(err))
		}
	}()

	if strings.TrimSpace(in.Name) == "" {
		result = errs.ErrInvoiceRecipientNameReq.Error()
		return "", errs.ErrInvoiceRecipientNameReq
	}

	id, err := s.createRecipient(ctx, in)
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrInvoice, err)
	}
	idStr := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", idStr)
	return idStr, nil
}

func (s *InvoiceService) createRecipient(ctx context.Context, in InvoiceRecipientInput) (int64, error) {
	return s.dbRW.GetQueries().CreateInvoiceRecipient(ctx, queries.CreateInvoiceRecipientParams{
		Name:          strings.TrimSpace(in.Name),
		Email:         strings.TrimSpace(in.Email),
		ContactNumber: strings.TrimSpace(in.ContactNumber),
		Address:       strings.TrimSpace(in.Address),
		Tin:           strings.TrimSpace(in.TIN),
		Notes:         strings.TrimSpace(in.Notes),
	})
}

func (s *InvoiceService) UpdateRecipient(ctx context.Context, staffID string, id string, in InvoiceRecipientInput) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionUpdate, constants.ModuleInvoiceRecipients, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] update recipient log", zap.Error(err))
		}
	}()

	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}
	if strings.TrimSpace(in.Name) == "" {
		result = errs.ErrInvoiceRecipientNameReq.Error()
		return errs.ErrInvoiceRecipientNameReq
	}

	if err := s.dbRW.GetQueries().UpdateInvoiceRecipient(ctx, queries.UpdateInvoiceRecipientParams{
		ID:            decoded,
		Name:          strings.TrimSpace(in.Name),
		Email:         strings.TrimSpace(in.Email),
		ContactNumber: strings.TrimSpace(in.ContactNumber),
		Address:       strings.TrimSpace(in.Address),
		Tin:           strings.TrimSpace(in.TIN),
		Notes:         strings.TrimSpace(in.Notes),
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoice, err)
	}
	return nil
}

func (s *InvoiceService) DeleteRecipient(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionDelete, constants.ModuleInvoiceRecipients, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] delete recipient log", zap.Error(err))
		}
	}()

	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		result = errs.ErrDecode.Error()
		return errs.ErrDecode
	}
	if err := s.dbRW.GetQueries().SoftDeleteInvoiceRecipient(ctx, decoded); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoice, err)
	}
	return nil
}

func (s *InvoiceService) ListProductsForLineItems(ctx context.Context) ([]InvoiceProductOption, error) {
	rows, err := s.dbRO.GetQueries().ListProductsForQuotations(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrInvoice, err)
	}
	result := make([]InvoiceProductOption, 0, len(rows))
	for _, row := range rows {
		price := row.UnitPriceWithVat
		currency := row.UnitPriceWithVatCurrency
		if row.IsOnSale == 1 && row.SalePriceWithVat.Valid {
			price = row.SalePriceWithVat.Int64
			if row.SalePriceWithVatCurrency.Valid {
				currency = row.SalePriceWithVatCurrency.String
			}
		}
		result = append(result, InvoiceProductOption{
			ID:        s.encoder.Encode(row.ID),
			Name:      row.Name,
			Serial:    row.Serial,
			BrandName: row.BrandName,
			UnitPrice: price,
			Currency:  currency,
			Price:     utils.NewMoney(price, currency).Display(),
		})
	}
	return result, nil
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, staffID string, in CreateInvoiceInput) (Invoice, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionCreate, constants.ModuleInvoices, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] create invoice log", zap.Error(err))
		}
	}()

	if len(in.Lines) == 0 {
		result = errs.ErrInvoiceNoLines.Error()
		return Invoice{}, errs.ErrInvoiceNoLines
	}

	config, err := s.GetConfig(ctx)
	if err != nil {
		result = err.Error()
		return Invoice{}, err
	}
	currency := cmpOr(config.Currency, constants.PHP)

	recipient, err := s.resolveRecipient(ctx, in)
	if err != nil {
		result = err.Error()
		return Invoice{}, err
	}

	lines, subtotal, err := s.resolveLines(ctx, in.Lines, currency)
	if err != nil {
		result = err.Error()
		return Invoice{}, err
	}

	vatPct, _ := strconv.ParseFloat(cmpOr(config.VATPercentage, "0"), 64)
	vatAmount := int64(math.Round(float64(subtotal) * vatPct / 100.0))
	total := subtotal + vatAmount

	staffDBID := s.encoder.Decode(staffID)
	if staffDBID == encode.INVALID {
		staffDBID = 0
	}

	recipientID := sql.NullInt64{}
	if recipient.dbID != 0 {
		recipientID = sql.NullInt64{Int64: recipient.dbID, Valid: true}
	}

	issueDate := utils.NowPH().Format(constants.DateLayoutISO)
	invoiceDBID, err := s.dbRW.GetQueries().CreateInvoice(ctx, queries.CreateInvoiceParams{
		InvoiceNumber:          "",
		RecipientID:            recipientID,
		RecipientName:          recipient.Name,
		RecipientEmail:         recipient.Email,
		RecipientContactNumber: recipient.ContactNumber,
		RecipientAddress:       recipient.Address,
		RecipientTin:           recipient.TIN,
		Status:                 enums.INVOICE_STATUS_ISSUED.String(),
		IssueDate:              issueDate,
		DueDate:                strings.TrimSpace(in.DueDate),
		Notes:                  strings.TrimSpace(in.Notes),
		Currency:               currency,
		Subtotal:               subtotal,
		VatPercentage:          cmpOr(config.VATPercentage, "0"),
		VatAmount:              vatAmount,
		Total:                  total,
		CreatedBy:              staffDBID,
	})
	if err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
	}

	invoiceNumber := fmt.Sprintf("INV-%s-%05d", strings.ReplaceAll(issueDate, "-", ""), invoiceDBID)
	if err := s.dbRW.GetQueries().SetInvoiceNumber(ctx, queries.SetInvoiceNumberParams{
		InvoiceNumber: invoiceNumber,
		ID:            invoiceDBID,
	}); err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
	}

	for _, ln := range lines {
		if err := s.dbRW.GetQueries().CreateInvoiceLine(ctx, queries.CreateInvoiceLineParams{
			InvoiceID:   invoiceDBID,
			ProductID:   ln.ProductID,
			Description: ln.Description,
			Quantity:    ln.Quantity,
			UnitPrice:   ln.UnitPriceRaw,
			LineTotal:   ln.LineTotalRaw,
			Currency:    currency,
		}); err != nil {
			result = err.Error()
			return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
		}
	}

	idStr := s.encoder.Encode(invoiceDBID)
	result = fmt.Sprintf("success. ID '%s' number '%s'", idStr, invoiceNumber)

	invoice, _, err := s.GetInvoice(ctx, idStr)
	if err != nil {
		return Invoice{}, err
	}
	return invoice, nil
}

type resolvedRecipient struct {
	Name          string
	Email         string
	ContactNumber string
	Address       string
	TIN           string
	dbID          int64
}

func (s *InvoiceService) resolveRecipient(ctx context.Context, in CreateInvoiceInput) (resolvedRecipient, error) {
	if in.NewRecipient != nil && strings.TrimSpace(in.NewRecipient.Name) != "" {
		id, err := s.createRecipient(ctx, *in.NewRecipient)
		if err != nil {
			return resolvedRecipient{}, errors.Join(errs.ErrInvoice, err)
		}
		return resolvedRecipient{
			dbID:          id,
			Name:          strings.TrimSpace(in.NewRecipient.Name),
			Email:         strings.TrimSpace(in.NewRecipient.Email),
			ContactNumber: strings.TrimSpace(in.NewRecipient.ContactNumber),
			Address:       strings.TrimSpace(in.NewRecipient.Address),
			TIN:           strings.TrimSpace(in.NewRecipient.TIN),
		}, nil
	}

	if strings.TrimSpace(in.RecipientID) == "" {
		return resolvedRecipient{}, errs.ErrInvoiceRecipientRequired
	}

	r, err := s.GetRecipientByID(ctx, in.RecipientID)
	if err != nil {
		return resolvedRecipient{}, err
	}
	return resolvedRecipient{
		dbID:          s.encoder.Decode(r.ID),
		Name:          r.Name,
		Email:         r.Email,
		ContactNumber: r.ContactNumber,
		Address:       r.Address,
		TIN:           r.TIN,
	}, nil
}

type resolvedLine struct {
	Description  string
	ProductID    sql.NullInt64
	Quantity     int64
	UnitPriceRaw int64
	LineTotalRaw int64
}

func (s *InvoiceService) resolveLines(ctx context.Context, inputs []InvoiceLineInput, currency string) ([]resolvedLine, int64, error) {
	lines := make([]resolvedLine, 0, len(inputs))
	var subtotal int64
	for _, li := range inputs {
		qty := li.Quantity
		if qty <= 0 {
			qty = 1
		}

		description := strings.TrimSpace(li.Description)
		var unitPrice int64
		productID := sql.NullInt64{}

		if strings.TrimSpace(li.ProductID) != "" {
			decoded := s.encoder.Decode(li.ProductID)
			if decoded != encode.INVALID {
				productID = sql.NullInt64{Int64: decoded, Valid: true}
				product, err := s.dbRO.GetQueries().GetProductsByID(ctx, decoded)
				if err == nil {
					if description == "" {
						description = product.Name
					}
					unitPrice = product.UnitPriceWithVat
				}
			}
		}

		if strings.TrimSpace(li.UnitPrice) != "" {
			m, err := utils.NewMoneyFromString(strings.ReplaceAll(strings.TrimSpace(li.UnitPrice), ",", ""), currency)
			if err == nil {
				unitPrice = m.Amount()
			}
		}

		if description == "" {
			continue
		}

		lineTotal := unitPrice * qty
		subtotal += lineTotal
		lines = append(lines, resolvedLine{
			Description:  description,
			ProductID:    productID,
			Quantity:     qty,
			UnitPriceRaw: unitPrice,
			LineTotalRaw: lineTotal,
		})
	}

	if len(lines) == 0 {
		return nil, 0, errs.ErrInvoiceNoLines
	}
	return lines, subtotal, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, id string) (Invoice, []InvoiceLine, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return Invoice{}, nil, errs.ErrDecode
	}

	row, err := s.dbRO.GetQueries().GetInvoiceByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Invoice{}, nil, errs.ErrInvoiceNotFound
		}
		return Invoice{}, nil, errors.Join(errs.ErrInvoice, err)
	}
	inv := row.TblInvoice
	currency := cmpOr(inv.Currency, constants.PHP)

	lineRows, err := s.dbRO.GetQueries().GetInvoiceLinesByInvoiceID(ctx, decoded)
	if err != nil {
		return Invoice{}, nil, errors.Join(errs.ErrInvoice, err)
	}
	lines := make([]InvoiceLine, 0, len(lineRows))
	for _, lr := range lineRows {
		l := lr.TblInvoiceLine
		lines = append(lines, InvoiceLine{
			Description:  l.Description,
			Quantity:     l.Quantity,
			ProductID:    l.ProductID,
			UnitPriceRaw: l.UnitPrice,
			LineTotalRaw: l.LineTotal,
			UnitPrice:    utils.NewMoney(l.UnitPrice, currency).Display(),
			LineTotal:    utils.NewMoney(l.LineTotal, currency).Display(),
		})
	}

	invoice := Invoice{
		ID:                     s.encoder.Encode(inv.ID),
		InvoiceNumber:          inv.InvoiceNumber,
		Status:                 enums.ParseInvoiceStatusToEnum(inv.Status),
		RecipientName:          inv.RecipientName,
		RecipientEmail:         inv.RecipientEmail,
		RecipientContactNumber: inv.RecipientContactNumber,
		RecipientAddress:       inv.RecipientAddress,
		RecipientTIN:           inv.RecipientTin,
		IssueDate:              inv.IssueDate,
		DueDate:                inv.DueDate,
		Notes:                  inv.Notes,
		Currency:               currency,
		VATPercentage:          inv.VatPercentage,
		Subtotal:               utils.NewMoney(inv.Subtotal, currency).Display(),
		VATAmount:              utils.NewMoney(inv.VatAmount, currency).Display(),
		Total:                  utils.NewMoney(inv.Total, currency).Display(),
		EmailedAt:              inv.EmailedAt,
		CreatedAt:              inv.CreatedAt,
		SubtotalRaw:            inv.Subtotal,
		VATAmountRaw:           inv.VatAmount,
		TotalRaw:               inv.Total,
	}
	return invoice, lines, nil
}

func (s *InvoiceService) GetAllInvoices(ctx context.Context) ([]InvoiceListItem, error) {
	rows, err := s.dbRO.GetQueries().GetAllInvoices(ctx)
	if err != nil {
		return nil, errors.Join(errs.ErrInvoice, err)
	}
	result := make([]InvoiceListItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, InvoiceListItem{
			ID:            s.encoder.Encode(r.ID),
			InvoiceNumber: r.InvoiceNumber,
			RecipientName: r.RecipientName,
			Status:        enums.ParseInvoiceStatusToEnum(r.Status),
			IssueDate:     r.IssueDate,
			Total:         utils.NewMoney(r.Total, cmpOr(r.Currency, constants.PHP)).Display(),
			Emailed:       r.EmailedAt != "",
			CreatedAt:     r.CreatedAt,
		})
	}
	return result, nil
}

// SendInvoiceEmail renders the invoice to a PDF and emails it to the invoice
// recipient with the PDF attached.
func (s *InvoiceService) SendInvoiceEmail(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionSend, constants.ModuleInvoices, result, nil); err != nil {
			logs.Log().Warn("[InvoiceService] send invoice log", zap.Error(err))
		}
	}()

	if s.mailService == nil {
		result = errs.ErrInvoiceEmailNotConfigured.Error()
		return errs.ErrInvoiceEmailNotConfigured
	}

	invoice, lines, err := s.GetInvoice(ctx, id)
	if err != nil {
		result = err.Error()
		return err
	}
	if strings.TrimSpace(invoice.RecipientEmail) == "" {
		result = errs.ErrInvoiceRecipientNoEmail.Error()
		return errs.ErrInvoiceRecipientNoEmail
	}

	config, err := s.GetConfig(ctx)
	if err != nil {
		result = err.Error()
		return err
	}

	pdfBytes, err := RenderInvoicePDF(config, invoice, lines)
	if err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoicePDFFailed, err)
	}

	businessName := cmpOr(config.BusinessName, "C-Choice")
	subject := fmt.Sprintf("Invoice %s from %s", invoice.InvoiceNumber, businessName)

	lineItems := make([]map[string]any, 0, len(lines))
	for _, l := range lines {
		lineItems = append(lineItems, map[string]any{
			"Description": l.Description,
			"Quantity":    l.Quantity,
			"UnitPrice":   l.UnitPrice,
			"LineTotal":   l.LineTotal,
		})
	}

	data := mail.TemplateData{
		"LogoURL":       cmpOr(config.LogoURL, constants.PathEmailLogoCDN),
		"BusinessName":  businessName,
		"InvoiceNumber": invoice.InvoiceNumber,
		"RecipientName": invoice.RecipientName,
		"IssueDate":     invoice.IssueDate,
		"DueDate":       invoice.DueDate,
		"LineItems":     lineItems,
		"Subtotal":      invoice.Subtotal,
		"VATAmount":     invoice.VATAmount,
		"VATPercentage": invoice.VATPercentage,
		"Total":         invoice.Total,
		"Notes":         invoice.Notes,
		"MobileNo":      config.ContactNumber,
		"EMail":         config.Email,
	}

	attachments := []mail.Attachment{
		{
			FileName:    fmt.Sprintf("%s.pdf", cmpOr(invoice.InvoiceNumber, "invoice")),
			ContentType: "application/pdf",
			Content:     pdfBytes,
		},
	}

	if err := s.mailService.SendTemplateEmailWithAttachments(
		invoice.RecipientEmail,
		nil,
		subject,
		enums.EMAIL_TEMPLATE_INVOICE.FileName(),
		data,
		attachments,
	); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoiceEmailFailed, err)
	}

	if err := s.dbRW.GetQueries().MarkInvoiceEmailed(ctx, s.encoder.Decode(id)); err != nil {
		logs.LogCtx(ctx).Warn("[InvoiceService] mark emailed", zap.Error(err))
	}

	return nil
}

func cmpOr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func (s *InvoiceService) ID() string {
	return "Invoice"
}

func (s *InvoiceService) Log() {
	logs.Log().Info("[InvoiceService] Loaded")
}

var _ IService = (*InvoiceService)(nil)

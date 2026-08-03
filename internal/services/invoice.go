package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	BusinessName         string
	Address              string
	TIN                  string
	VATRegistration      string
	Email                string
	ContactNumber        string
	Website              string
	FooterNotes          string
	LogoURL              string
	LogoPath             string
	Currency             string
	VATPercentage        string
	ProprietorName       string
	BIRBookletsInfo      string
	BIRAuthorityToPrint  string
	BIRDateIssued        string
}

type InvoiceConfigInput struct {
	BusinessName        string
	Address             string
	TIN                 string
	VATRegistration     string
	Email               string
	ContactNumber       string
	Website             string
	FooterNotes         string
	Currency            string
	VATPercentage       string
	LogoURL             string
	LogoPath            string
	ProprietorName      string
	BIRBookletsInfo     string
	BIRAuthorityToPrint string
	BIRDateIssued       string
}

type InvoiceRecipient struct {
	ID              string
	Name            string
	Email           string
	ContactNumber   string
	Address         string
	TIN             string
	RegisteredName  string
	Notes           string
}

type InvoiceRecipientInput struct {
	Name           string
	Email          string
	ContactNumber  string
	Address        string
	TIN            string
	RegisteredName string
	Notes          string
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

type InvoiceLineInput struct {
	ProductID   string
	Description string
	UnitPrice   string
	Quantity    int64
	TaxType     string
}

type CreateInvoiceInput struct {
	RecipientID             string
	NewRecipient            *InvoiceRecipientInput
	TransactionType         string
	RecipientRegisteredName string
	Notes                   string
	IssueDate               string
	DeliveryDate            string
	DueDate                 string
	PaymentTermsValue       int64
	PaymentTermsUnit        string
	WithholdingTax          string
	SCPWDDiscount           string
	AddVAT                  string
	Lines                   []InvoiceLineInput
}

type InvoiceLine struct {
	Description  string
	UnitPrice    string
	LineTotal    string
	TaxType      enums.InvoiceLineTaxType
	ProductID    sql.NullInt64
	UnitPriceRaw int64
	LineTotalRaw int64
	Quantity     int64
}

type Invoice struct {
	ID                      string
	InvoiceNumber           string
	Status                  enums.InvoiceStatus
	TransactionType         enums.InvoiceTransactionType
	RecipientName           string
	RecipientRegisteredName string
	RecipientEmail          string
	RecipientContactNumber  string
	RecipientAddress        string
	RecipientTIN            string
	IssueDate               string
	DeliveryDate            string
	DueDate                 string
	PaymentTermsValue       int64
	PaymentTermsUnit        string
	Notes                   string
	Currency                string
	VATPercentage           string
	Subtotal                string
	VATAmount               string
	Total                   string
	EmailedAt               string
	CreatedAt               string
	PDFPath                 string
	ReceivedAmount          string
	SCPWDIDNo               string
	SubtotalRaw             int64
	VATAmountRaw            int64
	TotalRaw                int64
	VATableSalesRaw         int64
	VATExemptSalesRaw       int64
	ZeroRatedSalesRaw       int64
	TotalSalesRaw           int64
	TotalSalesVATInclusiveRaw int64
	LessVATRaw              int64
	WithholdingTaxRaw       int64
	AmountNetOfVATRaw       int64
	SCPWDDiscountRaw        int64
	AddVATRaw               int64
}

type InvoiceListItem struct {
	ID             string
	InvoiceNumber  string
	RecipientEmail string
	Status         enums.InvoiceStatus
	IssueDate      string
	Subtotal       string
	Total          string
	Emailed        bool
	PDFReady       bool
	CreatedAt      string
}

type InvoiceJobStatusView struct {
	PDFStatus   string
	EmailStatus string
	PDFError    string
	EmailError  string
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
		BusinessName:        c.BusinessName,
		Address:             c.Address,
		TIN:                 c.Tin,
		VATRegistration:     c.VatRegistration,
		Email:               c.Email,
		ContactNumber:       c.ContactNumber,
		Website:             c.Website,
		FooterNotes:         c.FooterNotes,
		LogoURL:             c.LogoUrl,
		LogoPath:            c.LogoPath,
		Currency:            cmpOr(c.Currency, constants.PHP),
		VATPercentage:       cmpOr(c.VatPercentage, "0"),
		ProprietorName:      c.ProprietorName,
		BIRBookletsInfo:     c.BirBookletsInfo,
		BIRAuthorityToPrint: c.BirAuthorityToPrint,
		BIRDateIssued:       c.BirDateIssued,
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
		BusinessName:        in.BusinessName,
		Address:             in.Address,
		Tin:                 in.TIN,
		VatRegistration:     in.VATRegistration,
		Email:               in.Email,
		ContactNumber:       in.ContactNumber,
		Website:             in.Website,
		FooterNotes:         in.FooterNotes,
		LogoUrl:             in.LogoURL,
		LogoPath:            in.LogoPath,
		Currency:            currency,
		VatPercentage:       vat,
		ProprietorName:      in.ProprietorName,
		BirBookletsInfo:     in.BIRBookletsInfo,
		BirAuthorityToPrint: in.BIRAuthorityToPrint,
		BirDateIssued:       in.BIRDateIssued,
	}); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoice, err)
	}
	return nil
}

func (s *InvoiceService) validateConfigForCreate(config InvoiceConfig) error {
	if strings.TrimSpace(config.BusinessName) == "" {
		return errs.ErrInvoiceConfigRequired
	}
	return nil
}

func (s *InvoiceService) GetRecipients(ctx context.Context, search string) ([]InvoiceRecipient, error) {
	recipients, _, err := s.GetRecipientsPaginated(ctx, search, 1, 10000)
	return recipients, err
}

func (s *InvoiceService) GetRecipientsPaginated(ctx context.Context, search string, page, perPage int) ([]InvoiceRecipient, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = constants.DefaultAdminTablePageSize
	}
	searchArg := strings.TrimSpace(search)

	total, err := s.dbRO.GetQueries().CountInvoiceRecipients(ctx, searchArg)
	if err != nil {
		return nil, 0, errors.Join(errs.ErrInvoice, err)
	}

	offset := int64((page - 1) * perPage)
	rows, err := s.dbRO.GetQueries().ListInvoiceRecipientsPaginated(ctx, queries.ListInvoiceRecipientsPaginatedParams{
		Search: searchArg,
		Limit:  int64(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, errors.Join(errs.ErrInvoice, err)
	}

	result := make([]InvoiceRecipient, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.toRecipient(row.TblInvoiceRecipient))
	}
	return result, total, nil
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
		ID:             s.encoder.Encode(r.ID),
		Name:           r.Name,
		Email:          r.Email,
		ContactNumber:  r.ContactNumber,
		Address:        r.Address,
		TIN:            r.Tin,
		RegisteredName: r.RegisteredName,
		Notes:          r.Notes,
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

	id, err := s.createRecipient(ctx, s.dbRW.GetQueries(), in)
	if err != nil {
		result = err.Error()
		return "", errors.Join(errs.ErrInvoice, err)
	}
	idStr := s.encoder.Encode(id)
	result = fmt.Sprintf("success. ID '%s'", idStr)
	return idStr, nil
}

func (s *InvoiceService) createRecipient(ctx context.Context, q *queries.Queries, in InvoiceRecipientInput) (int64, error) {
	regName := strings.TrimSpace(in.RegisteredName)
	if regName == "" {
		regName = strings.TrimSpace(in.Name)
	}
	return q.CreateInvoiceRecipient(ctx, queries.CreateInvoiceRecipientParams{
		Name:           strings.TrimSpace(in.Name),
		Email:          strings.TrimSpace(in.Email),
		ContactNumber:  strings.TrimSpace(in.ContactNumber),
		Address:        strings.TrimSpace(in.Address),
		Tin:            strings.TrimSpace(in.TIN),
		RegisteredName: regName,
		Notes:          strings.TrimSpace(in.Notes),
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
	regName := strings.TrimSpace(in.RegisteredName)
	if regName == "" {
		regName = strings.TrimSpace(in.Name)
	}

	if err := s.dbRW.GetQueries().UpdateInvoiceRecipient(ctx, queries.UpdateInvoiceRecipientParams{
		ID:             decoded,
		Name:           strings.TrimSpace(in.Name),
		Email:          strings.TrimSpace(in.Email),
		ContactNumber:  strings.TrimSpace(in.ContactNumber),
		Address:        strings.TrimSpace(in.Address),
		Tin:            strings.TrimSpace(in.TIN),
		RegisteredName: regName,
		Notes:          strings.TrimSpace(in.Notes),
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

func (s *InvoiceService) SearchProductsForLineItems(ctx context.Context, search string, limit int) ([]InvoiceProductOption, error) {
	limit = max(1, min(limit, 50))
	rows, err := s.dbRO.GetQueries().SearchProductsForInvoiceLineItems(ctx, queries.SearchProductsForInvoiceLineItemsParams{
		Search: strings.TrimSpace(search),
		Limit:  int64(limit),
	})
	if err != nil {
		return nil, errors.Join(errs.ErrInvoice, err)
	}
	result := make([]InvoiceProductOption, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.mapInvoiceProductOption(row))
	}
	return result, nil
}

func ProductOptionLabel(p InvoiceProductOption) string {
	label := p.Name
	if p.Serial != "" {
		label = p.Serial + " · " + label
	}
	if p.BrandName != "" {
		label = p.BrandName + " - " + label
	}
	if p.Price != "" {
		label += " · " + p.Price
	}
	return label
}

func InvoiceSearchLabel(inv InvoiceListItem) string {
	label := inv.InvoiceNumber
	if inv.RecipientEmail != "" {
		label += " · " + inv.RecipientEmail
	}
	if inv.Total != "" {
		label += " · " + inv.Total
	}
	return label
}

func (s *InvoiceService) mapInvoiceProductOption(row queries.SearchProductsForInvoiceLineItemsRow) InvoiceProductOption {
	price := row.UnitPriceWithVat
	currency := row.UnitPriceWithVatCurrency
	if row.IsOnSale == 1 && row.SalePriceWithVat.Valid {
		price = row.SalePriceWithVat.Int64
		if row.SalePriceWithVatCurrency.Valid {
			currency = row.SalePriceWithVatCurrency.String
		}
	}
	return InvoiceProductOption{
		ID:        s.encoder.Encode(row.ID),
		Name:      row.Name,
		Serial:    row.Serial,
		BrandName: row.BrandName,
		UnitPrice: price,
		Currency:  currency,
		Price:     utils.NewMoney(price, currency).Display(),
	}
}

func (s *InvoiceService) productUnitPrice(ctx context.Context, productID int64) (int64, string, string, error) {
	product, err := s.dbRO.GetQueries().GetProductsByID(ctx, productID)
	if err != nil {
		return 0, "", "", err
	}
	return product.UnitPriceWithVat, product.UnitPriceWithVatCurrency, product.Name, nil
}

func parseMoneyCentavos(raw, currency string) (int64, error) {
	cleaned := stripMoneyDecorations(raw)
	if cleaned == "" {
		return 0, nil
	}
	m, err := utils.NewMoneyFromString(cleaned, currency)
	if err != nil {
		return 0, err
	}
	return m.Amount(), nil
}

func stripMoneyDecorations(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, ",", "")
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseVATRate(v string) float64 {
	vatPct, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil || vatPct < 0 {
		return 0
	}
	return vatPct
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
	if err := s.validateConfigForCreate(config); err != nil {
		result = err.Error()
		return Invoice{}, err
	}
	currency := cmpOr(config.Currency, constants.PHP)

	lines, err := s.resolveLines(ctx, in.Lines, currency)
	if err != nil {
		result = err.Error()
		return Invoice{}, err
	}

	vatPct := parseVATRate(config.VATPercentage)
	taxLines := make([]BIRTaxLineInput, 0, len(lines))
	for _, ln := range lines {
		taxLines = append(taxLines, BIRTaxLineInput{LineTotal: ln.LineTotalRaw, TaxType: ln.TaxType})
	}
	withholding, _ := parseMoneyCentavos(in.WithholdingTax, currency)
	scPwd, _ := parseMoneyCentavos(in.SCPWDDiscount, currency)
	addVAT, _ := parseMoneyCentavos(in.AddVAT, currency)
	summary := ComputeBIRTaxSummary(taxLines, vatPct, BIRTaxAdjustments{
		WithholdingTax: withholding,
		SCPWDDiscount:  scPwd,
		AddVAT:         addVAT,
	})

	staffDBID := s.encoder.Decode(staffID)
	if staffDBID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return Invoice{}, errs.ErrDecode
	}

	issueDate := strings.TrimSpace(in.IssueDate)
	if issueDate == "" {
		issueDate = utils.NowPH().Format(constants.DateLayoutISO)
	}
	txType := defaultTransactionType(in.TransactionType)

	tx, err := s.dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.dbRW.GetQueries().WithTx(tx)

	recipient, err := s.resolveRecipient(ctx, qtx, in)
	if err != nil {
		result = err.Error()
		return Invoice{}, err
	}

	regName := strings.TrimSpace(in.RecipientRegisteredName)
	if regName == "" {
		regName = recipient.Name
	}

	recipientID := sql.NullInt64{}
	if recipient.dbID != 0 {
		recipientID = sql.NullInt64{Int64: recipient.dbID, Valid: true}
	}

	invoiceDBID, err := qtx.CreateInvoice(ctx, queries.CreateInvoiceParams{
		InvoiceNumber:           "",
		RecipientID:             recipientID,
		RecipientName:           recipient.Name,
		RecipientEmail:          recipient.Email,
		RecipientContactNumber:  recipient.ContactNumber,
		RecipientAddress:        recipient.Address,
		RecipientTin:            recipient.TIN,
		RecipientRegisteredName: regName,
		TransactionType:         txType.String(),
		Status:                  enums.INVOICE_STATUS_PROCESSING.String(),
		IssueDate:               issueDate,
		DueDate:                 strings.TrimSpace(in.DueDate),
		DeliveryDate:            strings.TrimSpace(in.DeliveryDate),
		PaymentTermsValue:       in.PaymentTermsValue,
		PaymentTermsUnit:        strings.TrimSpace(in.PaymentTermsUnit),
		Notes:                   strings.TrimSpace(in.Notes),
		Currency:                currency,
		Subtotal:                summary.Subtotal,
		VatPercentage:           cmpOr(config.VATPercentage, "0"),
		VatAmount:               summary.VATAmount,
		Total:                   summary.Total,
		VatableSales:            summary.VATableSales,
		VatExemptSales:          summary.VATExemptSales,
		ZeroRatedSales:          summary.ZeroRatedSales,
		TotalSales:              summary.TotalSales,
		TotalSalesVatInclusive:  summary.TotalSalesVATInclusive,
		LessVat:                 summary.LessVAT,
		WithholdingTax:          summary.WithholdingTax,
		AmountNetOfVat:          summary.AmountNetOfVAT,
		ScPwdDiscount:           summary.SCPWDDiscount,
		AddVat:                  summary.AddVAT,
		ReceivedAmount:          "",
		ScPwdIDNo:               "",
		PdfPath:                 "",
		CreatedBy:               staffDBID,
	})
	if err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
	}

	invoiceNumber := fmt.Sprintf("INV-%s-%05d", strings.ReplaceAll(issueDate, "-", ""), invoiceDBID)
	if err := qtx.SetInvoiceNumber(ctx, queries.SetInvoiceNumberParams{
		InvoiceNumber: invoiceNumber,
		ID:            invoiceDBID,
	}); err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
	}

	for _, ln := range lines {
		if err := qtx.CreateInvoiceLine(ctx, queries.CreateInvoiceLineParams{
			InvoiceID:   invoiceDBID,
			ProductID:   ln.ProductID,
			Description: ln.Description,
			Quantity:    ln.Quantity,
			UnitPrice:   ln.UnitPriceRaw,
			LineTotal:   ln.LineTotalRaw,
			TaxType:     ln.TaxType.String(),
			Currency:    currency,
		}); err != nil {
			result = err.Error()
			return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
		}
	}

	if err := tx.Commit(); err != nil {
		result = err.Error()
		return Invoice{}, errors.Join(errs.ErrInvoiceCreateFailed, err)
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

func (s *InvoiceService) resolveRecipient(ctx context.Context, q *queries.Queries, in CreateInvoiceInput) (resolvedRecipient, error) {
	if in.NewRecipient != nil && strings.TrimSpace(in.NewRecipient.Name) != "" {
		id, err := s.createRecipient(ctx, q, *in.NewRecipient)
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
	TaxType      enums.InvoiceLineTaxType
}

func (s *InvoiceService) resolveLines(ctx context.Context, inputs []InvoiceLineInput, currency string) ([]resolvedLine, error) {
	lines := make([]resolvedLine, 0, len(inputs))
	for _, li := range inputs {
		qty := li.Quantity
		if qty <= 0 {
			qty = 1
		}

		description := strings.TrimSpace(li.Description)
		var unitPrice int64
		productID := sql.NullInt64{}
		taxType := defaultTaxType(li.TaxType)

		if strings.TrimSpace(li.ProductID) != "" {
			decoded := s.encoder.Decode(li.ProductID)
			if decoded != encode.INVALID {
				productID = sql.NullInt64{Int64: decoded, Valid: true}
				price, _, name, err := s.productUnitPrice(ctx, decoded)
				if err == nil {
					if description == "" {
						description = name
					}
					unitPrice = price
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
		lines = append(lines, resolvedLine{
			Description:  description,
			ProductID:    productID,
			Quantity:     qty,
			UnitPriceRaw: unitPrice,
			LineTotalRaw: lineTotal,
			TaxType:      taxType,
		})
	}

	if len(lines) == 0 {
		return nil, errs.ErrInvoiceNoLines
	}
	return lines, nil
}

func (s *InvoiceService) mapInvoice(inv queries.TblInvoice, currency string) Invoice {
	return Invoice{
		ID:                        s.encoder.Encode(inv.ID),
		InvoiceNumber:             inv.InvoiceNumber,
		Status:                    enums.ParseInvoiceStatusToEnum(inv.Status),
		TransactionType:           defaultTransactionType(inv.TransactionType),
		RecipientName:             inv.RecipientName,
		RecipientRegisteredName:   inv.RecipientRegisteredName,
		RecipientEmail:            inv.RecipientEmail,
		RecipientContactNumber:    inv.RecipientContactNumber,
		RecipientAddress:          inv.RecipientAddress,
		RecipientTIN:              inv.RecipientTin,
		IssueDate:                 inv.IssueDate,
		DeliveryDate:              inv.DeliveryDate,
		DueDate:                   inv.DueDate,
		PaymentTermsValue:         inv.PaymentTermsValue,
		PaymentTermsUnit:          inv.PaymentTermsUnit,
		Notes:                     inv.Notes,
		Currency:                  currency,
		VATPercentage:             inv.VatPercentage,
		Subtotal:                  utils.NewMoney(inv.Subtotal, currency).Display(),
		VATAmount:                 utils.NewMoney(inv.VatAmount, currency).Display(),
		Total:                     utils.NewMoney(inv.Total, currency).Display(),
		EmailedAt:                 utils.ConvertToPH(inv.EmailedAt),
		CreatedAt:                 utils.ConvertToPH(inv.CreatedAt),
		PDFPath:                   inv.PdfPath,
		ReceivedAmount:            inv.ReceivedAmount,
		SCPWDIDNo:                 inv.ScPwdIDNo,
		SubtotalRaw:               inv.Subtotal,
		VATAmountRaw:              inv.VatAmount,
		TotalRaw:                  inv.Total,
		VATableSalesRaw:           inv.VatableSales,
		VATExemptSalesRaw:         inv.VatExemptSales,
		ZeroRatedSalesRaw:         inv.ZeroRatedSales,
		TotalSalesRaw:             inv.TotalSales,
		TotalSalesVATInclusiveRaw: inv.TotalSalesVatInclusive,
		LessVATRaw:                inv.LessVat,
		WithholdingTaxRaw:         inv.WithholdingTax,
		AmountNetOfVATRaw:         inv.AmountNetOfVat,
		SCPWDDiscountRaw:          inv.ScPwdDiscount,
		AddVATRaw:                 inv.AddVat,
	}
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
			TaxType:      defaultTaxType(l.TaxType),
			UnitPriceRaw: l.UnitPrice,
			LineTotalRaw: l.LineTotal,
			UnitPrice:    utils.NewMoney(l.UnitPrice, currency).Display(),
			LineTotal:    utils.NewMoney(l.LineTotal, currency).Display(),
		})
	}

	return s.mapInvoice(inv, currency), lines, nil
}

func (s *InvoiceService) GetInvoiceDBID(ctx context.Context, id string) (int64, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return 0, errs.ErrDecode
	}
	return decoded, nil
}

func (s *InvoiceService) mapInvoiceListItem(r queries.ListInvoicesPaginatedRow) InvoiceListItem {
	currency := cmpOr(r.Currency, constants.PHP)
	return InvoiceListItem{
		ID:             s.encoder.Encode(r.ID),
		InvoiceNumber:  r.InvoiceNumber,
		RecipientEmail: r.RecipientEmail,
		Status:         enums.ParseInvoiceStatusToEnum(r.Status),
		IssueDate:      r.IssueDate,
		Subtotal:       utils.NewMoney(r.Subtotal, currency).Display(),
		Total:          utils.NewMoney(r.Total, currency).Display(),
		Emailed:        r.EmailedAt != "",
		PDFReady:       strings.TrimSpace(r.PdfPath) != "",
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
}

func (s *InvoiceService) mapSearchInvoiceRow(r queries.SearchInvoicesRow) InvoiceListItem {
	currency := cmpOr(r.Currency, constants.PHP)
	return InvoiceListItem{
		ID:             s.encoder.Encode(r.ID),
		InvoiceNumber:  r.InvoiceNumber,
		RecipientEmail: r.RecipientEmail,
		Status:         enums.ParseInvoiceStatusToEnum(r.Status),
		IssueDate:      r.IssueDate,
		Subtotal:       utils.NewMoney(r.Subtotal, currency).Display(),
		Total:          utils.NewMoney(r.Total, currency).Display(),
		Emailed:        r.EmailedAt != "",
		PDFReady:       strings.TrimSpace(r.PdfPath) != "",
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
}

func (s *InvoiceService) SearchInvoices(ctx context.Context, search string, limit int) ([]InvoiceListItem, error) {
	limit = max(1, min(limit, 50))
	rows, err := s.dbRO.GetQueries().SearchInvoices(ctx, queries.SearchInvoicesParams{
		Search: strings.TrimSpace(search),
		Limit:  int64(limit),
	})
	if err != nil {
		return nil, errors.Join(errs.ErrInvoice, err)
	}
	result := make([]InvoiceListItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.mapSearchInvoiceRow(row))
	}
	return result, nil
}

func (s *InvoiceService) GetInvoicesPaginated(ctx context.Context, page, perPage int) ([]InvoiceListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = constants.DefaultAdminTablePageSize
	}

	total, err := s.dbRO.GetQueries().CountInvoices(ctx)
	if err != nil {
		return nil, 0, errors.Join(errs.ErrInvoice, err)
	}

	offset := int64((page - 1) * perPage)
	rows, err := s.dbRO.GetQueries().ListInvoicesPaginated(ctx, queries.ListInvoicesPaginatedParams{
		Limit:  int64(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, errors.Join(errs.ErrInvoice, err)
	}

	result := make([]InvoiceListItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, s.mapInvoiceListItem(r))
	}
	return result, total, nil
}

func (s *InvoiceService) GetJobStatus(ctx context.Context, invoiceID int64) (InvoiceJobStatusView, error) {
	view := InvoiceJobStatusView{PDFStatus: "none", EmailStatus: "none"}
	if row, err := s.dbRO.GetQueries().GetLatestInvoiceJobByInvoiceIDAndType(ctx, queries.GetLatestInvoiceJobByInvoiceIDAndTypeParams{
		InvoiceID: invoiceID,
		JobType:   enums.INVOICE_JOB_GENERATE_PDF.String(),
	}); err == nil {
		view.PDFStatus = strings.ToLower(row.TblInvoiceJob.Status)
		view.PDFError = row.TblInvoiceJob.ErrorMessage
	}
	if row, err := s.dbRO.GetQueries().GetLatestInvoiceJobByInvoiceIDAndType(ctx, queries.GetLatestInvoiceJobByInvoiceIDAndTypeParams{
		InvoiceID: invoiceID,
		JobType:   enums.INVOICE_JOB_SEND_EMAIL.String(),
	}); err == nil {
		view.EmailStatus = strings.ToLower(row.TblInvoiceJob.Status)
		view.EmailError = row.TblInvoiceJob.ErrorMessage
	}
	return view, nil
}

func invoicePDFDir() string {
	return filepath.Join("cmd", "web", "static", "invoices")
}

func (s *InvoiceService) GenerateAndStorePDF(ctx context.Context, invoiceID int64) error {
	idStr := s.encoder.Encode(invoiceID)
	invoice, lines, err := s.GetInvoice(ctx, idStr)
	if err != nil {
		return err
	}
	config, err := s.GetConfig(ctx)
	if err != nil {
		return err
	}

	pdfBytes, err := RenderInvoicePDF(config, invoice, lines)
	if err != nil {
		return errors.Join(errs.ErrInvoicePDFFailed, err)
	}

	if err := os.MkdirAll(invoicePDFDir(), 0o755); err != nil {
		return err
	}
	localPath := filepath.Join(invoicePDFDir(), fmt.Sprintf("%d.pdf", invoiceID))
	if err := os.WriteFile(localPath, pdfBytes, 0o644); err != nil {
		return err
	}

	if err := s.dbRW.GetQueries().SetInvoicePDFPath(ctx, queries.SetInvoicePDFPathParams{
		PdfPath: localPath,
		ID:      invoiceID,
	}); err != nil {
		return errors.Join(errs.ErrInvoice, err)
	}
	return nil
}

func (s *InvoiceService) ReadStoredPDF(invoiceID int64, pdfPath string) ([]byte, error) {
	path := strings.TrimSpace(pdfPath)
	if path == "" {
		path = filepath.Join(invoicePDFDir(), fmt.Sprintf("%d.pdf", invoiceID))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Join(errs.ErrInvoicePDFFailed, err)
	}
	return data, nil
}

func (s *InvoiceService) SendInvoiceEmailByID(ctx context.Context, staffID string, invoiceID int64) error {
	return s.SendInvoiceEmail(ctx, staffID, s.encoder.Encode(invoiceID))
}

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

	decoded := s.encoder.Decode(id)
	if strings.TrimSpace(invoice.PDFPath) == "" {
		if err := s.GenerateAndStorePDF(ctx, decoded); err != nil {
			result = err.Error()
			return err
		}
		invoice, lines, err = s.GetInvoice(ctx, id)
		if err != nil {
			result = err.Error()
			return err
		}
	}

	config, err := s.GetConfig(ctx)
	if err != nil {
		result = err.Error()
		return err
	}

	pdfBytes, err := s.ReadStoredPDF(decoded, invoice.PDFPath)
	if err != nil {
		result = err.Error()
		return errors.Join(errs.ErrInvoicePDFFailed, err)
	}

	businessName := cmpOr(config.BusinessName, "C-Choice")
	subject := fmt.Sprintf("Invoice %s from %s", invoice.InvoiceNumber, businessName)
	data := s.BuildEmailTemplateData(config, invoice, lines)

	attachments := []mail.Attachment{
		{
			FileName:    cmpOr(invoice.InvoiceNumber, "invoice") + ".pdf",
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

	if err := s.dbRW.GetQueries().MarkInvoiceEmailed(ctx, decoded); err != nil {
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

// Compile-time check for invoice job runner interface.
var _ interface {
	GenerateAndStorePDF(ctx context.Context, invoiceID int64) error
	SendInvoiceEmailByID(ctx context.Context, staffID string, invoiceID int64) error
} = (*InvoiceService)(nil)

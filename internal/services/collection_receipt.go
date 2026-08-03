package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

type CollectionReceiptSettlementInput struct {
	InvoiceID     string
	InvoiceNumber string
	Amount        string
}

type CreateCollectionReceiptInput struct {
	InvoiceID     string
	ReceivedFrom  string
	RecipientEmail string
	TIN           string
	Address       string
	ReceiptDate   string
	Amount        string
	PaymentFor    string
	PaymentForm   string
	SCCitizenTIN  string
	OSCAPWDIDNo   string
	Settlements   []CollectionReceiptSettlementInput
}

type CollectionReceiptSettlement struct {
	InvoiceID     string
	InvoiceNumber string
	Amount        string
	AmountRaw     int64
}

type CollectionReceipt struct {
	ID             string
	ReceiptNumber  string
	InvoiceID      string
	ReceivedFrom   string
	RecipientEmail string
	TIN            string
	Address        string
	ReceiptDate    string
	Amount         string
	AmountRaw      int64
	AmountInWords  string
	PaymentFor     string
	PaymentForm    enums.ReceiptPaymentForm
	SCCitizenTIN   string
	OSCAPWDIDNo    string
	Status         enums.ReceiptStatus
	PDFPath        string
	EmailedAt      string
	CreatedAt      string
}

type CollectionReceiptListItem struct {
	ID             string
	ReceiptNumber  string
	ReceivedFrom   string
	RecipientEmail string
	Status         enums.ReceiptStatus
	ReceiptDate    string
	Amount         string
	Emailed        bool
	PDFReady       bool
	CreatedAt      string
}

type CollectionReceiptJobStatusView struct {
	PDFStatus   string
	EmailStatus string
	PDFError    string
	EmailError  string
}

type CollectionReceiptService struct {
	encoder        encode.IEncode
	dbRO           database.IService
	dbRW           database.IService
	staffLog       *StaffLogsService
	mailService    mail.IMailService
	invoiceService *InvoiceService
}

func NewCollectionReceiptService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
	mailService mail.IMailService,
	invoiceService *InvoiceService,
) *CollectionReceiptService {
	return &CollectionReceiptService{
		encoder:        encoder,
		dbRO:           dbRO,
		dbRW:           dbRW,
		staffLog:       staffLog,
		mailService:    mailService,
		invoiceService: invoiceService,
	}
}

func (s *CollectionReceiptService) CreateCollectionReceipt(ctx context.Context, staffID string, in CreateCollectionReceiptInput) (CollectionReceipt, []CollectionReceiptSettlement, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionCreate, constants.ModuleCollectionReceipts, result, nil); err != nil {
			logs.Log().Warn("[CollectionReceiptService] create log", zap.Error(err))
		}
	}()

	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, err
	}
	if err := s.invoiceService.validateConfigForCreate(config); err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, err
	}
	currency := cmpOr(config.Currency, constants.PHP)

	in, err = s.prefillFromInvoice(ctx, in)
	if err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, err
	}

	if strings.TrimSpace(in.ReceivedFrom) == "" {
		result = errs.ErrInvoiceRecipientNameReq.Error()
		return CollectionReceipt{}, nil, errs.ErrInvoiceRecipientNameReq
	}

	settlements, amountRaw, err := s.resolveSettlements(in.Settlements, in.Amount, currency)
	if err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, err
	}
	if amountRaw <= 0 {
		result = errs.ErrCollectionReceiptCreateFailed.Error()
		return CollectionReceipt{}, nil, errs.ErrCollectionReceiptCreateFailed
	}

	staffDBID := s.encoder.Decode(staffID)
	if staffDBID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return CollectionReceipt{}, nil, errs.ErrDecode
	}

	receiptDate := strings.TrimSpace(in.ReceiptDate)
	if receiptDate == "" {
		receiptDate = utils.NowPH().Format(constants.DateLayoutISO)
	}

	paymentForm := defaultReceiptPaymentForm(in.PaymentForm)
	amountInWords := utils.AmountInWordsPHP(amountRaw)

	invoiceID := sql.NullInt64{}
	if strings.TrimSpace(in.InvoiceID) != "" {
		decoded := s.encoder.Decode(in.InvoiceID)
		if decoded == encode.INVALID {
			result = errs.ErrDecode.Error()
			return CollectionReceipt{}, nil, errs.ErrDecode
		}
		invoiceID = sql.NullInt64{Int64: decoded, Valid: true}
	}

	tx, err := s.dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceiptCreateFailed, err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.dbRW.GetQueries().WithTx(tx)

	receiptDBID, err := qtx.CreateCollectionReceipt(ctx, queries.CreateCollectionReceiptParams{
		ReceiptNumber:  "",
		InvoiceID:      invoiceID,
		ReceivedFrom:   strings.TrimSpace(in.ReceivedFrom),
		RecipientEmail: strings.TrimSpace(in.RecipientEmail),
		Tin:            strings.TrimSpace(in.TIN),
		Address:        strings.TrimSpace(in.Address),
		ReceiptDate:    receiptDate,
		Amount:         amountRaw,
		AmountInWords:  amountInWords,
		PaymentFor:     strings.TrimSpace(in.PaymentFor),
		PaymentForm:    paymentForm.String(),
		ScCitizenTin:   strings.TrimSpace(in.SCCitizenTIN),
		OscaPwdIDNo:    strings.TrimSpace(in.OSCAPWDIDNo),
		Status:         enums.RECEIPT_STATUS_PROCESSING.String(),
		CreatedBy:      staffDBID,
	})
	if err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceiptCreateFailed, err)
	}

	receiptNumber := fmt.Sprintf("CR-%s-%05d", strings.ReplaceAll(receiptDate, "-", ""), receiptDBID)
	if err := qtx.SetCollectionReceiptNumber(ctx, queries.SetCollectionReceiptNumberParams{
		ReceiptNumber: receiptNumber,
		ID:            receiptDBID,
	}); err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceiptCreateFailed, err)
	}

	for i, st := range settlements {
		invID := sql.NullInt64{}
		if strings.TrimSpace(st.InvoiceID) != "" {
			decoded := s.encoder.Decode(st.InvoiceID)
			if decoded != encode.INVALID {
				invID = sql.NullInt64{Int64: decoded, Valid: true}
			}
		}
		if err := qtx.CreateCollectionReceiptSettlement(ctx, queries.CreateCollectionReceiptSettlementParams{
			CollectionReceiptID: receiptDBID,
			InvoiceID:           invID,
			InvoiceNumber:       strings.TrimSpace(st.InvoiceNumber),
			Amount:              st.AmountRaw,
			SortOrder:           int64(i),
		}); err != nil {
			result = err.Error()
			return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceiptCreateFailed, err)
		}
	}

	if err := tx.Commit(); err != nil {
		result = err.Error()
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceiptCreateFailed, err)
	}

	idStr := s.encoder.Encode(receiptDBID)
	result = fmt.Sprintf("success. ID '%s' number '%s'", idStr, receiptNumber)
	return s.GetCollectionReceipt(ctx, idStr)
}

func (s *CollectionReceiptService) prefillFromInvoice(ctx context.Context, in CreateCollectionReceiptInput) (CreateCollectionReceiptInput, error) {
	if strings.TrimSpace(in.InvoiceID) == "" {
		return in, nil
	}

	invoice, _, err := s.invoiceService.GetInvoice(ctx, in.InvoiceID)
	if err != nil {
		return in, err
	}

	if strings.TrimSpace(in.ReceivedFrom) == "" {
		in.ReceivedFrom = invoice.RecipientRegisteredName
		if strings.TrimSpace(in.ReceivedFrom) == "" {
			in.ReceivedFrom = invoice.RecipientName
		}
	}
	if strings.TrimSpace(in.RecipientEmail) == "" {
		in.RecipientEmail = invoice.RecipientEmail
	}
	if strings.TrimSpace(in.TIN) == "" {
		in.TIN = invoice.RecipientTIN
	}
	if strings.TrimSpace(in.Address) == "" {
		in.Address = invoice.RecipientAddress
	}
	currency := cmpOr(invoice.Currency, constants.PHP)
	if len(in.Settlements) == 0 {
		amountDisplay, _ := utils.SchemaPrice(invoice.TotalSalesVATInclusiveRaw, currency)
		in.Settlements = []CollectionReceiptSettlementInput{{
			InvoiceID:     in.InvoiceID,
			InvoiceNumber: invoice.InvoiceNumber,
			Amount:        amountDisplay,
		}}
	}
	if strings.TrimSpace(in.Amount) == "" {
		in.Amount, _ = utils.SchemaPrice(invoice.TotalSalesVATInclusiveRaw, currency)
	}
	return in, nil
}

type resolvedSettlement struct {
	CollectionReceiptSettlement
}

func (s *CollectionReceiptService) resolveSettlements(inputs []CollectionReceiptSettlementInput, amountRaw string, currency string) ([]resolvedSettlement, int64, error) {
	if len(inputs) == 0 {
		total, err := parseMoneyCentavos(amountRaw, currency)
		if err != nil {
			return nil, 0, err
		}
		return nil, total, nil
	}

	settlements := make([]resolvedSettlement, 0, len(inputs))
	var total int64
	for _, row := range inputs {
		amt, err := parseMoneyCentavos(row.Amount, currency)
		if err != nil {
			return nil, 0, err
		}
		total += amt
		settlements = append(settlements, resolvedSettlement{
			CollectionReceiptSettlement: CollectionReceiptSettlement{
				InvoiceID:     strings.TrimSpace(row.InvoiceID),
				InvoiceNumber: strings.TrimSpace(row.InvoiceNumber),
				Amount:        utils.NewMoney(amt, currency).Display(),
				AmountRaw:     amt,
			},
		})
	}

	if strings.TrimSpace(amountRaw) != "" {
		explicit, err := parseMoneyCentavos(amountRaw, currency)
		if err != nil {
			return nil, 0, err
		}
		if explicit > 0 && explicit != total {
			return nil, 0, errs.ErrCollectionReceiptSettlementMismatch
		}
	}

	return settlements, total, nil
}

func (s *CollectionReceiptService) mapCollectionReceipt(r queries.TblCollectionReceipt, currency string) CollectionReceipt {
	var invoiceID string
	if r.InvoiceID.Valid {
		invoiceID = s.encoder.Encode(r.InvoiceID.Int64)
	}
	return CollectionReceipt{
		ID:             s.encoder.Encode(r.ID),
		ReceiptNumber:  r.ReceiptNumber,
		InvoiceID:      invoiceID,
		ReceivedFrom:   r.ReceivedFrom,
		RecipientEmail: r.RecipientEmail,
		TIN:            r.Tin,
		Address:        r.Address,
		ReceiptDate:    r.ReceiptDate,
		Amount:         utils.NewMoney(r.Amount, currency).Display(),
		AmountRaw:      r.Amount,
		AmountInWords:  r.AmountInWords,
		PaymentFor:     r.PaymentFor,
		PaymentForm:    defaultReceiptPaymentForm(r.PaymentForm),
		SCCitizenTIN:   r.ScCitizenTin,
		OSCAPWDIDNo:    r.OscaPwdIDNo,
		Status:         enums.ParseReceiptStatusToEnum(r.Status),
		PDFPath:        r.PdfPath,
		EmailedAt:      utils.ConvertToPH(r.EmailedAt),
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
}

func (s *CollectionReceiptService) mapCollectionSettlement(row queries.TblCollectionReceiptSettlement, currency string) CollectionReceiptSettlement {
	var invoiceID string
	if row.InvoiceID.Valid {
		invoiceID = s.encoder.Encode(row.InvoiceID.Int64)
	}
	return CollectionReceiptSettlement{
		InvoiceID:     invoiceID,
		InvoiceNumber: row.InvoiceNumber,
		Amount:        utils.NewMoney(row.Amount, currency).Display(),
		AmountRaw:     row.Amount,
	}
}

func (s *CollectionReceiptService) GetCollectionReceipt(ctx context.Context, id string) (CollectionReceipt, []CollectionReceiptSettlement, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return CollectionReceipt{}, nil, errs.ErrDecode
	}

	row, err := s.dbRO.GetQueries().GetCollectionReceiptByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CollectionReceipt{}, nil, errs.ErrCollectionReceiptNotFound
		}
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceipt, err)
	}

	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		return CollectionReceipt{}, nil, err
	}
	currency := cmpOr(config.Currency, constants.PHP)

	settlementRows, err := s.dbRO.GetQueries().GetCollectionReceiptSettlementsByReceiptID(ctx, decoded)
	if err != nil {
		return CollectionReceipt{}, nil, errors.Join(errs.ErrCollectionReceipt, err)
	}

	settlements := make([]CollectionReceiptSettlement, 0, len(settlementRows))
	for _, sr := range settlementRows {
		settlements = append(settlements, s.mapCollectionSettlement(sr.TblCollectionReceiptSettlement, currency))
	}

	return s.mapCollectionReceipt(row.TblCollectionReceipt, currency), settlements, nil
}

func (s *CollectionReceiptService) GetCollectionReceiptDBID(ctx context.Context, id string) (int64, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return 0, errs.ErrDecode
	}
	return decoded, nil
}

func (s *CollectionReceiptService) mapCollectionReceiptListItem(r queries.ListCollectionReceiptsPaginatedRow, currency string) CollectionReceiptListItem {
	return CollectionReceiptListItem{
		ID:             s.encoder.Encode(r.ID),
		ReceiptNumber:  r.ReceiptNumber,
		ReceivedFrom:   r.ReceivedFrom,
		RecipientEmail: r.RecipientEmail,
		Status:         enums.ParseReceiptStatusToEnum(r.Status),
		ReceiptDate:    r.ReceiptDate,
		Amount:         utils.NewMoney(r.Amount, currency).Display(),
		Emailed:        r.EmailedAt != "",
		PDFReady:       strings.TrimSpace(r.PdfPath) != "",
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
}

func (s *CollectionReceiptService) GetCollectionReceiptsPaginated(ctx context.Context, page, perPage int) ([]CollectionReceiptListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = constants.DefaultAdminTablePageSize
	}

	total, err := s.dbRO.GetQueries().CountCollectionReceipts(ctx)
	if err != nil {
		return nil, 0, errors.Join(errs.ErrCollectionReceipt, err)
	}

	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		return nil, 0, err
	}
	currency := cmpOr(config.Currency, constants.PHP)

	offset := int64((page - 1) * perPage)
	rows, err := s.dbRO.GetQueries().ListCollectionReceiptsPaginated(ctx, queries.ListCollectionReceiptsPaginatedParams{
		Limit:  int64(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, errors.Join(errs.ErrCollectionReceipt, err)
	}

	result := make([]CollectionReceiptListItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, s.mapCollectionReceiptListItem(r, currency))
	}
	return result, total, nil
}

func (s *CollectionReceiptService) GetJobStatus(ctx context.Context, receiptID int64) (CollectionReceiptJobStatusView, error) {
	view := CollectionReceiptJobStatusView{PDFStatus: "none", EmailStatus: "none"}
	if row, err := s.dbRO.GetQueries().GetLatestCollectionReceiptJobByReceiptIDAndType(ctx, queries.GetLatestCollectionReceiptJobByReceiptIDAndTypeParams{
		CollectionReceiptID: receiptID,
		JobType:             enums.COLLECTION_RECEIPT_JOB_GENERATE_PDF.String(),
	}); err == nil {
		view.PDFStatus = strings.ToLower(row.TblCollectionReceiptJob.Status)
		view.PDFError = row.TblCollectionReceiptJob.ErrorMessage
	}
	if row, err := s.dbRO.GetQueries().GetLatestCollectionReceiptJobByReceiptIDAndType(ctx, queries.GetLatestCollectionReceiptJobByReceiptIDAndTypeParams{
		CollectionReceiptID: receiptID,
		JobType:             enums.COLLECTION_RECEIPT_JOB_SEND_EMAIL.String(),
	}); err == nil {
		view.EmailStatus = strings.ToLower(row.TblCollectionReceiptJob.Status)
		view.EmailError = row.TblCollectionReceiptJob.ErrorMessage
	}
	return view, nil
}

func collectionReceiptPDFDir() string {
	return filepath.Join("cmd", "web", "static", "collection_receipts")
}

func (s *CollectionReceiptService) GenerateAndStorePDF(ctx context.Context, receiptID int64) error {
	idStr := s.encoder.Encode(receiptID)
	receipt, settlements, err := s.GetCollectionReceipt(ctx, idStr)
	if err != nil {
		return err
	}
	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		return err
	}

	pdfBytes, err := RenderCollectionReceiptPDF(config, receipt, settlements)
	if err != nil {
		return errors.Join(errs.ErrCollectionReceiptPDFFailed, err)
	}

	if err := os.MkdirAll(collectionReceiptPDFDir(), 0o755); err != nil {
		return err
	}
	localPath := filepath.Join(collectionReceiptPDFDir(), fmt.Sprintf("%d.pdf", receiptID))
	if err := os.WriteFile(localPath, pdfBytes, 0o644); err != nil {
		return err
	}

	if err := s.dbRW.GetQueries().SetCollectionReceiptPDFPath(ctx, queries.SetCollectionReceiptPDFPathParams{
		PdfPath: localPath,
		ID:      receiptID,
	}); err != nil {
		return errors.Join(errs.ErrCollectionReceipt, err)
	}
	return nil
}

func (s *CollectionReceiptService) ReadStoredPDF(receiptID int64, pdfPath string) ([]byte, error) {
	path := strings.TrimSpace(pdfPath)
	if path == "" {
		path = filepath.Join(collectionReceiptPDFDir(), fmt.Sprintf("%d.pdf", receiptID))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Join(errs.ErrCollectionReceiptPDFFailed, err)
	}
	return data, nil
}

func (s *CollectionReceiptService) SendCollectionReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error {
	return s.SendCollectionReceiptEmail(ctx, staffID, s.encoder.Encode(receiptID))
}

func (s *CollectionReceiptService) SendCollectionReceiptEmail(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionSend, constants.ModuleCollectionReceipts, result, nil); err != nil {
			logs.Log().Warn("[CollectionReceiptService] send email log", zap.Error(err))
		}
	}()

	if s.mailService == nil {
		result = errs.ErrCollectionReceiptEmailNotConfigured.Error()
		return errs.ErrCollectionReceiptEmailNotConfigured
	}

	receipt, settlements, err := s.GetCollectionReceipt(ctx, id)
	if err != nil {
		result = err.Error()
		return err
	}
	if strings.TrimSpace(receipt.RecipientEmail) == "" {
		result = errs.ErrCollectionReceiptRecipientNoEmail.Error()
		return errs.ErrCollectionReceiptRecipientNoEmail
	}

	decoded := s.encoder.Decode(id)
	if strings.TrimSpace(receipt.PDFPath) == "" {
		if err := s.GenerateAndStorePDF(ctx, decoded); err != nil {
			result = err.Error()
			return err
		}
		receipt, settlements, err = s.GetCollectionReceipt(ctx, id)
		if err != nil {
			result = err.Error()
			return err
		}
	}

	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		result = err.Error()
		return err
	}

	pdfBytes, err := s.ReadStoredPDF(decoded, receipt.PDFPath)
	if err != nil {
		result = err.Error()
		return errors.Join(errs.ErrCollectionReceiptPDFFailed, err)
	}

	businessName := cmpOr(config.BusinessName, "C-Choice")
	subject := fmt.Sprintf("Collection Receipt %s from %s", receipt.ReceiptNumber, businessName)
	data := s.BuildEmailTemplateData(config, receipt, settlements)

	attachments := []mail.Attachment{
		{
			FileName:    cmpOr(receipt.ReceiptNumber, "collection_receipt") + ".pdf",
			ContentType: "application/pdf",
			Content:     pdfBytes,
		},
	}

	if err := s.mailService.SendTemplateEmailWithAttachments(
		receipt.RecipientEmail,
		nil,
		subject,
		enums.EMAIL_TEMPLATE_COLLECTION_RECEIPT.FileName(),
		data,
		attachments,
	); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrCollectionReceiptEmailFailed, err)
	}

	if err := s.dbRW.GetQueries().MarkCollectionReceiptEmailed(ctx, decoded); err != nil {
		logs.LogCtx(ctx).Warn("[CollectionReceiptService] mark emailed", zap.Error(err))
	}

	return nil
}

func defaultReceiptPaymentForm(s string) enums.ReceiptPaymentForm {
	if pf := enums.ParseReceiptPaymentFormToEnum(s); pf != enums.RECEIPT_PAYMENT_UNDEFINED {
		return pf
	}
	return enums.RECEIPT_PAYMENT_CASH
}

func (s *CollectionReceiptService) ID() string {
	return "CollectionReceipt"
}

func (s *CollectionReceiptService) Log() {
	logs.Log().Info("[CollectionReceiptService] Loaded")
}

var _ IService = (*CollectionReceiptService)(nil)

var _ interface {
	GenerateAndStorePDF(ctx context.Context, receiptID int64) error
	SendCollectionReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error
} = (*CollectionReceiptService)(nil)

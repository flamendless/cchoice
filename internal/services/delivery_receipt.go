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

type DeliveryReceipt struct {
	ID             string
	ReceiptNumber  string
	InvoiceID      string
	InvoiceNumber  string
	DeliveredTo    string
	RecipientEmail string
	TIN            string
	Address        string
	ReceiptDate    string
	Terms          string
	PONumber       string
	Status         enums.ReceiptStatus
	PDFPath        string
	EmailedAt      string
	CreatedAt      string
}

type DeliveryReceiptLine struct {
	Quantity    string
	Unit        string
	Description string
}

type DeliveryReceiptListItem struct {
	ID             string
	ReceiptNumber  string
	DeliveredTo    string
	RecipientEmail string
	Status         enums.ReceiptStatus
	ReceiptDate    string
	Emailed        bool
	PDFReady       bool
	CreatedAt      string
}

type DeliveryReceiptLineInput struct {
	Quantity    string
	Unit        string
	Description string
}

type CreateDeliveryReceiptInput struct {
	InvoiceID      string
	DeliveredTo    string
	RecipientEmail string
	TIN            string
	Address        string
	ReceiptDate    string
	Terms          string
	PONumber       string
	Lines          []DeliveryReceiptLineInput
}

type DeliveryReceiptRenderData struct {
	Config  InvoiceConfig
	Receipt DeliveryReceipt
	Lines   []DeliveryReceiptLine
}

type DeliveryReceiptJobStatusView struct {
	PDFStatus   string
	EmailStatus string
	PDFError    string
	EmailError  string
}

type DeliveryReceiptService struct {
	encoder        encode.IEncode
	dbRO           database.IService
	dbRW           database.IService
	staffLog       *StaffLogsService
	mailService    mail.IMailService
	invoiceService *InvoiceService
}

func NewDeliveryReceiptService(
	encoder encode.IEncode,
	dbRO database.IService,
	dbRW database.IService,
	staffLog *StaffLogsService,
	mailService mail.IMailService,
	invoiceService *InvoiceService,
) *DeliveryReceiptService {
	return &DeliveryReceiptService{
		encoder:        encoder,
		dbRO:           dbRO,
		dbRW:           dbRW,
		staffLog:       staffLog,
		mailService:    mailService,
		invoiceService: invoiceService,
	}
}

func (s *DeliveryReceiptService) CreateDeliveryReceipt(ctx context.Context, staffID string, in CreateDeliveryReceiptInput) (DeliveryReceipt, error) {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionCreate, constants.ModuleDeliveryReceipts, result, nil); err != nil {
			logs.Log().Warn("[DeliveryReceiptService] create delivery receipt log", zap.Error(err))
		}
	}()

	in, err := s.applyInvoicePrefill(ctx, in)
	if err != nil {
		result = err.Error()
		return DeliveryReceipt{}, err
	}

	lines := s.resolveLines(in.Lines)
	if len(lines) == 0 {
		result = errs.ErrDeliveryReceiptNoLines.Error()
		return DeliveryReceipt{}, errs.ErrDeliveryReceiptNoLines
	}

	staffDBID := s.encoder.Decode(staffID)
	if staffDBID == encode.INVALID {
		result = errs.ErrDecode.Error()
		return DeliveryReceipt{}, errs.ErrDecode
	}

	receiptDate := strings.TrimSpace(in.ReceiptDate)
	if receiptDate == "" {
		receiptDate = utils.NowPH().Format(constants.DateLayoutISO)
	}

	invoiceID := sql.NullInt64{}
	if strings.TrimSpace(in.InvoiceID) != "" {
		decoded := s.encoder.Decode(in.InvoiceID)
		if decoded == encode.INVALID {
			result = errs.ErrDecode.Error()
			return DeliveryReceipt{}, errs.ErrDecode
		}
		invoiceID = sql.NullInt64{Int64: decoded, Valid: true}
	}

	tx, err := s.dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		result = err.Error()
		return DeliveryReceipt{}, errors.Join(errs.ErrDeliveryReceiptCreateFailed, err)
	}
	defer func() { _ = tx.Rollback() }()
	qtx := s.dbRW.GetQueries().WithTx(tx)

	receiptDBID, err := qtx.CreateDeliveryReceipt(ctx, queries.CreateDeliveryReceiptParams{
		ReceiptNumber:  "",
		InvoiceID:      invoiceID,
		DeliveredTo:    strings.TrimSpace(in.DeliveredTo),
		RecipientEmail: strings.TrimSpace(in.RecipientEmail),
		Tin:            strings.TrimSpace(in.TIN),
		Address:        strings.TrimSpace(in.Address),
		ReceiptDate:    receiptDate,
		Terms:          strings.TrimSpace(in.Terms),
		PoNumber:       strings.TrimSpace(in.PONumber),
		Status:         enums.RECEIPT_STATUS_PROCESSING.String(),
		CreatedBy:      staffDBID,
	})
	if err != nil {
		result = err.Error()
		return DeliveryReceipt{}, errors.Join(errs.ErrDeliveryReceiptCreateFailed, err)
	}

	receiptNumber := fmt.Sprintf("DR-%s-%05d", strings.ReplaceAll(receiptDate, "-", ""), receiptDBID)
	if err := qtx.SetDeliveryReceiptNumber(ctx, queries.SetDeliveryReceiptNumberParams{
		ReceiptNumber: receiptNumber,
		ID:            receiptDBID,
	}); err != nil {
		result = err.Error()
		return DeliveryReceipt{}, errors.Join(errs.ErrDeliveryReceiptCreateFailed, err)
	}

	for i, ln := range lines {
		if err := qtx.CreateDeliveryReceiptLine(ctx, queries.CreateDeliveryReceiptLineParams{
			DeliveryReceiptID: receiptDBID,
			Quantity:          ln.Quantity,
			Unit:              ln.Unit,
			Description:       ln.Description,
			SortOrder:         int64(i),
		}); err != nil {
			result = err.Error()
			return DeliveryReceipt{}, errors.Join(errs.ErrDeliveryReceiptCreateFailed, err)
		}
	}

	if err := tx.Commit(); err != nil {
		result = err.Error()
		return DeliveryReceipt{}, errors.Join(errs.ErrDeliveryReceiptCreateFailed, err)
	}

	idStr := s.encoder.Encode(receiptDBID)
	result = fmt.Sprintf("success. ID '%s' number '%s'", idStr, receiptNumber)

	receipt, _, err := s.GetDeliveryReceipt(ctx, idStr)
	if err != nil {
		return DeliveryReceipt{}, err
	}
	return receipt, nil
}

func (s *DeliveryReceiptService) applyInvoicePrefill(ctx context.Context, in CreateDeliveryReceiptInput) (CreateDeliveryReceiptInput, error) {
	if strings.TrimSpace(in.InvoiceID) == "" || s.invoiceService == nil {
		return in, nil
	}

	invoice, invLines, err := s.invoiceService.GetInvoice(ctx, in.InvoiceID)
	if err != nil {
		return in, err
	}

	if strings.TrimSpace(in.DeliveredTo) == "" {
		in.DeliveredTo = cmpOr(invoice.RecipientRegisteredName, invoice.RecipientName)
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
	if strings.TrimSpace(in.ReceiptDate) == "" {
		in.ReceiptDate = cmpOr(invoice.DeliveryDate, invoice.IssueDate)
	}
	if strings.TrimSpace(in.Terms) == "" {
		in.Terms = formatPaymentTerms(invoice.PaymentTermsValue, invoice.PaymentTermsUnit)
	}
	if len(in.Lines) == 0 {
		for _, ln := range invLines {
			in.Lines = append(in.Lines, DeliveryReceiptLineInput{
				Quantity:    strconv.FormatInt(ln.Quantity, 10),
				Unit:        "pc",
				Description: ln.Description,
			})
		}
	}
	return in, nil
}

func formatPaymentTerms(value int64, unit string) string {
	if value <= 0 || strings.TrimSpace(unit) == "" {
		return ""
	}
	u := enums.ParsePaymentTermsUnitToEnum(unit)
	if u == enums.PAYMENT_TERMS_UNDEFINED {
		return ""
	}
	return strconv.FormatInt(value, 10) + " " + u.String()
}

func (s *DeliveryReceiptService) resolveLines(inputs []DeliveryReceiptLineInput) []DeliveryReceiptLine {
	lines := make([]DeliveryReceiptLine, 0, len(inputs))
	for _, li := range inputs {
		description := strings.TrimSpace(li.Description)
		if description == "" {
			continue
		}
		qty := strings.TrimSpace(li.Quantity)
		if qty == "" {
			qty = "1"
		}
		lines = append(lines, DeliveryReceiptLine{
			Quantity:    qty,
			Unit:        strings.TrimSpace(li.Unit),
			Description: description,
		})
	}
	return lines
}

func (s *DeliveryReceiptService) mapReceipt(r queries.TblDeliveryReceipt) DeliveryReceipt {
	receipt := DeliveryReceipt{
		ID:             s.encoder.Encode(r.ID),
		ReceiptNumber:  r.ReceiptNumber,
		DeliveredTo:    r.DeliveredTo,
		RecipientEmail: r.RecipientEmail,
		TIN:            r.Tin,
		Address:        r.Address,
		ReceiptDate:    r.ReceiptDate,
		Terms:          r.Terms,
		PONumber:       r.PoNumber,
		Status:         enums.ParseReceiptStatusToEnum(r.Status),
		PDFPath:        r.PdfPath,
		EmailedAt:      utils.ConvertToPH(r.EmailedAt),
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
	if r.InvoiceID.Valid {
		receipt.InvoiceID = s.encoder.Encode(r.InvoiceID.Int64)
	}
	return receipt
}

func (s *DeliveryReceiptService) enrichInvoiceNumber(ctx context.Context, receipt *DeliveryReceipt) {
	if strings.TrimSpace(receipt.InvoiceID) == "" || s.invoiceService == nil {
		return
	}
	invoice, _, err := s.invoiceService.GetInvoice(ctx, receipt.InvoiceID)
	if err != nil {
		return
	}
	receipt.InvoiceNumber = invoice.InvoiceNumber
}

func (s *DeliveryReceiptService) GetDeliveryReceipt(ctx context.Context, id string) (DeliveryReceipt, []DeliveryReceiptLine, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return DeliveryReceipt{}, nil, errs.ErrDecode
	}

	row, err := s.dbRO.GetQueries().GetDeliveryReceiptByID(ctx, decoded)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DeliveryReceipt{}, nil, errs.ErrDeliveryReceiptNotFound
		}
		return DeliveryReceipt{}, nil, errors.Join(errs.ErrDeliveryReceipt, err)
	}

	lineRows, err := s.dbRO.GetQueries().GetDeliveryReceiptLinesByReceiptID(ctx, decoded)
	if err != nil {
		return DeliveryReceipt{}, nil, errors.Join(errs.ErrDeliveryReceipt, err)
	}

	receipt := s.mapReceipt(row.TblDeliveryReceipt)
	s.enrichInvoiceNumber(ctx, &receipt)

	lines := make([]DeliveryReceiptLine, 0, len(lineRows))
	for _, lr := range lineRows {
		l := lr.TblDeliveryReceiptLine
		lines = append(lines, DeliveryReceiptLine{
			Quantity:    l.Quantity,
			Unit:        l.Unit,
			Description: l.Description,
		})
	}

	return receipt, lines, nil
}

func (s *DeliveryReceiptService) GetDeliveryReceiptDBID(ctx context.Context, id string) (int64, error) {
	decoded := s.encoder.Decode(id)
	if decoded == encode.INVALID {
		return 0, errs.ErrDecode
	}
	return decoded, nil
}

func (s *DeliveryReceiptService) mapListItem(r queries.ListDeliveryReceiptsPaginatedRow) DeliveryReceiptListItem {
	return DeliveryReceiptListItem{
		ID:             s.encoder.Encode(r.ID),
		ReceiptNumber:  r.ReceiptNumber,
		DeliveredTo:    r.DeliveredTo,
		RecipientEmail: r.RecipientEmail,
		Status:         enums.ParseReceiptStatusToEnum(r.Status),
		ReceiptDate:    r.ReceiptDate,
		Emailed:        r.EmailedAt != "",
		PDFReady:       strings.TrimSpace(r.PdfPath) != "",
		CreatedAt:      utils.ConvertToPH(r.CreatedAt),
	}
}

func (s *DeliveryReceiptService) GetDeliveryReceiptsPaginated(ctx context.Context, page, perPage int) ([]DeliveryReceiptListItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = constants.DefaultAdminTablePageSize
	}

	total, err := s.dbRO.GetQueries().CountDeliveryReceipts(ctx)
	if err != nil {
		return nil, 0, errors.Join(errs.ErrDeliveryReceipt, err)
	}

	offset := int64((page - 1) * perPage)
	rows, err := s.dbRO.GetQueries().ListDeliveryReceiptsPaginated(ctx, queries.ListDeliveryReceiptsPaginatedParams{
		Limit:  int64(perPage),
		Offset: offset,
	})
	if err != nil {
		return nil, 0, errors.Join(errs.ErrDeliveryReceipt, err)
	}

	result := make([]DeliveryReceiptListItem, 0, len(rows))
	for _, r := range rows {
		result = append(result, s.mapListItem(r))
	}
	return result, total, nil
}

func (s *DeliveryReceiptService) GetJobStatus(ctx context.Context, receiptID int64) (DeliveryReceiptJobStatusView, error) {
	view := DeliveryReceiptJobStatusView{PDFStatus: "none", EmailStatus: "none"}
	if row, err := s.dbRO.GetQueries().GetLatestDeliveryReceiptJobByReceiptIDAndType(ctx, queries.GetLatestDeliveryReceiptJobByReceiptIDAndTypeParams{
		DeliveryReceiptID: receiptID,
		JobType:           enums.DELIVERY_RECEIPT_JOB_GENERATE_PDF.String(),
	}); err == nil {
		view.PDFStatus = strings.ToLower(row.TblDeliveryReceiptJob.Status)
		view.PDFError = row.TblDeliveryReceiptJob.ErrorMessage
	}
	if row, err := s.dbRO.GetQueries().GetLatestDeliveryReceiptJobByReceiptIDAndType(ctx, queries.GetLatestDeliveryReceiptJobByReceiptIDAndTypeParams{
		DeliveryReceiptID: receiptID,
		JobType:           enums.DELIVERY_RECEIPT_JOB_SEND_EMAIL.String(),
	}); err == nil {
		view.EmailStatus = strings.ToLower(row.TblDeliveryReceiptJob.Status)
		view.EmailError = row.TblDeliveryReceiptJob.ErrorMessage
	}
	return view, nil
}

func deliveryReceiptPDFDir() string {
	return filepath.Join("cmd", "web", "static", "delivery_receipts")
}

func (s *DeliveryReceiptService) GenerateAndStorePDF(ctx context.Context, receiptID int64) error {
	idStr := s.encoder.Encode(receiptID)
	receipt, lines, err := s.GetDeliveryReceipt(ctx, idStr)
	if err != nil {
		return err
	}

	config, err := s.invoiceService.GetConfig(ctx)
	if err != nil {
		return err
	}

	pdfBytes, err := RenderDeliveryReceiptPDF(config, receipt, lines)
	if err != nil {
		return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err)
	}

	if err := os.MkdirAll(deliveryReceiptPDFDir(), 0o755); err != nil {
		return err
	}
	localPath := filepath.Join(deliveryReceiptPDFDir(), strconv.FormatInt(receiptID, 10)+".pdf")
	if err := os.WriteFile(localPath, pdfBytes, 0o644); err != nil {
		return err
	}

	if err := s.dbRW.GetQueries().SetDeliveryReceiptPDFPath(ctx, queries.SetDeliveryReceiptPDFPathParams{
		PdfPath: localPath,
		ID:      receiptID,
	}); err != nil {
		return errors.Join(errs.ErrDeliveryReceipt, err)
	}
	return nil
}

func (s *DeliveryReceiptService) ReadStoredPDF(receiptID int64, pdfPath string) ([]byte, error) {
	path := strings.TrimSpace(pdfPath)
	if path == "" {
		path = filepath.Join(deliveryReceiptPDFDir(), strconv.FormatInt(receiptID, 10)+".pdf")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Join(errs.ErrDeliveryReceiptPDFFailed, err)
	}
	return data, nil
}

func (s *DeliveryReceiptService) SendDeliveryReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error {
	return s.SendDeliveryReceiptEmail(ctx, staffID, s.encoder.Encode(receiptID))
}

func (s *DeliveryReceiptService) SendDeliveryReceiptEmail(ctx context.Context, staffID string, id string) error {
	result := "success"
	defer func() {
		if err := s.staffLog.CreateLog(ctx, staffID, constants.ActionSend, constants.ModuleDeliveryReceipts, result, nil); err != nil {
			logs.Log().Warn("[DeliveryReceiptService] send delivery receipt log", zap.Error(err))
		}
	}()

	if s.mailService == nil {
		result = errs.ErrDeliveryReceiptEmailNotConfigured.Error()
		return errs.ErrDeliveryReceiptEmailNotConfigured
	}

	receipt, lines, err := s.GetDeliveryReceipt(ctx, id)
	if err != nil {
		result = err.Error()
		return err
	}
	if strings.TrimSpace(receipt.RecipientEmail) == "" {
		result = errs.ErrDeliveryReceiptRecipientNoEmail.Error()
		return errs.ErrDeliveryReceiptRecipientNoEmail
	}

	decoded := s.encoder.Decode(id)
	if strings.TrimSpace(receipt.PDFPath) == "" {
		if err := s.GenerateAndStorePDF(ctx, decoded); err != nil {
			result = err.Error()
			return err
		}
		receipt, lines, err = s.GetDeliveryReceipt(ctx, id)
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
		return errors.Join(errs.ErrDeliveryReceiptPDFFailed, err)
	}

	businessName := cmpOr(config.BusinessName, "C-Choice")
	subject := fmt.Sprintf("Delivery Receipt %s from %s", receipt.ReceiptNumber, businessName)
	data := s.BuildEmailTemplateData(config, receipt, lines)

	attachments := []mail.Attachment{
		{
			FileName:    cmpOr(receipt.ReceiptNumber, "delivery_receipt") + ".pdf",
			ContentType: "application/pdf",
			Content:     pdfBytes,
		},
	}

	if err := s.mailService.SendTemplateEmailWithAttachments(
		receipt.RecipientEmail,
		nil,
		subject,
		enums.EMAIL_TEMPLATE_DELIVERY_RECEIPT.FileName(),
		data,
		attachments,
	); err != nil {
		result = err.Error()
		return errors.Join(errs.ErrDeliveryReceiptEmailFailed, err)
	}

	if err := s.dbRW.GetQueries().MarkDeliveryReceiptEmailed(ctx, decoded); err != nil {
		logs.LogCtx(ctx).Warn("[DeliveryReceiptService] mark emailed", zap.Error(err))
	}

	return nil
}

func (s *DeliveryReceiptService) ID() string {
	return "DeliveryReceipt"
}

func (s *DeliveryReceiptService) Log() {
	logs.Log().Info("[DeliveryReceiptService] Loaded")
}

var _ IService = (*DeliveryReceiptService)(nil)

var _ interface {
	GenerateAndStorePDF(ctx context.Context, receiptID int64) error
	SendDeliveryReceiptEmailByID(ctx context.Context, staffID string, receiptID int64) error
} = (*DeliveryReceiptService)(nil)

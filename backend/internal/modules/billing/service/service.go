package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stayflow/stayflow-track/internal/modules/billing/domain"
	"github.com/stayflow/stayflow-track/internal/modules/billing/repository"
	"github.com/stayflow/stayflow-track/internal/platform/storage"
	apperrors "github.com/stayflow/stayflow-track/internal/shared/errors"
)

type Service struct {
	repo  *repository.Repository
	store storage.Store
}

func New(repo *repository.Repository, store storage.Store) *Service {
	return &Service{repo: repo, store: store}
}

type AddChargeInput struct {
	FolioID     uuid.UUID       `json:"folio_id" validate:"required"`
	Category    string          `json:"category" validate:"required,oneof=room_charge food_beverage laundry minibar parking spa damage late_checkout extra_bed other"`
	Description string          `json:"description" validate:"required,min=1,max=500"`
	Quantity    int             `json:"quantity" validate:"required,min=1"`
	UnitPrice   decimal.Decimal `json:"unit_price" validate:"required"`
	TaxRate     decimal.Decimal `json:"tax_rate"`
	Date        string          `json:"date" validate:"omitempty"`
}

type RecordPaymentInput struct {
	FolioID         uuid.UUID       `json:"folio_id" validate:"required"`
	InvoiceID       *uuid.UUID      `json:"invoice_id"`
	PaymentType     string          `json:"payment_type" validate:"required,oneof=payment refund deposit deposit_release"`
	PaymentMethod   string          `json:"payment_method" validate:"required,oneof=cash upi card bank_transfer"`
	Amount          decimal.Decimal `json:"amount" validate:"required"`
	ReferenceNumber string          `json:"reference_number" validate:"omitempty,max=255"`
	Notes           string          `json:"notes" validate:"omitempty,max=500"`
}

type VoidItemInput struct {
	LineItemID uuid.UUID `json:"line_item_id" validate:"required"`
	Reason     string    `json:"reason" validate:"required,min=3,max=500"`
}

func (s *Service) GetFolio(ctx context.Context, folioID, tenantID uuid.UUID) (*domain.Folio, error) {
	return s.repo.GetFolioByID(ctx, folioID, tenantID)
}

func (s *Service) GetFolioByReservation(ctx context.Context, reservationID, tenantID uuid.UUID) (*domain.Folio, error) {
	return s.repo.GetFolioByReservation(ctx, reservationID, tenantID)
}

func (s *Service) GetFolioSummary(ctx context.Context, folioID, tenantID uuid.UUID) (*repository.FolioSummary, error) {
	return s.repo.GetFolioSummary(ctx, folioID, tenantID)
}

func (s *Service) AddCharge(ctx context.Context, tenantID, userID uuid.UUID, input AddChargeInput) (*domain.LineItem, error) {
	date := time.Now()
	if input.Date != "" {
		parsed, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			return nil, apperrors.BadRequest("invalid date format, use YYYY-MM-DD")
		}
		date = parsed
	}

	item := &domain.LineItem{
		TenantID:    tenantID,
		FolioID:     input.FolioID,
		Category:    domain.LineItemCategory(input.Category),
		Description: input.Description,
		Quantity:    input.Quantity,
		UnitPrice:   input.UnitPrice,
		TaxRate:     input.TaxRate,
		Date:        date,
		CreatedBy:   userID,
	}

	if err := s.repo.AddLineItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) VoidLineItem(ctx context.Context, tenantID uuid.UUID, input VoidItemInput) error {
	return s.repo.VoidLineItem(ctx, input.LineItemID, tenantID, input.Reason)
}

func (s *Service) ListLineItems(ctx context.Context, folioID, tenantID uuid.UUID) ([]domain.LineItem, error) {
	return s.repo.ListLineItems(ctx, folioID, tenantID)
}

func (s *Service) RecordPayment(ctx context.Context, tenantID, userID uuid.UUID, input RecordPaymentInput) (*domain.Payment, error) {
	// Validate folio exists
	folio, err := s.repo.GetFolioByID(ctx, input.FolioID, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate refund doesn't exceed paid amount
	if input.PaymentType == "refund" || input.PaymentType == "deposit_release" {
		if input.Amount.GreaterThan(folio.PaidAmount) {
			return nil, apperrors.BadRequest("refund amount cannot exceed total payments received")
		}
	}

	payment := &domain.Payment{
		TenantID:        tenantID,
		FolioID:         input.FolioID,
		InvoiceID:       input.InvoiceID,
		PaymentType:     domain.PaymentType(input.PaymentType),
		PaymentMethod:   domain.PaymentMethod(input.PaymentMethod),
		Amount:          input.Amount,
		ReferenceNumber: input.ReferenceNumber,
		Notes:           input.Notes,
		ReceivedBy:      userID,
	}

	if err := s.repo.RecordPayment(ctx, payment); err != nil {
		return nil, apperrors.Internal(err)
	}

	return payment, nil
}

func (s *Service) ListPayments(ctx context.Context, folioID, tenantID uuid.UUID) ([]domain.Payment, error) {
	return s.repo.ListPayments(ctx, folioID, tenantID)
}

func (s *Service) GetInvoice(ctx context.Context, invoiceID, tenantID uuid.UUID) (*domain.Invoice, error) {
	return s.repo.GetInvoiceByID(ctx, invoiceID, tenantID)
}

func (s *Service) ListInvoices(ctx context.Context, tenantID uuid.UUID, propertyID *uuid.UUID, status string, limit, offset int) ([]domain.Invoice, int64, error) {
	return s.repo.ListInvoices(ctx, tenantID, propertyID, status, limit, offset)
}

// GenerateInvoicePDF creates a PDF and stores it in S3.
func (s *Service) GenerateInvoicePDF(ctx context.Context, invoiceID, tenantID uuid.UUID) (string, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID, tenantID)
	if err != nil {
		return "", err
	}

	pdfBytes, err := generatePDF(invoice)
	if err != nil {
		return "", apperrors.Internal(fmt.Errorf("generate pdf: %w", err))
	}

	key := fmt.Sprintf("tenants/%s/invoices/%s/%s.pdf",
		tenantID.String(), invoice.InvoiceNumber, invoiceID.String())

	reader := bytes.NewReader(pdfBytes)
	_, err = s.store.Upload(ctx, key, reader, "application/pdf", int64(len(pdfBytes)))
	if err != nil {
		return "", apperrors.Internal(fmt.Errorf("upload pdf: %w", err))
	}

	if err := s.repo.UpdateInvoicePDF(ctx, invoiceID, key); err != nil {
		return "", apperrors.Internal(err)
	}

	url, err := s.store.GetPresignedURL(ctx, key, 24*time.Hour)
	if err != nil {
		return "", apperrors.Internal(err)
	}

	return url, nil
}

// GetInvoicePDFURL returns a presigned download URL for an invoice PDF.
func (s *Service) GetInvoicePDFURL(ctx context.Context, invoiceID, tenantID uuid.UUID) (string, error) {
	invoice, err := s.repo.GetInvoiceByID(ctx, invoiceID, tenantID)
	if err != nil {
		return "", err
	}

	if invoice.PDFKey == "" {
		return "", apperrors.NotFound("invoice_pdf", invoiceID.String())
	}

	url, err := s.store.GetPresignedURL(ctx, invoice.PDFKey, 1*time.Hour)
	if err != nil {
		return "", apperrors.Internal(err)
	}

	return url, nil
}

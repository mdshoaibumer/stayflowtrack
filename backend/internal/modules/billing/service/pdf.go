package service

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stayflow/stayflow-track/internal/modules/billing/domain"
)

// generatePDF creates a simple text-based invoice representation.
// In production, replace with a proper PDF library (e.g., gofpdf, unidoc).
func generatePDF(invoice *domain.Invoice) ([]byte, error) {
	var buf bytes.Buffer

	// Header
	buf.WriteString("========================================\n")
	buf.WriteString("              TAX INVOICE               \n")
	buf.WriteString("========================================\n\n")
	buf.WriteString(fmt.Sprintf("Invoice No: %s\n", invoice.InvoiceNumber))
	buf.WriteString(fmt.Sprintf("Date: %s\n\n", invoice.IssuedAt.Format("02-Jan-2006")))

	// Property details
	buf.WriteString("FROM:\n")
	buf.WriteString(fmt.Sprintf("%s\n", invoice.PropertyName))
	buf.WriteString(fmt.Sprintf("%s\n", invoice.PropertyAddress))
	if invoice.PropertyGST != "" {
		buf.WriteString(fmt.Sprintf("GSTIN: %s\n", invoice.PropertyGST))
	}
	buf.WriteString("\n")

	// Guest details
	buf.WriteString("BILL TO:\n")
	buf.WriteString(fmt.Sprintf("%s\n", invoice.GuestName))
	if invoice.GuestPhone != "" {
		buf.WriteString(fmt.Sprintf("Phone: %s\n", invoice.GuestPhone))
	}
	if invoice.GuestEmail != "" {
		buf.WriteString(fmt.Sprintf("Email: %s\n", invoice.GuestEmail))
	}
	if invoice.GuestAddress != "" {
		buf.WriteString(fmt.Sprintf("Address: %s\n", invoice.GuestAddress))
	}
	buf.WriteString("\n")

	// Stay details
	buf.WriteString("Stay Details:\n")
	buf.WriteString(fmt.Sprintf("Check-in: %s\n", invoice.CheckInDate.Format("02-Jan-2006")))
	buf.WriteString(fmt.Sprintf("Check-out: %s\n", invoice.CheckOutDate.Format("02-Jan-2006")))
	buf.WriteString(fmt.Sprintf("Nights: %d\n\n", invoice.NumNights))

	// Line items
	buf.WriteString(fmt.Sprintf("%-30s %5s %10s %6s %10s %10s\n",
		"Description", "Qty", "Rate", "Tax%", "Tax", "Total"))
	buf.WriteString(strings.Repeat("-", 80) + "\n")

	for _, item := range invoice.LineItems {
		buf.WriteString(fmt.Sprintf("%-30s %5d %10s %5s%% %10s %10s\n",
			truncate(item.Description, 30), item.Quantity, item.UnitPrice.StringFixed(2),
			item.TaxRate.StringFixed(0), item.TaxAmount.StringFixed(2), item.Total.StringFixed(2)))
	}

	buf.WriteString(strings.Repeat("-", 80) + "\n")

	// Totals
	buf.WriteString(fmt.Sprintf("%68s %10s\n", "Subtotal:", invoice.Subtotal.StringFixed(2)))
	if invoice.CGSTAmount.IsPositive() {
		buf.WriteString(fmt.Sprintf("%68s %10s\n", "CGST:", invoice.CGSTAmount.StringFixed(2)))
	}
	if invoice.SGSTAmount.IsPositive() {
		buf.WriteString(fmt.Sprintf("%68s %10s\n", "SGST:", invoice.SGSTAmount.StringFixed(2)))
	}
	if invoice.IGSTAmount.IsPositive() {
		buf.WriteString(fmt.Sprintf("%68s %10s\n", "IGST:", invoice.IGSTAmount.StringFixed(2)))
	}
	buf.WriteString(fmt.Sprintf("%68s %10s\n", "Total Tax:", invoice.TotalTax.StringFixed(2)))
	buf.WriteString(strings.Repeat("=", 80) + "\n")
	buf.WriteString(fmt.Sprintf("%68s %10s\n", "TOTAL:", invoice.TotalAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("%68s %10s\n", "Paid:", invoice.PaidAmount.StringFixed(2)))
	buf.WriteString(fmt.Sprintf("%68s %10s\n", "Balance Due:", invoice.BalanceDue.StringFixed(2)))

	buf.WriteString("\n\nThank you for staying with us!\n")

	return buf.Bytes(), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

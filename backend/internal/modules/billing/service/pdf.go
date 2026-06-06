package service

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/stayflow/stayflow-track/internal/modules/billing/domain"
)

// generatePDF creates a proper GST-compliant PDF invoice.
func generatePDF(invoice *domain.Invoice) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header - TAX INVOICE title
	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(180, 10, "TAX INVOICE", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Invoice Number & Date row
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(90, 6, fmt.Sprintf("Invoice No: %s", invoice.InvoiceNumber), "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("Date: %s", invoice.IssuedAt.Format("02-Jan-2006")), "", 1, "R", false, 0, "")
	pdf.Ln(2)

	// Divider
	pdf.SetDrawColor(0, 0, 0)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(4)

	// FROM / BILL TO section
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(90, 6, "FROM:", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, "BILL TO:", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(90, 5, invoice.PropertyName, "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 5, invoice.GuestName, "", 1, "L", false, 0, "")

	pdf.CellFormat(90, 5, invoice.PropertyAddress, "", 0, "L", false, 0, "")
	guestLine2 := ""
	if invoice.GuestPhone != "" {
		guestLine2 = fmt.Sprintf("Phone: %s", invoice.GuestPhone)
	}
	pdf.CellFormat(90, 5, guestLine2, "", 1, "L", false, 0, "")

	if invoice.PropertyGST != "" {
		pdf.CellFormat(90, 5, fmt.Sprintf("GSTIN: %s", invoice.PropertyGST), "", 0, "L", false, 0, "")
	} else {
		pdf.CellFormat(90, 5, "", "", 0, "L", false, 0, "")
	}
	guestLine3 := ""
	if invoice.GuestEmail != "" {
		guestLine3 = fmt.Sprintf("Email: %s", invoice.GuestEmail)
	}
	pdf.CellFormat(90, 5, guestLine3, "", 1, "L", false, 0, "")

	if invoice.GuestAddress != "" {
		pdf.CellFormat(90, 5, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(90, 5, invoice.GuestAddress, "", 1, "L", false, 0, "")
	}

	pdf.Ln(4)

	// Stay Details Box
	pdf.SetFillColor(245, 245, 245)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(60, 7, "Check-in", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "Check-out", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, "No. of Nights", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(60, 7, invoice.CheckInDate.Format("02-Jan-2006"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 7, invoice.CheckOutDate.Format("02-Jan-2006"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", invoice.NumNights), "1", 1, "C", false, 0, "")

	pdf.Ln(6)

	// Line Items Table Header
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetFillColor(50, 50, 50)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(8, 7, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(52, 7, "Description", "1", 0, "L", true, 0, "")
	pdf.CellFormat(18, 7, "SAC/HSN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(12, 7, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Rate", "1", 0, "R", true, 0, "")
	pdf.CellFormat(18, 7, "Tax %", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 7, "Tax", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 7, "Total", "1", 1, "R", true, 0, "")

	// Line Items
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Helvetica", "", 8)
	for i, item := range invoice.LineItems {
		sacCode := getSACCode(item.Category)
		pdf.CellFormat(8, 6, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
		pdf.CellFormat(52, 6, truncateStr(item.Description, 35), "1", 0, "L", false, 0, "")
		pdf.CellFormat(18, 6, sacCode, "1", 0, "C", false, 0, "")
		pdf.CellFormat(12, 6, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%s", item.UnitPrice.StringFixed(2)), "1", 0, "R", false, 0, "")
		pdf.CellFormat(18, 6, fmt.Sprintf("%s%%", item.TaxRate.StringFixed(0)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(22, 6, fmt.Sprintf("%s", item.TaxAmount.StringFixed(2)), "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%s", item.Total.StringFixed(2)), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(4)

	// Totals section - right aligned
	totalsX := float64(120)
	pdf.SetFont("Helvetica", "", 9)

	pdf.SetX(totalsX)
	pdf.CellFormat(40, 6, "Subtotal:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.Subtotal.StringFixed(2)), "0", 1, "R", false, 0, "")

	if invoice.CGSTAmount.IsPositive() {
		pdf.SetX(totalsX)
		pdf.CellFormat(40, 6, "CGST:", "0", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.CGSTAmount.StringFixed(2)), "0", 1, "R", false, 0, "")
	}
	if invoice.SGSTAmount.IsPositive() {
		pdf.SetX(totalsX)
		pdf.CellFormat(40, 6, "SGST:", "0", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.SGSTAmount.StringFixed(2)), "0", 1, "R", false, 0, "")
	}
	if invoice.IGSTAmount.IsPositive() {
		pdf.SetX(totalsX)
		pdf.CellFormat(40, 6, "IGST:", "0", 0, "R", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.IGSTAmount.StringFixed(2)), "0", 1, "R", false, 0, "")
	}

	pdf.SetX(totalsX)
	pdf.CellFormat(40, 6, "Total Tax:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.TotalTax.StringFixed(2)), "0", 1, "R", false, 0, "")

	// Grand total line
	pdf.SetX(totalsX)
	pdf.Line(totalsX, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetX(totalsX)
	pdf.CellFormat(40, 7, "GRAND TOTAL:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, fmt.Sprintf("INR %s", invoice.TotalAmount.StringFixed(2)), "0", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetX(totalsX)
	pdf.CellFormat(40, 6, "Amount Paid:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.PaidAmount.StringFixed(2)), "0", 1, "R", false, 0, "")

	if invoice.BalanceDue.IsPositive() {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetX(totalsX)
		pdf.CellFormat(40, 6, "Balance Due:", "0", 0, "R", false, 0, "")
		pdf.SetTextColor(200, 0, 0)
		pdf.CellFormat(35, 6, fmt.Sprintf("INR %s", invoice.BalanceDue.StringFixed(2)), "0", 1, "R", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	}

	pdf.Ln(10)

	// Footer
	pdf.SetFont("Helvetica", "I", 8)
	pdf.CellFormat(180, 5, "This is a computer-generated invoice and does not require a signature.", "", 1, "C", false, 0, "")
	pdf.CellFormat(180, 5, "Thank you for staying with us!", "", 1, "C", false, 0, "")

	// Amount in words at bottom
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(180, 5, fmt.Sprintf("Amount in words: %s", amountToWords(invoice.TotalAmount.IntPart())), "", 1, "L", false, 0, "")

	// Write to buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("pdf output: %w", err)
	}

	return buf.Bytes(), nil
}

// getSACCode returns the SAC/HSN code for a line item category (GST compliance).
func getSACCode(category string) string {
	switch category {
	case "room_charge":
		return "996311" // Accommodation services
	case "food_beverage":
		return "996331" // Food & beverage
	case "laundry":
		return "998112" // Laundry services
	case "parking":
		return "996742" // Parking services
	case "spa":
		return "999723" // Spa/wellness
	case "late_checkout":
		return "996311" // Same as room
	case "extra_bed":
		return "996311"
	default:
		return "999799" // Other services
	}
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// amountToWords converts an integer amount to Indian number words (simplified).
func amountToWords(n int64) string {
	if n == 0 {
		return "Zero Rupees"
	}

	ones := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
		"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	var convert func(n int64) string
	convert = func(n int64) string {
		if n == 0 {
			return ""
		} else if n < 20 {
			return ones[n]
		} else if n < 100 {
			return tens[n/10] + " " + ones[n%10]
		} else if n < 1000 {
			return ones[n/100] + " Hundred " + convert(n%100)
		} else if n < 100000 {
			return convert(n/1000) + " Thousand " + convert(n%1000)
		} else if n < 10000000 {
			return convert(n/100000) + " Lakh " + convert(n%100000)
		}
		return convert(n/10000000) + " Crore " + convert(n%10000000)
	}

	return convert(n) + " Rupees Only"
}

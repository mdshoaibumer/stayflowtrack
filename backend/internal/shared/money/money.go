package money

import "github.com/shopspring/decimal"

// Amount represents a monetary value with exact decimal precision.
// Use this for all financial calculations to avoid floating-point errors.
type Amount = decimal.Decimal

// Zero returns a zero amount.
func Zero() Amount {
	return decimal.Zero
}

// FromFloat converts a float64 to a precise decimal Amount.
// Use only for reading from external sources (JSON input, DB scan).
func FromFloat(f float64) Amount {
	return decimal.NewFromFloat(f)
}

// FromInt creates an Amount from an integer (e.g., quantity).
func FromInt(i int64) Amount {
	return decimal.NewFromInt(i)
}

// FromString parses a string to an Amount.
func FromString(s string) (Amount, error) {
	return decimal.NewFromString(s)
}

// Multiply returns a * b with full precision.
func Multiply(a, b Amount) Amount {
	return a.Mul(b)
}

// CalculateTax computes tax = amount * rate / 100, rounded to 2 decimal places.
func CalculateTax(amount Amount, ratePercent Amount) Amount {
	return amount.Mul(ratePercent).Div(decimal.NewFromInt(100)).Round(2)
}

// CalculateNightlyTotal returns nights * rate_per_night.
func CalculateNightlyTotal(nights int, ratePerNight Amount) Amount {
	return decimal.NewFromInt(int64(nights)).Mul(ratePerNight)
}

// Round2 rounds to 2 decimal places (standard for currency).
func Round2(a Amount) Amount {
	return a.Round(2)
}

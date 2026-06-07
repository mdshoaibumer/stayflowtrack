package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestZero(t *testing.T) {
	z := Zero()
	if !z.IsZero() {
		t.Error("Zero() should return zero")
	}
}

func TestFromInt(t *testing.T) {
	a := FromInt(5000)
	if !a.Equal(decimal.NewFromInt(5000)) {
		t.Errorf("FromInt(5000) = %s", a.String())
	}
}

func TestFromFloat(t *testing.T) {
	a := FromFloat(99.99)
	expected := decimal.NewFromFloat(99.99)
	if !a.Equal(expected) {
		t.Errorf("FromFloat(99.99) = %s", a.String())
	}
}

func TestFromString(t *testing.T) {
	a, err := FromString("1234.56")
	if err != nil {
		t.Fatalf("FromString error: %v", err)
	}
	expected, _ := decimal.NewFromString("1234.56")
	if !a.Equal(expected) {
		t.Errorf("FromString(\"1234.56\") = %s", a.String())
	}
}

func TestFromString_Invalid(t *testing.T) {
	_, err := FromString("not-a-number")
	if err == nil {
		t.Error("expected error for invalid string")
	}
}

func TestMultiply(t *testing.T) {
	a := FromInt(3000)
	b := FromInt(3)
	result := Multiply(a, b)
	if !result.Equal(FromInt(9000)) {
		t.Errorf("3000 * 3 = %s, want 9000", result.String())
	}
}

func TestCalculateTax(t *testing.T) {
	tests := []struct {
		name     string
		amount   Amount
		rate     Amount
		expected Amount
	}{
		{
			name:     "12% on 9000",
			amount:   FromInt(9000),
			rate:     FromInt(12),
			expected: FromInt(1080),
		},
		{
			name:     "18% on 2000",
			amount:   FromInt(2000),
			rate:     FromInt(18),
			expected: FromInt(360),
		},
		{
			name:     "5% on 1000",
			amount:   FromInt(1000),
			rate:     FromInt(5),
			expected: FromInt(50),
		},
		{
			name:     "28% on 5000",
			amount:   FromInt(5000),
			rate:     FromInt(28),
			expected: FromInt(1400),
		},
		{
			name:     "0% tax",
			amount:   FromInt(1000),
			rate:     Zero(),
			expected: Zero(),
		},
		{
			name:     "tax on fractional amount",
			amount:   FromFloat(999.50),
			rate:     FromInt(12),
			expected: Round2(FromFloat(999.50).Mul(FromInt(12)).Div(FromInt(100))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTax(tt.amount, tt.rate)
			if !result.Equal(tt.expected) {
				t.Errorf("CalculateTax(%s, %s%%) = %s, want %s",
					tt.amount.String(), tt.rate.String(), result.String(), tt.expected.String())
			}
		})
	}
}

func TestCalculateNightlyTotal(t *testing.T) {
	tests := []struct {
		nights   int
		rate     Amount
		expected Amount
	}{
		{3, FromInt(3000), FromInt(9000)},
		{1, FromInt(5000), FromInt(5000)},
		{7, FromInt(2500), FromInt(17500)},
		{0, FromInt(3000), Zero()},
	}

	for _, tt := range tests {
		result := CalculateNightlyTotal(tt.nights, tt.rate)
		if !result.Equal(tt.expected) {
			t.Errorf("CalculateNightlyTotal(%d, %s) = %s, want %s",
				tt.nights, tt.rate.String(), result.String(), tt.expected.String())
		}
	}
}

func TestRound2(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123.456", "123.46"},
		{"123.454", "123.45"},
		{"123.455", "123.46"},
		{"100", "100"},
		{"0.001", "0"},
		{"99.999", "100"},
	}

	for _, tt := range tests {
		a, _ := FromString(tt.input)
		result := Round2(a)
		if result.StringFixed(2) != decimal.RequireFromString(tt.expected).StringFixed(2) {
			t.Errorf("Round2(%s) = %s, want %s", tt.input, result.StringFixed(2), tt.expected)
		}
	}
}

func TestDecimalPrecision_NoFloatErrors(t *testing.T) {
	// Classic float64 issue: 0.1 + 0.2 != 0.3 in float64
	a, _ := FromString("0.1")
	b, _ := FromString("0.2")
	expected, _ := FromString("0.3")

	result := a.Add(b)
	if !result.Equal(expected) {
		t.Errorf("0.1 + 0.2 = %s, want 0.3 (decimal precision test)", result.String())
	}
}

func TestDecimalPrecision_LargeAmounts(t *testing.T) {
	// Test that large monetary amounts don't lose precision
	// 9999999.99 * 18 / 100 = 1799999.9982, rounds to 1800000.00
	amount, _ := FromString("9999999.99")
	tax := CalculateTax(amount, FromInt(18))

	expected, _ := FromString("1800000.00")
	if !tax.Equal(expected) {
		t.Errorf("18%% of 9999999.99 = %s, want %s", tax.String(), expected.String())
	}
}

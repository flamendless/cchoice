package utils

import (
	"strings"
	"time"

	"cchoice/internal/constants"
	"cchoice/internal/enums"
	"cchoice/internal/errs"
)

func AmountInWordsPHP(centavos int64) string {
	if centavos == 0 {
		return "Zero Pesos"
	}
	negative := centavos < 0
	if negative {
		centavos = -centavos
	}
	pesos := centavos / 100
	cents := centavos % 100

	var parts []string
	if negative {
		parts = append(parts, "Negative")
	}
	if pesos > 0 {
		parts = append(parts, numberToWords(pesos))
		if pesos == 1 {
			parts = append(parts, "Peso")
		} else {
			parts = append(parts, "Pesos")
		}
	}
	if cents > 0 {
		if pesos > 0 {
			parts = append(parts, "and")
		}
		parts = append(parts, numberToWords(cents))
		if cents == 1 {
			parts = append(parts, "Centavo")
		} else {
			parts = append(parts, "Centavos")
		}
	}
	return strings.Join(parts, " ")
}

func numberToWords(n int64) string {
	if n == 0 {
		return "Zero"
	}
	if n < 0 {
		return "Negative " + numberToWords(-n)
	}

	ones := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine",
		"Ten", "Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	switch {
	case n < 20:
		return ones[n]
	case n < 100:
		return strings.TrimSpace(tens[n/10] + " " + ones[n%10])
	case n < 1000:
		return strings.TrimSpace(ones[n/100] + " Hundred " + numberToWords(n%100))
	case n < 1_000_000:
		return strings.TrimSpace(numberToWords(n/1000) + " Thousand " + numberToWords(n%1000))
	case n < 1_000_000_000:
		return strings.TrimSpace(numberToWords(n/1_000_000) + " Million " + numberToWords(n%1_000_000))
	default:
		return strings.TrimSpace(numberToWords(n/1_000_000_000) + " Billion " + numberToWords(n%1_000_000_000))
	}
}

// ComputeDueDateFromTerms calculates due date from payment terms and a base date (YYYY-MM-DD).
// For DAYS: every 30 days counts as one calendar month, with any remainder added as days
// (e.g. 30 days from 2026-08-01 → 2026-09-01).
func ComputeDueDateFromTerms(baseDate string, value int, unit enums.PaymentTermsUnit) (string, error) {
	if value <= 0 || unit == enums.PAYMENT_TERMS_UNDEFINED {
		return "", errs.ErrInvalidPaymentTerms
	}
	if strings.TrimSpace(baseDate) == "" {
		baseDate = NowPH().Format(constants.DateLayoutISO)
	}
	t, err := timeParseDate(baseDate)
	if err != nil {
		return "", err
	}
	switch unit {
	case enums.PAYMENT_TERMS_DAYS:
		t = t.AddDate(0, value/30, value%30)
	case enums.PAYMENT_TERMS_MONTHS:
		t = t.AddDate(0, value, 0)
	case enums.PAYMENT_TERMS_YEARS:
		t = t.AddDate(value, 0, 0)
	default:
		return "", errs.ErrInvalidPaymentTermsUnit
	}
	return t.Format(constants.DateLayoutISO), nil
}

func timeParseDate(s string) (t time.Time, err error) {
	return time.Parse(constants.DateLayoutISO, s)
}

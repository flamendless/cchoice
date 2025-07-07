package enums

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
)

var tblPaymentMethod = map[PaymentMethod]string{
	PAYMENT_METHOD_UNDEFINED:         "UNDEFINED",
	PAYMENT_METHOD_QRPH:              "QRPH",
	PAYMENT_METHOD_BILLEASE:          "BILLEASE",
	PAYMENT_METHOD_CARD:              "CARD",
	PAYMENT_METHOD_DOB:               "DOB",
	PAYMENT_METHOD_DOB_UBP:           "DOB_UBP",
	PAYMENT_METHOD_BRANKAS_BDO:       "BRANKAS_BDO",
	PAYMENT_METHOD_BRANKAS_LANDBANK:  "BRANKAS_LANDBANK",
	PAYMENT_METHOD_BRANKAS_METROBANK: "BRANKAS_METROBANK",
	PAYMENT_METHOD_GCASH:             "GCASH",
	PAYMENT_METHOD_GRAB_PAY:          "GRAB_PAY",
	PAYMENT_METHOD_PAYMAYA:           "PAYMAYA",
}

func TestPaymentMethodToString(t *testing.T) {
	for paymentMethod, val := range tblPaymentMethod {
		t.Run(val, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, val, paymentMethod.String())
		})
	}
}

func TestParsePaymentMethodEnum(t *testing.T) {
	for paymentMethod, val := range tblPaymentMethod {
		t.Run(val, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, paymentMethod, ParsePaymentMethodToEnum(val))
		})
	}
}

func TestPaymentMethodEnumToJSON(t *testing.T) {
	for paymentMethod, val := range tblPaymentMethod {
		t.Run(val, func(t *testing.T) {
			t.Parallel()
			lower, err := json.Marshal(paymentMethod)
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("\"%s\"", strings.ToLower(val)), string(lower))
		})
	}
}

func BenchmarkPaymentMethodToString(b *testing.B) {
	for paymentMethod := range tblPaymentMethod {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_ = paymentMethod.String()
			}
		})
	}
}

func BenchmarkParsePaymentMethodEnum(b *testing.B) {
	for paymentMethod, val := range tblPaymentMethod {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_ = ParsePaymentMethodToEnum(val)
			}
		})
	}
}

func BenchmarkPaymentMethodEnumToJSON(b *testing.B) {
	for paymentMethod := range tblPaymentMethod {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_, _ = json.Marshal(paymentMethod)
			}
		})
	}
}

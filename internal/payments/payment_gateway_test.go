package payments

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var tblPaymentGateway = map[PaymentGateway]string{
	PAYMENT_GATEWAY_PAYMONGO: "PAYMONGO",
}

func TestPaymentGatewayToString(t *testing.T) {
	for paymentMethod, val := range tblPaymentGateway {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, val, paymentMethod.String())
		})
	}
}

func TestParsePaymentGatewayEnum(t *testing.T) {
	for paymentMethod, val := range tblPaymentGateway {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, paymentMethod, ParsePaymentGatewayToEnum(val))
		})
	}
}

func TestParsePaymentGatewayEnumLower(t *testing.T) {
	for paymentMethod, val := range tblPaymentGateway {
		t.Run(val, func(t *testing.T) {
			require.Equal(t, paymentMethod, ParsePaymentGatewayToEnum(strings.ToLower(val)))
		})
	}
}

func BenchmarkPaymentGatewayToString(b *testing.B) {
	for paymentMethod := range tblPaymentGateway {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_ = paymentMethod.String()
			}
		})
	}
}

func BenchmarkParsePaymentGatewayEnum(b *testing.B) {
	for paymentMethod, val := range tblPaymentGateway {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_ = ParsePaymentGatewayToEnum(val)
			}
		})
	}
}

func BenchmarkParsePaymentGatewayEnumLower(b *testing.B) {
	for paymentMethod, val := range tblPaymentGateway {
		b.Run(paymentMethod.String(), func(b *testing.B) {
			for b.Loop() {
				_ = ParsePaymentGatewayToEnum(strings.ToLower(val))
			}
		})
	}
}

package paymongo

import (
	"cchoice/internal/payments"
	"testing"
)

func getPayMongo() PayMongo {
	pm := PayMongo{
		apiKey:         "TEST_API_KEY",
		successURL:     "TEST_SUCCESS_URL",
		cancelURL:      "TEST_CANCEL_URL",
		paymentGateway: payments.PAYMENT_GATEWAY_PAYMONGO,
	}
	return pm
}

func TestGenerateRefNo(t *testing.T) {
	pm := getPayMongo()
	cache := map[string]bool{}
	for range 1_000_000 {
		ref := pm.GenerateRefNo()
		if _, exists := cache[ref]; exists {
			t.Log(ref, len(cache))
			t.FailNow()
		}
		cache[ref] = true
	}
}

func BenchmarkGenerateRefNo(b *testing.B) {
	pm := getPayMongo()
	b.ResetTimer()
	for b.Loop() {
		_ = pm.GenerateRefNo()
	}
}

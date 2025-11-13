package paymongo

import (
	"cchoice/internal/payments"
	"testing"

	"github.com/stretchr/testify/require"
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
	duplicates := 0
	const numGenerations = 1_000_000

	for range numGenerations {
		ref := pm.GenerateRefNo()
		if _, exists := cache[ref]; exists {
			duplicates++
			if duplicates == 1 {
				t.Logf("First duplicate found: %s (after %d generations)", ref, len(cache))
			}
		}
		cache[ref] = true
	}

	require.LessOrEqual(t, duplicates, 100, "too many duplicates: %d out of %d (%.4f%%)", duplicates, numGenerations, float64(duplicates)/float64(numGenerations)*100)
}

func BenchmarkGenerateRefNo(b *testing.B) {
	pm := getPayMongo()
	b.ResetTimer()
	for b.Loop() {
		_ = pm.GenerateRefNo()
	}
}

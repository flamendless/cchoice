package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndVerifyRedirectToken(t *testing.T) {
	t.Parallel()

	secret := "test-webhook-secret"
	paymentRef := "ref-12345"
	token := GenerateRedirectToken(secret, paymentRef)

	assert.True(t, VerifyRedirectToken(secret, paymentRef, token))
	assert.False(t, VerifyRedirectToken(secret, "other-ref", token))
	assert.False(t, VerifyRedirectToken("wrong-secret", paymentRef, token))
	assert.False(t, VerifyRedirectToken(secret, paymentRef, ""))
	assert.False(t, VerifyRedirectToken(secret, paymentRef, "invalid"))
}

func TestVerifyRedirectToken_Expired(t *testing.T) {
	t.Parallel()

	secret := "test-webhook-secret"
	paymentRef := "ref-expired"
	exp := time.Now().Add(-time.Hour).Unix()

	msg := redirectTokenMessage(paymentRef, exp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	token := fmt.Sprintf("%d.%s", exp, sig)

	assert.False(t, VerifyRedirectToken(secret, paymentRef, token))
}

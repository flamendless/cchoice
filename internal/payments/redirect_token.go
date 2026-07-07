package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const RedirectTokenTTL = 24 * time.Hour

func GenerateRedirectToken(secret, paymentRef string) string {
	exp := time.Now().Add(RedirectTokenTTL).Unix()
	msg := redirectTokenMessage(paymentRef, exp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d.%s", exp, sig)
}

func VerifyRedirectToken(secret, paymentRef, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}

	msg := redirectTokenMessage(paymentRef, exp)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expected)) == 1
}

func redirectTokenMessage(paymentRef string, exp int64) string {
	return fmt.Sprintf("%s:%d", paymentRef, exp)
}

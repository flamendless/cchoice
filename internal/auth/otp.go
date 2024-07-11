package auth

import (
	"bytes"
	"image/png"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateOTP(issuer, accountName string) (*otp.Key, []byte, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, nil, err
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, nil, err
	}
	png.Encode(&buf, img)

	return key, buf.Bytes(), nil
}

func ValidateOTP(passcode string, secret string) bool {
	valid := totp.Validate(passcode, secret)
	return valid
}

func GenerateRecoveryCodes() string {
	//TODO: Brandon
	return "test"
}

func GeneratePassCode(secret string) (string, error) {
	passcode, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return "", err
	}
	return passcode, nil
}

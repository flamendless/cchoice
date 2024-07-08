package auth

import (
	"bytes"
	"image/png"

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

func GenerateRecoverCodes() string {
	//TODO: Brandon
	return "test"
}

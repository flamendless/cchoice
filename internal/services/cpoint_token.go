package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"cchoice/internal/errs"
	"cchoice/internal/logs"
)

type CPointTokenService struct {
	secret []byte
}

type CPointTokenPayload struct {
	Code   string `json:"code"`
	UserID string `json:"uid"`
	Exp    int64  `json:"exp"`
}

func NewCPointTokenService(secret string) *CPointTokenService {
	if secret == "" {
		panic("secret is required")
	}
	return &CPointTokenService{
		secret: []byte(secret),
	}
}

func (s *CPointTokenService) Generate(code string, userID string, ttl time.Duration) (string, error) {
	payload := CPointTokenPayload{
		Code:   code,
		UserID: userID,
		Exp:    time.Now().Add(ttl).Unix(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.URLEncoding.EncodeToString(payloadBytes)

	h := hmac.New(sha256.New, s.secret)
	h.Write(payloadBytes)
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return encodedPayload + "." + signature, nil
}

func (s *CPointTokenService) Verify(token string) (*CPointTokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errs.ErrCpointInvalidTokenFormat
	}

	encodedPayload, signature := parts[0], parts[1]

	payloadBytes, err := base64.URLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, errs.ErrCpointInvalidPayloadEncoding
	}

	h := hmac.New(sha256.New, s.secret)
	h.Write(payloadBytes)
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSig)) != 1 {
		return nil, errs.ErrCpointInvalidSignature
	}

	var payload CPointTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, errs.ErrCpointInvalidPayload
	}

	if payload.Code == "" || payload.UserID == "" {
		return nil, errs.ErrCpointMissingRequiredFields
	}

	if payload.Exp < time.Now().Unix() {
		return nil, errs.ErrCpointTokenExpired
	}

	return &payload, nil
}

func (s CPointTokenService) Log() {
	logs.Log().Info("[CPointToken] Loaded")
}

var _ IService = (*CPointTokenService)(nil)

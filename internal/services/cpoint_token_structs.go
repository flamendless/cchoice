package services

type CPointTokenPayload struct {
	Code   string `json:"code"`
	UserID string `json:"uid"`
	Exp    int64  `json:"exp"`
}

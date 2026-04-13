package services

import "cchoice/internal/enums"

type ResetContext struct {
	UserID   int64
	UserType enums.UserType
	Email    string
}

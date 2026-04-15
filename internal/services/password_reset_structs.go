package services

import "cchoice/internal/enums"

type ResetContext struct {
	Email    string
	UserID   int64
	UserType enums.UserType
}

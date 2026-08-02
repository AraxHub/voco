package domain

import "errors"

var (
	ErrRoomNotFound          = errors.New("room not found")
	ErrUserNotFound          = errors.New("user not found")
	ErrForbidden             = errors.New("forbidden")
	ErrConflict              = errors.New("conflict")
	ErrValidation            = errors.New("validation")
	ErrNotFound              = errors.New("not found")
	ErrBlocked               = errors.New("blocked")
	ErrMessageRequestPending = errors.New("message request pending")
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrEventNotFound         = errors.New("event not found")
	ErrNicknameTaken         = errors.New("nickname taken")
)

package dsErrors

import "errors"

var (
	ErrInvalidEmailAddress = errors.New("email is invalid")
	ErrPasswordTooBig      = errors.New("password too big")
	ErrPasswordTooShort    = errors.New("password too short")
)

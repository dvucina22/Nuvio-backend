package models

import "errors"

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrMissingFields         = errors.New("missing required fields")
	ErrPasswordWeak          = errors.New("password does not meet security requirements")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordNoUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoDigit       = errors.New("password must contain at least one number")
	ErrPasswordInvalid       = errors.New("password must be at least 8 characters long, contain one uppercase letter and one number")
	ErrPasswordHashFailed    = errors.New("failed to hash password")
	ErrPasswordCompareFailed = errors.New("failed to compare passwords")
	ErrUserDoesNotHaveRole   = errors.New("user does not have the required role")
	ErrCannotAssignRole      = errors.New("cannot assign this role to user")
	ErrUserAlreadyHasRole    = errors.New("user already has this role")
	ErrCannotRemoveRole      = errors.New("cannot remove this role from user")
)

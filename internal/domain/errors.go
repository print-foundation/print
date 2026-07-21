package domain

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrVerificationFailed = errors.New("verification failed")

	ErrUnsupported = errors.New("unsupported")

	ErrCancelled = errors.New("operation cancelled")

	ErrInsufficientSpace = errors.New("insufficient space")

	ErrConfirmationRequired = errors.New("explicit confirmation required")
)

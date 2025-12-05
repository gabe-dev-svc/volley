package errors

import "errors"

var (
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrNotFound         = errors.New("resource not found")
	ErrMissingAuthToken = errors.New("missing authentication token")
	ErrInvalidAuthToken = errors.New("provided authentication token is invalid")
)

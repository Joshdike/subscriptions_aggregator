package errors

import "errors"

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrAlreadyExists        = errors.New("subscription already exists")
	ErrDecodingJSON         = errors.New("error decoding json")
	ErrEncodingJSON         = errors.New("error encoding json")
)

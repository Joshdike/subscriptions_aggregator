//Package errors contains custom error types for the application
//These errors are used for common error cases to enable consistent error handling and checking.
package errors

import "errors"

var (
	
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrInvalidInput         = errors.New("invalid input") //generic error for invalid or malformed input
	ErrAlreadyExists        = errors.New("subscription already exists") 
	ErrDecodingJSON         = errors.New("error decoding json")
	ErrEncodingJSON         = errors.New("error encoding json")
	ErrUnauthorized         = errors.New("unauthorized")
)

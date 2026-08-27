package calderadb

import (
	"errors"
	"fmt"
)

// Error types
var (
	ErrConnectionFailed   = errors.New("connection failed")
	ErrTimeout            = errors.New("operation timed out")
	ErrKeyNotFound        = errors.New("key not found")
	ErrInvalidResponse    = errors.New("invalid response from server")
	ErrCollectionNotFound = errors.New("collection not found")
	ErrCollectionExists   = errors.New("collection already exists")
	ErrPoolExhausted      = errors.New("connection pool exhausted")
)

// Error represents a CalderaDB error with additional context
type Error struct {
	Op  string // Operation that failed
	Err error  // Underlying error
	Key string // Optional key involved
}

func (e *Error) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("calderadb: %s %s: %v", e.Op, e.Key, e.Err)
	}
	return fmt.Sprintf("calderadb: %s: %v", e.Op, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// Error helper functions
func newError(op string, err error) error {
	return &Error{Op: op, Err: err}
}

func newErrorWithKey(op, key string, err error) error {
	return &Error{Op: op, Key: key, Err: err}
}

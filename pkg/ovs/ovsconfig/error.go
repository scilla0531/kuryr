package ovsconfig

import (
	"strings"
)

type Error interface {
	error
	Timeout() bool   // Is the error a timeout?
	Temporary() bool // Is the error temporary?
}

type TransactionError struct {
	error
	temporary bool
}

func NewTransactionError(err error, temporary bool) *TransactionError {
	return &TransactionError{err, temporary}
}

func (e *TransactionError) Temporary() bool {
	return e.temporary || e.Timeout()
}

func (e *TransactionError) Timeout() bool {
	return strings.HasPrefix(e.Error(), "timed out:")
}

type InvalidArgumentsError string

func newInvalidArgumentsError(err string) InvalidArgumentsError {
	return InvalidArgumentsError(err)
}

func (e InvalidArgumentsError) Error() string {
	return string(e)
}

func (e InvalidArgumentsError) Temporary() bool {
	return false
}

func (e InvalidArgumentsError) Timeout() bool {
	return false
}

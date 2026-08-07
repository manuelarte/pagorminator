package pagegeneric

import (
	"errors"
	"fmt"
)

var (
	// ErrTotalElementsNotSet is an error that is returned when the total elements value is not set.
	ErrTotalElementsNotSet       = errors.New("total elements not set")
	ErrNoNextPage                = errors.New("next page not available")
	_                      error = new(TotalElementsNotValidError)
)

type (
	// TotalElementsNotValidError is an error type that represents an invalid total elements value.
	TotalElementsNotValidError struct {
		TotalElements int64
	}
)

func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.TotalElements)
}

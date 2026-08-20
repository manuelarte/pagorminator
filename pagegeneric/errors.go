package pagegeneric

import (
	"fmt"
)

var _ error = new(TotalElementsNotValidError)

type (
	// TotalElementsNotValidError is an error type that represents an invalid total elements value.
	TotalElementsNotValidError struct {
		TotalElements int64
	}
)

// Error returns the error message.
func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.TotalElements)
}

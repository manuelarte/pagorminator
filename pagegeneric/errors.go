package pagegeneric

import "fmt"

// TotalElementsNotValidError is an error type that represents an invalid total elements value.
type TotalElementsNotValidError struct {
	TotalElements int64
}

func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.TotalElements)
}

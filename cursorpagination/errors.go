package cursorpagination

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrSizeCantBeNegative is returned when the size is negative.
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	// ErrCursorsRequired is returned when the order is required.
	ErrCursorsRequired = errors.New("order is required")
	// ErrOrderNotValid is returned when the order is not valid.
	ErrOrderNotValid       = errors.New("order is not valid")
	_                error = new(CursorValuesNotValidError)
)

// CursorValuesNotValidError is an error type that represents an invalid cursor values.
type CursorValuesNotValidError struct {
	CursorsHaveValues []string
	CursorsNilValue   []string
}

// Error returns the error message.
func (c CursorValuesNotValidError) Error() string {
	return fmt.Sprintf(
		"some cursor have values and some others don't, with:[%v], without[%v]",
		c.CursorsHaveValues,
		c.CursorsNilValue,
	)
}

// Is allows [errors.Is] to compare CursorValuesNotValidError values even though
// the struct contains slice fields (which are not directly comparable).
// It returns true when the target error is a CursorValuesNotValidError (or
// pointer to) and both slices have the same contents in the same order.
func (c CursorValuesNotValidError) Is(target error) bool {
	switch t := target.(type) {
	case CursorValuesNotValidError:
		return reflect.DeepEqual(c.CursorsHaveValues, t.CursorsHaveValues) &&
			reflect.DeepEqual(c.CursorsNilValue, t.CursorsNilValue)
	case *CursorValuesNotValidError:
		return reflect.DeepEqual(c.CursorsHaveValues, t.CursorsHaveValues) &&
			reflect.DeepEqual(c.CursorsNilValue, t.CursorsNilValue)
	default:
		return false
	}
}

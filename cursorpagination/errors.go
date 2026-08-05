package cursorpagination

import (
	"errors"
	"fmt"
)

var (
	ErrSizeCantBeNegative       = errors.New("size can't be negative")
	ErrOrderRequired            = errors.New("order is required")
	ErrOrderNotValid            = errors.New("order is not valid")
	_                     error = new(CursorValuesNotValidError)
)

type CursorValuesNotValidError struct {
	CursorsHaveValues []string
	CursorsNilValue   []string
}

func (c CursorValuesNotValidError) Error() string {
	return fmt.Sprintf(
		"some cursor have values and some others don't, with:[%v], without[%v]",
		c.CursorsHaveValues,
		c.CursorsNilValue,
	)
}

package cursorpagination

import "errors"

var (
	ErrSizeCantBeNegative  = errors.New("size can't be negative")
	ErrOrderRequired       = errors.New("order is required")
	ErrOrderNotValid       = errors.New("order is not valid")
	ErrCursorValueNotValid = errors.New("cursor value is not valid")
)

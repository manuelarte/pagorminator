package pagepagination

import "errors"

var (
	// ErrPageCantBeNegative is an error type that represents an invalid page value.
	ErrPageCantBeNegative = errors.New("page number can't be negative")
	// ErrSizeCantBeNegative is an error type that represents an invalid size value.
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	// ErrSizeNotAllowed is an error type that represents an invalid size value.
	ErrSizeNotAllowed = errors.New("size is not allowed")
)

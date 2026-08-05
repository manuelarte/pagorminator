package page

import "errors"

var (
	ErrPageCantBeNegative = errors.New("page number can't be negative")
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	ErrSizeNotAllowed     = errors.New("size is not allowed")
)

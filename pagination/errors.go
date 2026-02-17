package pagination

import (
	"errors"
	"fmt"
)

var (
	ErrPageCantBeNegative = errors.New("page number can't be negative")
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	ErrSizeNotAllowed     = errors.New("size is not allowed")
)

type TotalElementsNotValidError struct {
	totalElements int64
}

func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.totalElements)
}

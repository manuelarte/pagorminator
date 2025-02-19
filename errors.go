package pagorminator

import (
	"errors"
	"fmt"
)

var (
	ErrPageCantBeNegative   = errors.New("page number can't be negative")
	ErrSizeCantBeNegative   = errors.New("size can't be negative")
	ErrSizeNotAllowed       = errors.New("size is not allowed")
	ErrOrderPropertyIsEmpty = errors.New("order property is empty")
)

type ErrOrderDirectionNotValid struct {
	Direction Direction
}

func (e ErrOrderDirectionNotValid) Error() string {
	return fmt.Sprintf("order direction is not valid: %s", e.Direction)
}

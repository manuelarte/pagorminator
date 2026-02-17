package sort

import (
	"errors"
	"fmt"
)

var ErrOrderPropertyIsEmpty = errors.New("order property is empty")

type OrderDirectionNotValidError struct {
	Direction Direction
}

func (e OrderDirectionNotValidError) Error() string {
	return fmt.Sprintf("order direction is not valid: %s", e.Direction)
}

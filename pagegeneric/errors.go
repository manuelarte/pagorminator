package pagegeneric

import (
	"fmt"
)

type TotalElementsNotValidError struct {
	TotalElements int64
}

func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.TotalElements)
}

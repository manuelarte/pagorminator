package domain

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TotalElementsNotValidError struct {
	TotalElements int64
}

func (e TotalElementsNotValidError) Error() string {
	return fmt.Sprintf("total elements is not valid: %d", e.TotalElements)
}

type (
	PaginationRequest interface {
		GetSize() int
	}

	PaginationResponse interface {
		SetTotalElements(totalElements int64) error
		IsTotalElementsSet() bool
	}

	Pagination interface {
		PaginationRequest
		PaginationResponse
		clause.Expression
		gorm.StatementModifier
	}
)

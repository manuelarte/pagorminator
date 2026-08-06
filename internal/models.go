package internal

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	PaginationRequest interface {
		GetSize() int
	}

	PaginationResponse interface {
		// SetTotalElements sets the total elements.
		//
		// Errors:
		//   - ErrTotalElementsNotValid if the total elements are below zero.
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

package pagorminator

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagepagination"
)

var (
	_ Pagination = new(cursorpagination.Pagination)
	_ Pagination = new(pagepagination.Pagination)
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

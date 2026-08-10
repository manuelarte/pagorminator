package pagorminator

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagegeneric"
	"github.com/manuelarte/pagorminator/pagepagination"
)

var (
	_ Pagination                             = new(cursorpagination.Pagination)
	_ Nextable[*cursorpagination.Pagination] = new(cursorpagination.Pagination)
	_ Pagination                             = new(pagepagination.Pagination)
	_ Nextable[*pagepagination.Pagination]   = new(pagepagination.Pagination)
	_ Prevable[*pagepagination.Pagination]   = new(pagepagination.Pagination)
)

type (
	PaginationRequest interface {
		// GetSize returns the pagination size, a.k.a. limit
		GetSize() int
		// IsUnPaged returns true if the pagination is unpaged, meaning no pagination is applied.
		IsUnPaged() bool
	}

	Nextable[P PaginationRequest] interface {
		// Next returns the next page and whether the next page could be retrieved.
		// If the next page could not be retrieved, the second return value is false.
		// Examples of cases can't be retrieved:
		//   - No next page
		//   - The total elements are not set
		Next() (P, pagegeneric.PrevNextPossible)
	}

	Prevable[P PaginationRequest] interface {
		// Prev returns the previous page and whether the previous page could be retrieved.
		// If the previous page could not be retrieved, the second return value is false.
		// Examples of cases can't be retrieved:
		//   - No previous page
		//   - The total elements are not set
		Prev() (P, pagegeneric.PrevNextPossible)
	}

	PaginationResponse interface {
		// GetTotalElements returns the total elements.
		// It also returns if the total element was set.
		GetTotalElements() (int64, bool)
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

package pagorminator

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagegeneric"
	"github.com/manuelarte/pagorminator/pagepagination"
)

var (
	_ Pagination                           = new(cursorpagination.Pagination)
	_ Nexter[*cursorpagination.Pagination] = new(cursorpagination.Pagination)
	_ Pagination                           = new(pagepagination.Pagination)
	_ Nexter[*pagepagination.Pagination]   = new(pagepagination.Pagination)
	_ Prever[*pagepagination.Pagination]   = new(pagepagination.Pagination)
)

type (
	// PaginationRequest is the interface that contains the information about the pagination.
	PaginationRequest interface {
		// Size returns the pagination size, a.k.a. limit
		Size() int
		// IsUnPaged returns true if the pagination is unpaged, meaning no pagination is applied.
		IsUnPaged() bool
	}

	// Nexter is the interface that indicates that you can ask for the next page.
	Nexter[P PaginationRequest] interface {
		// Next returns the next page and whether the next page could be retrieved.
		// If the next page could not be retrieved, the second return value is false.
		// Examples of cases can't be retrieved:
		//   - No next page
		//   - The total elements are not set
		Next() (P, pagegeneric.PrevNextPossible)
	}

	// Prever is the interface that indicates that you can ask for the previous page.
	Prever[P PaginationRequest] interface {
		// Prev returns the previous page and whether the previous page could be retrieved.
		// If the previous page could not be retrieved, the second return value is false.
		// Examples of cases can't be retrieved:
		//   - No previous page
		//   - The total elements are not set
		Prev() (P, pagegeneric.PrevNextPossible)
	}

	// PaginationCountResponse is the interface that contains the information after a count query.
	PaginationCountResponse interface {
		// GetTotalElements returns the total elements.
		// It also returns if the total element was set.
		TotalElements() (int64, bool)
		// SetTotalElements sets the total elements.
		//
		// Errors:
		//   - ErrTotalElementsNotValid if the total elements are below zero.
		SetTotalElements(totalElements int64) error
		IsTotalElementsSet() bool
	}

	// Pagination is the interface that combines the pagination request and the pagination count response.
	Pagination interface {
		PaginationRequest
		PaginationCountResponse
		clause.Expression
		gorm.StatementModifier
	}
)

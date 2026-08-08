package internal

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	PaginationRequest interface {
		// GetSize returns the pagination size, a.k.a. limit
		GetSize() int
		// IsUnPaged returns true if the pagination is unpaged, meaning no pagination is applied.
		IsUnPaged() bool
	}

	Nextable[P PaginationRequest] interface {
		Next() (P, error)
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

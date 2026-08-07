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

	PaginationResponse interface {
		GetTotalElements() int64
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

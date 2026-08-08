package cursorpagination

import (
	"slices"
	"strings"
	"sync"

	"github.com/manuelarte/pagorminator/internal"
	"github.com/manuelarte/pagorminator/pagegeneric"
)

var _ internal.Pagination = new(Pagination)

type (
	//go:structinit
	Cursor struct {
		column string
		value  any
		order  pagegeneric.Order
	}

	// Pagination Clause to apply cursor pagination.
	//
	//go:structinit
	Pagination struct {
		size    int
		cursors []Cursor

		mu               sync.RWMutex
		totalElements    int64
		totalElementsSet bool

		// latestLen represents the number of rows returned in the latest query using this pagination.
		latestLen int
		// latestCursorValues represents the cursor latest values of the latest query using this pagination.
		latestCursorValues map[string]any
	}
)

// New Create a cursor page given cursor value, size and order.
// It returns the pagination object and any error encountered.
//
// Errors:
//   - ErrSizeCantBeNegative if the size value is below zero.
//   - ErrOrderRequired if the cursors are empty.
//   - ErrOrderNotValid if the order is not Asc or Desc
//   - CursorValuesNotValidError if some cursors have values and some others do not.
func New(size int, cursors ...Cursor) (*Pagination, error) {
	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}

	if size > 0 && len(cursors) == 0 {
		return nil, ErrOrderRequired
	}

	hasValue := make([]string, 0, len(cursors))
	hasNilValue := make([]string, 0, len(cursors))

	for _, cursor := range cursors {
		switch cursor.order.(type) {
		case pagegeneric.Asc, pagegeneric.Desc:
			// valid
		default:
			return nil, ErrOrderNotValid
		}

		if cursor.value == nil {
			hasNilValue = append(hasNilValue, cursor.column)
		} else {
			hasValue = append(hasValue, cursor.column)
		}
	}

	if len(hasValue) > 0 && len(hasNilValue) > 0 {
		return nil, CursorValuesNotValidError{
			CursorsHaveValues: hasValue,
			CursorsNilValue:   hasNilValue,
		}
	}

	return &Pagination{
		size:    size,
		cursors: slices.Clone(cursors),
	}, nil
}

// Must Create a cursor page given cursor value, size and order.
// It returns the pagination object or panic if any error is encountered.
func Must(size int, cursors ...Cursor) *Pagination {
	pagination, err := New(size, cursors...)
	if err != nil {
		panic(err)
	}

	return pagination
}

// UnPaged Create an unpaged request (no pagination is applied).
func UnPaged() *Pagination {
	return &Pagination{size: 0}
}

// GetSize Get the page size.
func (p *Pagination) GetSize() int {
	return p.size
}

// GetCursors Get the cursor values.
func (p *Pagination) GetCursors() []Cursor {
	return slices.Clone(p.cursors)
}

func (p *Pagination) GetTotalElements() (int64, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElements, p.totalElementsSet
}

// SetTotalElements sets the total elements.
// Method to be used by the plugin callbacks.
//
// Errors:
//   - ErrTotalElementsNotValid if the total elements are below zero.
func (p *Pagination) SetTotalElements(totalElements int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if totalElements < 0 {
		return pagegeneric.TotalElementsNotValidError{TotalElements: totalElements}
	}

	p.totalElementsSet = true
	p.totalElements = totalElements

	return nil
}

// IsUnPaged Check whether the pagination is applicable.
func (p *Pagination) IsUnPaged() bool {
	return p.size == 0 && len(p.cursors) == 0
}

func (p *Pagination) IsTotalElementsSet() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElementsSet
}

func (p *Pagination) GetLatestQueryValues() (int, map[string]any) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.latestLen, p.latestCursorValues
}

// SetLatestQueryValues sets the latest query values.
// Method to be used by the plugin callbacks.
func (p *Pagination) SetLatestQueryValues(latestLen int, latestCursorValues map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.latestLen = latestLen
	p.latestCursorValues = latestCursorValues
}

// Next Get the next cursor pagination request.
//
// Errors:
//   - pagegeneric.ErrTotalElementsNotSet if the total elements are not set.
//   - pagegeneric.ErrNoNextPage if there is no next page.
//   - ErrLatestCursorValuesNotSet if the latest values from the query were not set.
func (p *Pagination) Next() (*Pagination, error) {
	if len(p.latestCursorValues) == 0 {
		return nil, ErrLatestCursorValuesNotSet
	}

	if !p.totalElementsSet {
		return nil, pagegeneric.ErrTotalElementsNotSet
	}

	if p.latestLen < p.size {
		return nil, pagegeneric.ErrNoNextPage
	}

	return New(p.size, p.cursors...)
}

func (p *Pagination) sortString() string {
	orderStrings := make([]string, len(p.cursors))
	for i, cursor := range p.cursors {
		orderStrings[i] = cursor.order.GormString()
	}

	return strings.Join(orderStrings, ", ")
}

func (p *Pagination) hasCursorValues() bool {
	return len(p.cursors) > 0 && p.cursors[0].value != nil
}

// NewCursor creates a cursor definition for a column, optional value and sort order.
func NewCursor(column string, value any, order pagegeneric.Order) Cursor {
	return Cursor{column: column, value: value, order: order}
}

// Asc creates an ascending cursor for a column.
func Asc(column string, value any) Cursor {
	return NewCursor(column, value, pagegeneric.Asc(column))
}

// Desc creates a descending cursor for a column.
func Desc(column string, value any) Cursor {
	return NewCursor(column, value, pagegeneric.Desc(column))
}

func (c Cursor) GetColumn() string {
	return c.column
}

func (c Cursor) GetValue() any {
	return c.value
}

// GetOrder returns the order definition of a cursor.
func (c Cursor) GetOrder() pagegeneric.Order {
	return c.order
}

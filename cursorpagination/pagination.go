package cursorpagination

import (
	"slices"
	"sync"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

type (
	//go:structinit
	Cursor struct {
		order pagegeneric.Order
		value any
	}

	// Pagination Clause to apply cursor pagination.
	//
	//go:structinit
	Pagination struct {
		size    int
		cursors []Cursor

		mu               sync.RWMutex
		totalElementsSet bool
		totalElements    int64

		latestCursorValuesSet bool
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
//   - ErrCursorsRequired if the cursors are empty.
//   - ErrOrderNotValid if the order is not Asc or Desc
//   - CursorValuesNotValidError if some cursors have values and some others do not.
func New(size int, cursors ...Cursor) (*Pagination, error) {
	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}

	// if size is not zero, but no cursors are provided, we don't assume to uso the primary key as cursor.
	if size > 0 && len(cursors) == 0 {
		return nil, ErrCursorsRequired
	}

	hasValue := make([]string, 0, len(cursors))
	hasNilValue := make([]string, 0, len(cursors))

	for _, cursor := range cursors {
		switch cursor.order.(type) {
		case pagegeneric.Asc, pagegeneric.Desc:
			column := cursor.GetColumn()
			if cursor.value == nil {
				hasNilValue = append(hasNilValue, column)
			} else {
				hasValue = append(hasValue, column)
			}
		default:
			return nil, ErrOrderNotValid
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
	return &Pagination{}
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

// SetLatestQueryValues sets the latest query values.
// Method to be used by the plugin callbacks.
func (p *Pagination) SetLatestQueryValues(latestLen int, latestCursorValues map[string]any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.latestCursorValuesSet = true
	p.latestLen = latestLen
	p.latestCursorValues = latestCursorValues
}

// Next Get the next cursor pagination request.
//
// Errors:
//   - pagegeneric.PreviousCursorValuesNotSet if the latest values from the query were not set.
//   - pagegeneric.NoNextPage if there is no next page.
func (p *Pagination) Next() (*Pagination, pagegeneric.PrevNextPossible) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.latestCursorValuesSet {
		return nil, pagegeneric.PreviousCursorValuesNotSet
	}

	if p.latestLen < p.size {
		return nil, pagegeneric.NoNextPage
	}

	newCursors := make([]Cursor, len(p.cursors))
	for i, c := range p.cursors {
		newCursors[i] = Cursor{
			order: c.order,
			value: p.latestCursorValues[c.GetColumn()],
		}
	}

	return Must(p.size, newCursors...), true
}

// Asc creates an ascending cursor for a column.
func Asc(column string, value any) Cursor {
	return newCursor(pagegeneric.Asc(column), value)
}

// Desc creates a descending cursor for a column.
func Desc(column string, value any) Cursor {
	return newCursor(pagegeneric.Desc(column), value)
}

func newCursor(order pagegeneric.Order, value any) Cursor {
	return Cursor{order: order, value: value}
}

func (c Cursor) GetColumn() string {
	return c.order.Column()
}

// GetValue returns the cursor value.
func (c Cursor) GetValue() any {
	return c.value
}

// GetOrder returns the order definition of a cursor.
func (c Cursor) GetOrder() pagegeneric.Order {
	return c.order
}

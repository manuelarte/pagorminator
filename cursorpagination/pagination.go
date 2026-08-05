package cursorpagination

import (
	"slices"
	"strings"
	"sync"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

type (
	//go:structinit
	Cursor struct {
		Column string
		Value  any
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
	}
)

// New Create a cursor page given cursor value, size and order.
// It returns the pagination object and any error encountered.
func New(size int, cursors ...Cursor) (*Pagination, error) {
	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}

	if len(cursors) == 0 {
		return nil, ErrOrderRequired
	}

	var (
		hasValue    bool
		hasNilValue bool
	)

	for _, cursor := range cursors {
		switch cursor.order.(type) {
		case pagegeneric.Asc, pagegeneric.Desc:
			// valid
		default:
			return nil, ErrOrderNotValid
		}

		if cursor.Value == nil {
			hasNilValue = true
		} else {
			hasValue = true
		}
	}

	if hasValue && hasNilValue {
		return nil, ErrCursorValueNotValid
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

// GetSize Get the page size.
func (p *Pagination) GetSize() int {
	return p.size
}

// GetCursors Get the cursor values.
func (p *Pagination) GetCursors() []Cursor {
	return slices.Clone(p.cursors)
}

func (p *Pagination) SetTotalElements(totalElements int64) error {
	if totalElements < 0 {
		return pagegeneric.TotalElementsNotValidError{TotalElements: totalElements}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalElementsSet = true
	p.totalElements = totalElements

	return nil
}

func (p *Pagination) GetTotalElements() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElements
}

func (p *Pagination) IsTotalElementsSet() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElementsSet
}

func (p *Pagination) sortString() string {
	orderStrings := make([]string, len(p.cursors))
	for i, cursor := range p.cursors {
		orderStrings[i] = cursor.order.GormString()
	}

	return strings.Join(orderStrings, ", ")
}

func (p *Pagination) hasCursorValues() bool {
	return len(p.cursors) > 0 && p.cursors[0].Value != nil
}

// NewCursor creates a cursor definition for a column, optional value and sort order.
func NewCursor(column string, value any, order pagegeneric.Order) Cursor {
	return Cursor{Column: column, Value: value, order: order}
}

// Asc creates an ascending cursor for a column.
func Asc(column string, value any) Cursor {
	return NewCursor(column, value, pagegeneric.Asc(column))
}

// Desc creates a descending cursor for a column.
func Desc(column string, value any) Cursor {
	return NewCursor(column, value, pagegeneric.Desc(column))
}

// GetOrder returns the order definition of a cursor.
func (c Cursor) GetOrder() pagegeneric.Order {
	return c.order
}

package pagepagination

import (
	"math"
	"slices"
	"sync"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

// Pagination Clause to apply pagination.
//
//go:structinit
type Pagination struct {
	page int
	size int
	sort pagegeneric.Sort

	mu               sync.RWMutex
	totalElements    int64
	totalElementsSet bool
}

// New Create page given page, size, and orders.
// It returns the pagination object and any error encountered.
//
// Errors:
//   - ErrPageCantBeNegative if the page value is below zero.
//   - ErrSizeCantBeNegative if the size value is below zero.
//   - ErrSizeNotAllowed if the size is zero and the page is greater than zero.
func New(page, size int, orders ...pagegeneric.Order) (*Pagination, error) {
	if page < 0 {
		return nil, ErrPageCantBeNegative
	}

	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}

	if page > 0 && size == 0 {
		return nil, ErrSizeNotAllowed
	}

	sort := pagegeneric.NewSort(orders...)

	return &Pagination{page: page, size: size, sort: sort}, nil
}

// Must Create page given page, size, and orders.
// It returns the pagination object or panic if any error is encountered.
func Must(page, size int, orders ...pagegeneric.Order) *Pagination {
	pagination, err := New(page, size, orders...)
	if err != nil {
		panic(err)
	}

	return pagination
}

// UnPaged Create an unpaged request (no pagination is applied).
func UnPaged() *Pagination {
	return &Pagination{page: 0, size: 0}
}

// Page Get the page number.
func (p *Pagination) Page() int {
	return p.page
}

// Size Get the page size.
func (p *Pagination) Size() int {
	return p.size
}

// Sort Get the sort constraints.
func (p *Pagination) Sort() pagegeneric.Sort {
	return slices.Clone(p.sort)
}

// Offset Get the offset.
func (p *Pagination) Offset() int {
	return p.page * p.size
}

// TotalPages Get the total number of pages.
func (p *Pagination) TotalPages() int {
	if p.size > 0 {
		return calculateTotalPages(p.totalElements, p.size)
	}

	return 1
}

// TotalElements returns the total elements.
func (p *Pagination) TotalElements() (int64, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElements, p.totalElementsSet
}

// SetTotalElements sets the total elements.
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
	return p.page == 0 && p.size == 0
}

// IsSort Checks if sorting is also requested.
func (p *Pagination) IsSort() bool {
	return len(p.sort) > 0
}

func (p *Pagination) IsTotalElementsSet() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalElementsSet
}

// Next Get the next page pagination request.
//
// Cases in which the next page could not be retrieved:
//   - pagegeneric.NoTotalElements if the total elements are not set.
//   - pagegeneric.NoNextPage if there is no next page.
func (p *Pagination) Next() (*Pagination, pagegeneric.PrevNextPossible) {
	p.mu.RLock()
	totalElementsSet := p.totalElementsSet
	totalElements := p.totalElements
	p.mu.RUnlock()

	if !totalElementsSet {
		return nil, pagegeneric.NoTotalElements
	}

	totalPages := 1
	if p.size > 0 {
		totalPages = calculateTotalPages(totalElements, p.size)
	}

	nextPage := p.page + 1
	if nextPage >= totalPages {
		return nil, pagegeneric.NoNextPage
	}

	return Must(nextPage, p.size, p.Sort()...), true
}

// Prev Get the previous page pagination request.
//
// Cases in which the previous page could not be retrieved:
//   - pagegeneric.NoTotalElements if the total elements are not set.
//   - pagegeneric.NoPrevPage if there is no next page.
func (p *Pagination) Prev() (*Pagination, pagegeneric.PrevNextPossible) {
	p.mu.RLock()
	totalElementsSet := p.totalElementsSet
	p.mu.RUnlock()

	if !totalElementsSet {
		return nil, pagegeneric.NoTotalElements
	}

	prevPage := p.page - 1
	if prevPage < 0 {
		return nil, pagegeneric.NoPrevPage
	}

	return Must(prevPage, p.size, p.Sort()...), true
}

func calculateTotalPages(totalElements int64, size int) int {
	return int(math.Ceil(float64(totalElements) / float64(size)))
}

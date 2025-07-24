package pagorminator

import (
	"math"
	"sync"
)

// Pagination Clause to apply pagination.
type Pagination struct {
	page             int
	size             int
	sort             Sort
	mu               sync.RWMutex
	totalElementsSet bool
	totalElements    int64
}

// NewPageRequest Create page given page, size and orders.
// It returns the pagination object and any error encountered.
func NewPageRequest(page, size int, orders ...Order) (*Pagination, error) {
	if page < 0 {
		return nil, ErrPageCantBeNegative
	}

	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}

	if page > 0 && size == 0 {
		return nil, ErrSizeNotAllowed
	}

	sort := NewSort(orders...)

	return &Pagination{page: page, size: size, sort: sort}, nil
}

// MustPageRequest Create page given page, size and orders.
// It returns the pagination object or panic if any error is encountered.
func MustPageRequest(page, size int, orders ...Order) *Pagination {
	pagination, err := NewPageRequest(page, size, orders...)
	if err != nil {
		panic(err)
	}
	return pagination
}

// UnPaged Create an unpaged request (no pagination is applied).
func UnPaged() *Pagination {
	return &Pagination{page: 0, size: 0}
}

// GetPage Get the page number.
func (p *Pagination) GetPage() int {
	return p.page
}

// GetSize Get the page size.
func (p *Pagination) GetSize() int {
	return p.size
}

// GetOffset Get the offset.
func (p *Pagination) GetOffset() int {
	return p.page * p.size
}

// GetTotalPages Get the total number of pages.
func (p *Pagination) GetTotalPages() int {
	if p.size > 0 {
		return calculateTotalPages(p.totalElements, p.size)
	}

	return 1
}

// GetTotalElements returns the total elements.
func (p *Pagination) GetTotalElements() int64 {
	return p.totalElements
}

// SetTotalElements manually sets the total elements.
func (p *Pagination) SetTotalElements(totalElements int64) error {
	if totalElements < 0 {
		return TotalElementsNotValidError{totalElements: totalElements}
	}

	p.setTotalElements(totalElements)

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

func (p *Pagination) setTotalElements(totalElements int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalElementsSet = true
	p.totalElements = totalElements
}

func (p *Pagination) isTotalElementsSet() bool {
	return p.totalElementsSet
}

func calculateTotalPages(totalElements int64, size int) int {
	return int(math.Ceil(float64(totalElements) / float64(size)))
}

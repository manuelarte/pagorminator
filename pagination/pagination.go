package pagination

import (
	"math"
	"sync"

	"github.com/manuelarte/pagorminator/pagination/sort"
)

// Pagination Clause to apply pagination.
type (
	Page uint
	Size uint

	Pagination struct {
		page Page
		size Size
		sort sort.Sort

		mu               sync.RWMutex
		totalElementsSet bool
		totalElements    uint64
	}
)

// New pagination given page, size and orders.
// It returns the pagination object and any error encountered.
func New(page Page, size Size, orders ...sort.Order) (*Pagination, error) {
	if page > 0 && size == 0 {
		return nil, ErrSizeNotAllowed
	}

	sorting := sort.New(orders...)

	return &Pagination{page: page, size: size, sort: sorting}, nil
}

// Must Create pagination given page, size and orders.
// It returns the pagination object or panic if any error is encountered.
func Must(page Page, size Size, orders ...sort.Order) *Pagination {
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

// Page Get the pagination number.
func (p *Pagination) Page() Page {
	return p.page
}

// Size Get the pagination size.
func (p *Pagination) Size() Size {
	return p.size
}

// Offset Get the offset.
func (p *Pagination) Offset() uint {
	return uint(p.page) * uint(p.size)
}

// TotalPages Get the total number of pages.
func (p *Pagination) TotalPages() uint {
	if p.size > 0 {
		return calculateTotalPages(p.totalElements, p.size)
	}

	return 1
}

// TotalElements returns the total elements.
func (p *Pagination) TotalElements() uint64 {
	return p.totalElements
}

// SetTotalElements manually sets the total elements.
func (p *Pagination) SetTotalElements(totalElements uint64) {
	p.setTotalElements(totalElements)
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
	return p.totalElementsSet
}

func (p *Pagination) setTotalElements(totalElements uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalElementsSet = true
	p.totalElements = totalElements
}

func calculateTotalPages(totalElements uint64, size Size) uint {
	return uint(math.Ceil(float64(totalElements) / float64(size)))
}

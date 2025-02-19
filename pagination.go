package pagorminator

import (
	"math"
	"sync"
)

// PageRequest Create page to query the database
func PageRequest(page, size int, orders ...Order) (*Pagination, error) {
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

// UnPaged Create an unpaged request (no pagination is applied)
func UnPaged() *Pagination {
	return &Pagination{page: 0, size: 0}
}

// Pagination Clause to apply pagination
type Pagination struct {
	page             int
	size             int
	sort             Sort
	teMutex          sync.RWMutex
	totalElementsSet bool
	totalElements    int64
}

// GetPage Get the page number
func (p *Pagination) GetPage() int {
	return p.page
}

// GetSize Get the page size
func (p *Pagination) GetSize() int {
	return p.size
}

// GetOffset Get the offset
func (p *Pagination) GetOffset() int {
	return p.page * p.size
}

// GetTotalPages Get the total number of pages
func (p *Pagination) GetTotalPages() int {
	if p.size > 0 {
		return calculateTotalPages(p.totalElements, p.size)
	} else {
		return 1
	}
}

func (p *Pagination) setTotalElements(totalElements int64) {
	p.teMutex.Lock()
	defer p.teMutex.Unlock()
	p.totalElementsSet = true
	p.totalElements = totalElements
}

func (p *Pagination) isTotalElementsSet() bool {
	return p.totalElementsSet
}

func (p *Pagination) GetTotalElements() int64 {
	return p.totalElements
}

func (p *Pagination) IsUnPaged() bool {
	return p.page == 0 && p.size == 0
}

func (p *Pagination) IsSort() bool {
	return p.sort != nil && len(p.sort) > 0
}

func calculateTotalPages(totalElements int64, size int) int {
	return int(math.Ceil(float64(totalElements) / float64(size)))
}

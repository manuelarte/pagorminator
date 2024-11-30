package pagorminator

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
)

const pagorminatorClause = "pagorminator:clause"

var (
	ErrPageCantBeNegative = errors.New("page number can't be negative")
	ErrSizeCantBeNegative = errors.New("size can't be negative")
	ErrSizeNotAllowed     = errors.New("size is not allowed")
)

var _ clause.Expression = new(Pagination)
var _ gorm.StatementModifier = new(Pagination)

// PageRequest Create page to query the database
func PageRequest(page, size int) (*Pagination, error) {
	if page < 0 {
		return nil, ErrPageCantBeNegative
	}
	if size < 0 {
		return nil, ErrSizeCantBeNegative
	}
	if page > 0 && size == 0 {
		return nil, ErrSizeNotAllowed
	}
	return &Pagination{page: page, size: size}, nil
}

// UnPaged Create an unpaged request (no pagination is applied)
func UnPaged() *Pagination {
	return &Pagination{page: 0, size: 0}
}

// Pagination Clause to apply pagination
type Pagination struct {
	page          int
	size          int
	totalElements int64
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
	return (p.page - 1) * p.size
}

// GetTotalPages Get the total number of pages
func (p *Pagination) GetTotalPages() int {
	if p.size > 0 {
		return calculateTotalPages(p.totalElements, p.size)
	} else {
		return 1
	}
}

func (p *Pagination) GetTotalElements() int64 {
	return p.totalElements
}

func (p *Pagination) IsUnPaged() bool {
	return p.page == 0 && p.size == 0
}

// ModifyStatement Modify the query clause to apply pagination
func (p *Pagination) ModifyStatement(stm *gorm.Statement) {
	db := stm.DB
	db.Set(pagorminatorClause, p)
	if !p.IsUnPaged() {
		stm.DB.Limit(p.size).Offset(p.GetOffset())
	}
}

// Build N/A for pagination
func (p *Pagination) Build(_ clause.Builder) {
}

func calculateTotalPages(totalElements int64, size int) int {
	return int(math.Ceil(float64(totalElements) / float64(size)))
}

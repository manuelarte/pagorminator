package pagepagination

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

// ModifyStatement Modify the query clause to apply pagination.
func (p *Pagination) ModifyStatement(stm *gorm.Statement) {
	tx := stm.DB
	tx.Set(pagegeneric.PagorminatorClause, p)

	if !p.IsUnPaged() {
		tx = tx.Limit(p.size).Offset(p.Offset())
	}

	if p.IsSort() {
		tx.Order(p.sort.String())
	}
}

// Build N/A for pagination.
func (p *Pagination) Build(_ clause.Builder) {
	// method needed to implement interface [clause.Expression]
}

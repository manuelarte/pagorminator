package pagination

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const PagorminatorClause = "pagorminator:clause"

var (
	_ clause.Expression      = new(Pagination)
	_ gorm.StatementModifier = new(Pagination)
)

// ModifyStatement Modify the query clause to apply pagination.
func (p *Pagination) ModifyStatement(stm *gorm.Statement) {
	db := stm.DB
	db.Set(PagorminatorClause, p)

	tx := stm.DB
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

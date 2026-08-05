package cursorpagination

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/internal"
	"github.com/manuelarte/pagorminator/pagegeneric"
)

// ModifyStatement Modify the query clause to apply cursor pagination.
func (p *Pagination) ModifyStatement(stm *gorm.Statement) {
	tx := stm.DB
	tx.Set(internal.PagorminatorClause, p)

	if p.hasCursorValues() {
		cursorWhereSQL, cursorVars := p.buildCursorWhere()
		tx = tx.Where(cursorWhereSQL, cursorVars...)
	}

	if len(p.cursors) > 0 {
		tx = tx.Order(p.sortString())
	}

	if p.size > 0 {
		tx.Limit(p.size)
	}
}

// Build N/A for pagination.
func (p *Pagination) Build(_ clause.Builder) {
	// method needed to implement interface [clause.Expression]
}

func (p *Pagination) buildCursorWhere() (string, []any) {
	var (
		whereSQL strings.Builder
		vars     = make([]any, 0, len(p.cursors)*len(p.cursors))
	)

	for i := range p.cursors {
		if i > 0 {
			whereSQL.WriteString(" OR ")
		}

		whereSQL.WriteString("(")

		for j := range i {
			if j > 0 {
				whereSQL.WriteString(" AND ")
			}

			whereSQL.WriteString(p.cursors[j].Column)
			whereSQL.WriteString(" = ?")

			vars = append(vars, p.cursors[j].Value)
		}

		if i > 0 {
			whereSQL.WriteString(" AND ")
		}

		whereSQL.WriteString(p.cursors[i].Column)

		switch p.cursors[i].order.(type) {
		case pagegeneric.Asc:
			whereSQL.WriteString(" > ?")
		case pagegeneric.Desc:
			whereSQL.WriteString(" < ?")
		}

		vars = append(vars, p.cursors[i].Value)

		whereSQL.WriteString(")")
	}

	return whereSQL.String(), vars
}

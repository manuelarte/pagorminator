package cursorpagination

import (
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

// ModifyStatement Modify the query clause to apply cursor pagination.
func (p *Pagination) ModifyStatement(stm *gorm.Statement) {
	tx := stm.DB
	tx.Set(pagegeneric.PagorminatorClause, p)

	if p.hasCursorValues() {
		cursorWhereSQL, cursorVars := p.buildCursorWhere()
		tx.Set(pagegeneric.PagorminatorCursorWhereSQL, cursorWhereSQL)
		tx.Set(pagegeneric.PagorminatorCursorWhereVars, cursorVars)
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

			whereSQL.WriteString(p.cursors[j].GetColumn())
			whereSQL.WriteString(" = ?")

			vars = append(vars, p.cursors[j].value)
		}

		if i > 0 {
			whereSQL.WriteString(" AND ")
		}

		whereSQL.WriteString(p.cursors[i].GetColumn())

		switch p.cursors[i].order.(type) {
		case pagegeneric.Asc:
			whereSQL.WriteString(" > ?")
		case pagegeneric.Desc:
			whereSQL.WriteString(" < ?")
		}

		vars = append(vars, p.cursors[i].value)

		whereSQL.WriteString(")")
	}

	return whereSQL.String(), vars
}

func (p *Pagination) hasCursorValues() bool {
	return len(p.cursors) > 0 && p.cursors[0].value != nil
}

func (p *Pagination) sortString() string {
	orderStrings := make([]string, len(p.cursors))
	for i, cursor := range p.cursors {
		orderStrings[i] = cursor.order.GormString()
	}

	return strings.Join(orderStrings, ", ")
}

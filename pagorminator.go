package pagorminator

import (
	"gorm.io/gorm"
)

const (
	countKey = "pagorminator.count"
)

var _ gorm.Plugin = new(PaGormMinator)

// PaGormMinator Gorm plugin to add total elements and total pages to your pagination query
type PaGormMinator struct {
}

func (p PaGormMinator) Name() string {
	return "pagorminator"
}

func (p PaGormMinator) Initialize(db *gorm.DB) error {
	err := db.Callback().Query().Before("gorm:query").Register("pagorminator:count", p.count)
	if err != nil {
		return err
	}
	return nil
}

func (p PaGormMinator) count(db *gorm.DB) {
	if db.Statement.Schema != nil {
		if pageable, ok := p.getPageRequest(db); ok {
			if value, ok := db.Get(countKey); !ok || !value.(bool) {
				newDb := db.Session(&gorm.Session{NewDB: true})
				newDb.Statement = db.Statement.Statement

				var totalElements int64
				tx := newDb.Set(countKey, true).Model(newDb.Statement.Model)
				if whereClause, existWhere := db.Statement.Clauses["WHERE"]; existWhere {
					tx.Where(whereClause.Expression)
				}
				tx.Count(&totalElements)
				if tx.Error != nil {
					_ = db.AddError(tx.Error)
				} else {
					pageable.totalElements = totalElements
				}
			}
		}

	}
}

func (p PaGormMinator) getPageRequest(db *gorm.DB) (*Pagination, bool) {
	if value, ok := db.Get(pagorminatorClause); ok {
		if paginationClause, ok := value.(*Pagination); ok {
			return paginationClause, true
		}
	}
	return nil, false
}

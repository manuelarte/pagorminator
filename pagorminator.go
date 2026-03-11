package pagorminator

import (
	"gorm.io/gorm"
)

const (
	countKey = "pagorminator.count"
)

var _ gorm.Plugin = new(PaGorminator)

// PaGorminator Gorm plugin to add total elements and total pages to your pagination query.
type PaGorminator struct {
	Debug bool
}

func (p PaGorminator) Name() string {
	return "pagorminator"
}

func (p PaGorminator) Initialize(db *gorm.DB) error {
	err := db.Callback().Query().Before("gorm:query").Register("pagorminator:count", p.count)
	if err != nil {
		return err
	}

	return nil
}

func (p PaGorminator) count(db *gorm.DB) {
	if db.Statement.Schema == nil && db.Statement.Table == "" {
		return
	}

	if pageable, ok := p.getPageRequest(db); ok && !pageable.isTotalElementsSet() {
		tx := db.Session(&gorm.Session{Context: db.Statement.Context})
		if p.Debug {
			tx = tx.Debug()
		}

		delete(tx.Statement.Clauses, "LIMIT")
		delete(tx.Statement.Clauses, "OFFSET")

		var totalElements int64

		tx = tx.Set(countKey, true)
		tx.Count(&totalElements)

		if tx.Error != nil {
			_ = db.AddError(tx.Error)
			return
		}

		pageable.setTotalElements(totalElements)
	}
}

func (p PaGorminator) getPageRequest(db *gorm.DB) (*Pagination, bool) {
	if value, ok := db.Get(pagorminatorClause); ok { //nolint:nestif // checking many fields in an if way
		if paginationClause, okP := value.(*Pagination); okP {
			if countValue, okCount := db.Get(countKey); !okCount {
				if isCount, hasCount := countValue.(bool); !hasCount || !isCount {
					return paginationClause, true
				}
			}
		}
	}

	return nil, false
}

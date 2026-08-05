package pagorminator

import (
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

const (
	countKey = "pagorminator.count"
)

var _ gorm.Plugin = new(PaGorminator)

// PaGorminator Gorm plugin to add pagination information to your pagination query.
type PaGorminator struct {
	Debug bool
}

func (p PaGorminator) Name() string {
	return "pagorminator"
}

// Initialize initializes the plugin and registers the callback for counting total elements.
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

	if pageable, ok := p.getPageRequest(db); ok && !pageable.IsTotalElementsSet() {
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

		_ = pageable.SetTotalElements(totalElements)
	}
}

func (p PaGorminator) getPageRequest(db *gorm.DB) (PaginationResponse, bool) {
	value, hasPagorminatorClause := db.Get(pagegeneric.PagorminatorClause)
	if !hasPagorminatorClause {
		return nil, false
	}

	paginationClause, okP := value.(PaginationResponse)
	if !okP {
		return nil, false
	}

	countValue, okCount := db.Get(countKey)
	if okCount {
		return nil, false
	}

	isCount, hasCount := countValue.(bool)
	if hasCount || isCount {
		return nil, false
	}

	return paginationClause, true
}

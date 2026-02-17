package pagorminator

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/pagination"
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
	//nolint: nestif // not so complex
	if pageable, ok := p.getPageRequest(db); ok && !pageable.IsTotalElementsSet() {
		if p.Debug {
			db.Debug()
		}

		newDB := db.Session(&gorm.Session{NewDB: true})
		newDB.Statement = db.Statement.Statement

		var totalElements int64

		tx := newDB.Set(countKey, true)
		if db.Statement.Schema != nil {
			tx.Model(newDB.Statement.Model)
		} else if db.Statement.Table != "" {
			tx.Table(db.Statement.Table)
		}

		if db.Statement.Distinct {
			tx.Distinct(db.Statement.Selects)
		}

		for _, join := range db.Statement.Joins {
			args := join.Conds
			//nolint:exhaustive // other cases not supported
			switch join.JoinType {
			case clause.InnerJoin:
				tx.InnerJoins(join.Name, args...)
			case clause.LeftJoin:
				tx.Joins(join.Name, args...)
			default:
				continue
			}
		}

		if whereClause, existWhere := db.Statement.Clauses["WHERE"]; existWhere {
			tx.Where(whereClause.Expression)
		}

		tx.Count(&totalElements)

		if tx.Error != nil {
			_ = db.AddError(tx.Error)
			return
		}

		// #nosec G115 // ignoring since the value needs to be positive
		pageable.SetTotalElements(uint64(totalElements))
	}
}

func (p PaGorminator) getPageRequest(db *gorm.DB) (*pagination.Pagination, bool) {
	value, ok := db.Get(pagination.PagorminatorClause)
	if !ok {
		return nil, false
	}

	paginationClause, okP := value.(*pagination.Pagination)
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

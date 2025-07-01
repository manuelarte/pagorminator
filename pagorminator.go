package pagorminator

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	countKey = "pagorminator.count"
)

var _ gorm.Plugin = new(PaGormMinator)

// PaGormMinator Gorm plugin to add total elements and total pages to your pagination query.
type PaGormMinator struct {
	Debug bool
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

//nolint:gocognit // many ifs to check conditions
func (p PaGormMinator) count(db *gorm.DB) {
	if db.Statement.Schema == nil && db.Statement.Table == "" {
		return
	}
	//nolint: nestif // not so complex
	if pageable, ok := p.getPageRequest(db); ok && !pageable.isTotalElementsSet() {
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

		//nolint:asasalint // it is working
		for _, join := range db.Statement.Joins {
			args := join.Conds
			//nolint:exhaustive // other cases not supported
			switch join.JoinType {
			case clause.InnerJoin:
				tx.InnerJoins(join.Name, args)
			case clause.LeftJoin:
				tx.Joins(join.Name, args)
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
		} else {
			pageable.setTotalElements(totalElements)
		}
	}
}

func (p PaGormMinator) getPageRequest(db *gorm.DB) (*Pagination, bool) {
	if value, ok := db.Get(_pagorminatorClause); ok { //nolint:nestif // checking many fields in an if way
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

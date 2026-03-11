package pagorminator

import (
	"maps"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if pageable, ok := p.getPageRequest(db); ok && !pageable.isTotalElementsSet() {
		newDB := db.Session(&gorm.Session{NewDB: true, Context: db.Statement.Context})
		if p.Debug {
			newDB = newDB.Debug()
		}

		newDB.Statement = clone(db.Statement.Statement)
		delete(newDB.Statement.Clauses, "LIMIT")
		delete(newDB.Statement.Clauses, "OFFSET")

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

// clone almost identically copied from gorm (only removing copying the scopes because they are not exported).
func clone(stmt *gorm.Statement) *gorm.Statement {
	newStmt := &gorm.Statement{
		TableExpr:            stmt.TableExpr,
		Table:                stmt.Table,
		Model:                stmt.Model,
		Unscoped:             stmt.Unscoped,
		Dest:                 stmt.Dest,
		ReflectValue:         stmt.ReflectValue,
		Clauses:              map[string]clause.Clause{},
		Distinct:             stmt.Distinct,
		Selects:              stmt.Selects,
		Omits:                stmt.Omits,
		ColumnMapping:        stmt.ColumnMapping,
		Preloads:             map[string][]any{},
		ConnPool:             stmt.ConnPool,
		Schema:               stmt.Schema,
		Context:              stmt.Context,
		RaiseErrorOnNotFound: stmt.RaiseErrorOnNotFound,
		SkipHooks:            stmt.SkipHooks,
		Result:               stmt.Result,
	}

	if stmt.SQL.Len() > 0 {
		newStmt.SQL.WriteString(stmt.SQL.String())
		newStmt.Vars = make([]any, 0, len(stmt.Vars))
		newStmt.Vars = append(newStmt.Vars, stmt.Vars...)
	}

	maps.Copy(newStmt.Clauses, stmt.Clauses)

	maps.Copy(newStmt.Preloads, stmt.Preloads)

	if len(stmt.Joins) > 0 {
		newStmt.Joins = stmt.Joins
	}

	stmt.Settings.Range(func(k, v any) bool {
		newStmt.Settings.Store(k, v)
		return true
	})

	return newStmt
}

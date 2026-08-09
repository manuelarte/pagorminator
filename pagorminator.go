package pagorminator

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/internal"
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
	if err := db.Callback().Query().
		Before("gorm:query").Register("pagorminator:count", p.count); err != nil {
		return fmt.Errorf("failed to register count callback: %w", err)
	}

	if err := db.Callback().Query().
		Before("gorm:after_query").Register("pagorminator:cursor:next", p.cursorNext); err != nil {
		return fmt.Errorf("failed to register cursor callback: %w", err)
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
		p.removeCursorWhereClause(tx)

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

func (p PaGorminator) removeCursorWhereClause(tx *gorm.DB) {
	cursorWhereSQLRaw, hasCursorWhereSQL := tx.Get(pagegeneric.PagorminatorCursorWhereSQL)
	if !hasCursorWhereSQL {
		return
	}

	cursorWhereSQL, ok := cursorWhereSQLRaw.(string)
	if !ok || cursorWhereSQL == "" {
		return
	}

	cursorWhereVarsRaw, hasCursorWhereVars := tx.Get(pagegeneric.PagorminatorCursorWhereVars)
	if !hasCursorWhereVars {
		return
	}

	cursorWhereVars, ok := cursorWhereVarsRaw.([]any)
	if !ok {
		return
	}

	whereClause, hasWhere := tx.Statement.Clauses["WHERE"]
	if !hasWhere {
		return
	}

	where, ok := whereClause.Expression.(clause.Where)
	if !ok {
		return
	}

	filteredExpressions, removed := removeCursorWhereExpr(where.Exprs, cursorWhereSQL, cursorWhereVars)
	if !removed {
		return
	}

	if len(filteredExpressions) == 0 {
		delete(tx.Statement.Clauses, "WHERE")
		return
	}

	where.Exprs = filteredExpressions
	whereClause.Expression = where
	tx.Statement.Clauses["WHERE"] = whereClause
}

func removeCursorWhereExpr(
	expressions []clause.Expression,
	cursorWhereSQL string,
	cursorWhereVars []any,
) ([]clause.Expression, bool) {
	filteredExpressions := make([]clause.Expression, 0, len(expressions))
	removedAny := false

	for _, expression := range expressions {
		filteredExpression, removed, keep := removeExpression(expression, cursorWhereSQL, cursorWhereVars)
		if removed {
			removedAny = true
		}

		if keep {
			filteredExpressions = append(filteredExpressions, filteredExpression)
		}
	}

	return filteredExpressions, removedAny
}

func removeExpression(
	expression clause.Expression,
	cursorWhereSQL string,
	cursorWhereVars []any,
) (clause.Expression, bool, bool) {
	switch expressionTyped := expression.(type) {
	case clause.Expr:
		if expressionTyped.SQL == cursorWhereSQL && reflect.DeepEqual(expressionTyped.Vars, cursorWhereVars) {
			return nil, true, false
		}
	case clause.AndConditions:
		filteredExpressions, removed := removeCursorWhereExpr(expressionTyped.Exprs, cursorWhereSQL, cursorWhereVars)
		if !removed {
			return expressionTyped, false, true
		}

		if len(filteredExpressions) == 0 {
			return nil, true, false
		}

		expressionTyped.Exprs = filteredExpressions

		return expressionTyped, true, true
	case clause.OrConditions:
		filteredExpressions, removed := removeCursorWhereExpr(expressionTyped.Exprs, cursorWhereSQL, cursorWhereVars)
		if !removed {
			return expressionTyped, false, true
		}

		if len(filteredExpressions) == 0 {
			return nil, true, false
		}

		expressionTyped.Exprs = filteredExpressions

		return expressionTyped, true, true
	case clause.NotConditions:
		filteredExpressions, removed := removeCursorWhereExpr(expressionTyped.Exprs, cursorWhereSQL, cursorWhereVars)
		if !removed {
			return expressionTyped, false, true
		}

		if len(filteredExpressions) == 0 {
			return nil, true, false
		}

		expressionTyped.Exprs = filteredExpressions

		return expressionTyped, true, true
	}

	return expression, false, true
}

func (p PaGorminator) cursorNext(db *gorm.DB) {
	if db.Error != nil || db.Statement.Dest == nil || db.Statement.Schema == nil {
		return
	}

	pagination, hasPagination := p.getPageRequest(db)
	if !hasPagination {
		return
	}

	cursorPagination, ok := pagination.(*cursorpagination.Pagination)
	if !ok {
		return
	}

	schema := db.Statement.Schema
	dest := db.Statement.Dest
	destValue := reflect.ValueOf(dest)

	if destValue.Kind() == reflect.Pointer {
		destValue = destValue.Elem()
	} else {
		return
	}

	if destValue.Kind() != reflect.Slice {
		return
	}

	latestLen := destValue.Len()
	if latestLen == 0 {
		cursorPagination.SetLatestQueryValues(latestLen, nil)
		return
	}

	destValue = destValue.Index(latestLen - 1)
	if destValue.Kind() == reflect.Pointer {
		destValue = destValue.Elem()
	} else {
		return
	}

	latestValues := make(map[string]any, len(cursorPagination.GetCursors()))
	for _, colName := range getCursorColumns(cursorPagination.GetCursors()) {
		if field := schema.LookUpField(colName); field != nil {
			fieldValue, _ := field.ValueOf(db.Statement.Context, destValue)
			latestValues[colName] = fieldValue
		} else {
			return
		}
	}

	cursorPagination.SetLatestQueryValues(latestLen, latestValues)
}

func (p PaGorminator) getPageRequest(db *gorm.DB) (internal.PaginationResponse, bool) {
	value, hasPagorminatorClause := db.Get(pagegeneric.PagorminatorClause)
	if !hasPagorminatorClause {
		return nil, false
	}

	paginationClause, okP := value.(internal.PaginationResponse)
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

func getCursorColumns(cursors []cursorpagination.Cursor) []string {
	columns := make([]string, len(cursors))
	for i, cursor := range cursors {
		columns[i] = cursor.GetColumn()
	}

	return columns
}

package pagorminator

import (
	"github.com/manuelarte/pagorminator/internal"
	"gorm.io/gorm"
)

const (
	countKey = "pagorminator.count"
)

func WithPagination(pageRequest PageRequest) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set("pagorminator:pageRequest", pageRequest)
	}
}

var _ gorm.Plugin = new(PaGormMinator)

// PaGormMinator Gorm plugin to add pagination to your queries
type PaGormMinator struct {
}

func (p PaGormMinator) Name() string {
	return "pagorminator"
}

func (p PaGormMinator) Initialize(db *gorm.DB) error {
	err := db.Callback().Query().Before("gorm:query").Register("pagorminator:addPagination", p.addPagination)
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("pagorminator:addPagination").Register("pagorminator:count", p.count)
	if err != nil {
		return err
	}
	return nil
}

func (p PaGormMinator) addPagination(db *gorm.DB) {
	if db.Statement.Schema != nil {
		if pageRequest, ok := p.getPageRequest(db); ok {
			if !pageRequest.IsUnPaged() {
				db.Limit(pageRequest.GetSize()).Offset(pageRequest.GetOffset())
			}
		}

	}
}

func (p PaGormMinator) count(db *gorm.DB) {
	if db.Statement.Schema != nil {
		if pageRequest, ok := p.getPageRequest(db); ok {
			if value, ok := db.Get(countKey); !ok || !value.(bool) {
				casted, _ := pageRequest.(*internal.PageRequestImpl)

				newDb := db.Session(&gorm.Session{NewDB: true})
				newDb.Statement = db.Statement.Statement

				var totalElements int64
				tx := newDb.Debug().Set(countKey, true).
					Model(newDb.Statement.Model)
				if whereClause, existWhere := db.Statement.Clauses["WHERE"]; existWhere {
					tx.Where(whereClause.Expression)
				}
				tx.Count(&totalElements)
				if tx.Error != nil {
					db.AddError(tx.Error)
				} else {
					casted.TotalElements = int(totalElements)
					if casted.IsUnPaged() {
						casted.Page = 0
						casted.TotalPages = 1
					} else {
						casted.TotalPages = int(totalElements) / casted.Size
					}
				}
			}
		}

	}
}

func (p PaGormMinator) getPageRequest(db *gorm.DB) (PageRequest, bool) {
	if value, ok := db.Get("pagorminator:pageRequest"); ok {
		if pageRequest, ok := value.(PageRequest); ok {
			return pageRequest, true
		}
	}
	return nil, false
}

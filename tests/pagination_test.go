package tests

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/manuelarte/pagorminator/cursorpagination"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSQLIsPortableAcrossDialects(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		dialector func(*sql.DB) gorm.Dialector
	}{
		"sqlite": {
			dialector: func(_ *sql.DB) gorm.Dialector {
				return sqlite.Open("file::memory:?cache=shared")
			},
		},
		"mysql": {
			dialector: func(db *sql.DB) gorm.Dialector {
				return mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true})
			},
		},
		"postgres": {
			dialector: func(db *sql.DB) gorm.Dialector {
				return postgres.New(postgres.Config{Conn: db, PreferSimpleProtocol: true})
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				sqlDB *sql.DB
				err   error
			)

			if name != "sqlite" {
				sqlDB, _, err = sqlmock.New()
				if err != nil {
					t.Fatalf("failed creating sqlmock db: %v", err)
				}
				defer func() { _ = sqlDB.Close() }()
			}

			db, err := gorm.Open(test.dialector(sqlDB), &gorm.Config{DryRun: true})
			if err != nil {
				t.Fatalf("failed opening db: %v", err)
			}

			pageRequest := cursorpagination.Must(5, cursorpagination.Asc("code", "A"), cursorpagination.Desc("price", 10))
			tx := db.Clauses(pageRequest).Model(&TestStruct{}).Find(&[]TestStruct{})
			if tx.Error != nil {
				t.Fatalf("unexpected query error: %v", tx.Error)
			}

			sqlString := tx.Statement.SQL.String()
			if !strings.Contains(sqlString, "ORDER BY") ||
				!strings.Contains(sqlString, "LIMIT") ||
				!strings.Contains(sqlString, " OR ") ||
				!strings.Contains(sqlString, " AND ") ||
				!strings.Contains(sqlString, "code") ||
				!strings.Contains(sqlString, "price") {
				t.Fatalf("unexpected generated SQL: %q", sqlString)
			}
		})
	}
}

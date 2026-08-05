package tests

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/mysql"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator"
	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagegeneric"
	"github.com/manuelarte/pagorminator/pagepagination"
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
				return postgresdriver.New(postgresdriver.Config{Conn: db, PreferSimpleProtocol: true})
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

// TODO(manuelarte): Add more tests to cover different engines.

func TestSimplePagination(t *testing.T) {
	t.Parallel()

	testData := func() []*TestStruct {
		return []*TestStruct{
			{Code: "A", Price: 2},
			{Code: "B", Price: 1},
			{Code: "C", Price: 3},
			{Code: "D", Price: 1},
		}
	}

	wantPage0 := []*TestStruct{
		{Code: "A", Price: 2},
		{Code: "B", Price: 1},
	}
	wantPage1 := []*TestStruct{
		{Code: "C", Price: 3},
		{Code: "D", Price: 1},
	}

	tests := map[string]struct {
		dbFunc func(context.Context) (*gorm.DB, func())
	}{
		"postgres:18-alpine": {
			dbFunc: func(ctx context.Context) (*gorm.DB, func()) {
				dbName := "users"
				dbUser := "user"
				dbPassword := "password"
				postgresContainer, err := postgres.Run(ctx,
					"postgres:18-alpine",
					postgres.WithDatabase(dbName),
					postgres.WithUsername(dbUser),
					postgres.WithPassword(dbPassword),
					postgres.BasicWaitStrategies(),
				)
				if err != nil {
					t.Fatalf("failed to start container: %s", err)
				}
				deferFunc := func() {
					if errTerminating := testcontainers.TerminateContainer(postgresContainer); errTerminating != nil {
						t.Logf("failed to terminate container: %s", err)
					}
				}
				dsn := postgresContainer.MustConnectionString(ctx)
				db, err := gorm.Open(postgresdriver.Open(dsn), &gorm.Config{})
				if err != nil {
					t.Fatalf("failed to open db: %s", err)
				}

				return db, deferFunc
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, deferFunc := test.dbFunc(t.Context())
			defer deferFunc()
			if err := db.Use(pagorminator.PaGorminator{}); err != nil {
				t.Fatalf("failed to use paginator: %v", err)
			}
			if err := db.AutoMigrate(&TestStruct{}); err != nil {
				t.Fatalf("failed to migrate db: %v", err)
			}
			if err := db.CreateInBatches(testData(), 2).Error; err != nil {
				t.Fatalf("failed to create test data: %v", err)
			}

			compareFunc := func(want, got []*TestStruct) {
				if diff := cmp.Diff(
					want,
					got,
					cmpopts.IgnoreFields(TestStruct{}, "ID", "CreatedAt", "UpdatedAt"),
				); diff != "" {
					t.Errorf("diff (-want +got):\n%s", diff)
				}
			}

			// 1st page pagination
			var gotPage0, gotPage1 []*TestStruct

			pageRequest0 := pagepagination.Must(0, 2, pagegeneric.Asc("code"))
			if err := db.Clauses(pageRequest0).Find(&gotPage0).Error; err != nil {
				t.Fatalf("failed to query page 0: %v", err)
			}
			compareFunc(wantPage0, gotPage0)

			pageRequest1 := pagepagination.Must(1, 2, pagegeneric.Asc("code"))
			if err := db.Clauses(pageRequest1).Find(&gotPage1).Error; err != nil {
				t.Fatalf("failed to query page 1: %v", err)
			}
			compareFunc(wantPage1, gotPage1)
			if pageRequest0.GetTotalElements() != 4 || pageRequest0.GetTotalElements() != pageRequest1.GetTotalElements() {
				t.Errorf("unexpected total elements: page0=%d, page1=%d, want=%d",
					pageRequest0.GetTotalElements(),
					pageRequest1.GetTotalElements(),
					4,
				)
			}

			// 2nd cursor pagination
			var gotCursor0, gotCursor1 []*TestStruct

			cursorRequest0 := cursorpagination.Must(2, cursorpagination.Asc("code", nil))
			if err := db.Clauses(cursorRequest0).Find(&gotCursor0).Error; err != nil {
				t.Fatalf("failed to query cursor 0: %v", err)
			}
			compareFunc(wantPage0, gotCursor0)

			cursorRequest1 := cursorpagination.Must(2, cursorpagination.Asc("code", gotCursor0[1].Code))
			if err := db.Clauses(cursorRequest1).Find(&gotCursor1).Error; err != nil {
				t.Fatalf("failed to query cursor 1: %v", err)
			}
			compareFunc(wantPage1, gotCursor1)
		})
	}
}

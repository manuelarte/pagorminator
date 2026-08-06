package tests

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	mysqlDriver "gorm.io/driver/mysql"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator"
	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagegeneric"
	"github.com/manuelarte/pagorminator/pagepagination"
)

type createGormDBFunc func(ctx context.Context) (*gorm.DB, func(), error)

var (
	postgres18 = func(ctx context.Context) (*gorm.DB, func(), error) {
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
			return nil, nil, fmt.Errorf("failed to start container: %w", err)
		}
		deferFunc := func() {
			_ = testcontainers.TerminateContainer(postgresContainer)
		}
		dsn := postgresContainer.MustConnectionString(ctx)
		db, err := gorm.Open(postgresdriver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open db: %w", err)
		}

		return db, deferFunc, nil
	}
	mysql8 = func(ctx context.Context) (*gorm.DB, func(), error) {
		dbName := "users"
		dbUser := "user"
		dbPassword := "password"
		mysqlContainer, err := mysql.Run(ctx,
			"mysql:8.0.36",
			mysql.WithDatabase(dbName),
			mysql.WithUsername(dbUser),
			mysql.WithPassword(dbPassword),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to start container: %w", err)
		}
		deferFunc := func() {
			_ = testcontainers.TerminateContainer(mysqlContainer)
		}
		dsn := mysqlContainer.MustConnectionString(ctx)
		dsn = fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", dsn)
		db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open db: %w", err)
		}

		return db, deferFunc, nil
	}

	dbEngines = map[string]createGormDBFunc{
		"postgres:18-alpine": postgres18,
		"mysql:8.0.36":       mysql8,
	}
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
				return mysqlDriver.New(mysqlDriver.Config{Conn: db, SkipInitializeWithVersion: true})
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

	for engine, dbFunc := range dbEngines {
		t.Run(engine, func(t *testing.T) {
			t.Parallel()

			db, deferFunc, errStartingContainer := dbFunc(t.Context())
			if errStartingContainer != nil {
				t.Fatalf("failed to start container: %v", errStartingContainer)
			}
			defer deferFunc()
			setupDB(t, db)
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

// TODO: add more tests, including SQL injection tests.
func TestCursorPaginationSQLInjection(t *testing.T) {
	t.Parallel()

	testData := func() []*TestStruct {
		return []*TestStruct{
			{Code: "A", Price: 2},
			{Code: "B", Price: 1},
			{Code: "C", Price: 3},
			{Code: "D", Price: 1},
		}
	}

	for engine, dbFunc := range dbEngines {
		t.Run(engine, func(t *testing.T) {
			t.Parallel()

			tests := map[string]struct {
				cursors []cursorpagination.Cursor
			}{
				"delete inside value": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); DELETE FROM test_structs; --"),
					},
				},
			}
			for name, test := range tests {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					db, deferFunc, errStartingContainer := dbFunc(t.Context())
					if errStartingContainer != nil {
						t.Fatalf("failed to start container %s: %v", engine, errStartingContainer)
					}
					defer deferFunc()
					setupDB(t, db)
					if err := db.CreateInBatches(testData(), 2).Error; err != nil {
						t.Fatalf("failed to create test data: %v", err)
					}

					cursorRequest := cursorpagination.Must(2, test.cursors...)
					_ = db.Clauses(cursorRequest).Find(&[]*TestStruct{})

					// TODO(manuelarte): not only do count, but check that the values are all the same
					var got int64
					if err := db.Model(&TestStruct{}).Count(&got).Error; err != nil {
						t.Fatalf("failed to count records: %v", err)
					}

					if got != 4 {
						t.Errorf("unexpected count of records: got=%d, want=%d", got, 4)
					}
				})
			}
		})
	}
}

func setupDB(t *testing.T, db *gorm.DB) {
	if err := db.Use(pagorminator.PaGorminator{Debug: true}); err != nil {
		t.Fatalf("failed to use paginator: %v", err)
	}
	if err := db.AutoMigrate(&TestStruct{}); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}
}

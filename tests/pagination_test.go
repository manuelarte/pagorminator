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
	"gorm.io/gorm/clause"

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

	for engine, dbFunc := range dbEngines {
		t.Run(engine, func(t *testing.T) {
			t.Parallel()

			tests := map[string]struct {
				pageRequests   []clause.Expression
				cursorRequests []clause.Expression
				want           [][]*TestStruct
			}{
				"unpaged": {
					pageRequests: []clause.Expression{
						pagepagination.UnPaged(),
					},
					cursorRequests: []clause.Expression{
						cursorpagination.UnPaged(),
					},
					want: [][]*TestStruct{
						testData(),
					},
				},
				"simple pagination": {
					pageRequests: []clause.Expression{
						pagepagination.Must(0, 2, pagegeneric.Asc("code")),
						pagepagination.Must(1, 2, pagegeneric.Asc("code")),
					},
					cursorRequests: []clause.Expression{
						cursorpagination.Must(2, cursorpagination.Asc("code", nil)),
						cursorpagination.Must(2, cursorpagination.Asc("code", "B")),
					},
					want: [][]*TestStruct{
						{
							{Code: "A", Price: 2},
							{Code: "B", Price: 1},
						},
						{
							{Code: "C", Price: 3},
							{Code: "D", Price: 1},
						},
					},
				},
				"Paged 1/2 items, sort by id asc": {
					pageRequests: []clause.Expression{
						pagepagination.Must(0, 2, pagegeneric.Asc("id")),
					},
					cursorRequests: []clause.Expression{
						cursorpagination.Must(2, cursorpagination.Asc("id", nil)),
					},
					want: [][]*TestStruct{
						{
							{Code: "A", Price: 2},
							{Code: "B", Price: 1},
						},
					},
				},
				"Paged 1/2 items, sort by id desc": {
					pageRequests: []clause.Expression{
						pagepagination.Must(0, 2, pagegeneric.Desc("id")),
					},
					cursorRequests: []clause.Expression{
						cursorpagination.Must(2, cursorpagination.Desc("id", nil)),
					},
					want: [][]*TestStruct{
						{
							{Code: "D", Price: 1},
							{Code: "C", Price: 3},
						},
					},
				},
			}
			for name, test := range tests {
				t.Run(name, func(t *testing.T) {
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

					for i, pageRequest := range test.pageRequests {
						var got []*TestStruct
						if err := db.Clauses(pageRequest).Find(&got).Error; err != nil {
							t.Fatalf("failed to query page 0: %v", err)
						}
						compareTestStructs(t, test.want[i], got)
					}

					for i, pageRequest := range test.cursorRequests {
						var got []*TestStruct
						if err := db.Clauses(pageRequest).Find(&got).Error; err != nil {
							t.Fatalf("failed to query page 0: %v", err)
						}
						compareTestStructs(t, test.want[i], got)
					}
				})
			}
		})
	}
}

func TestSimplePaginationUsingNext(t *testing.T) {
	t.Parallel()

	testData := func() []*TestStruct {
		return []*TestStruct{
			{Code: "A", Price: 2},
			{Code: "B", Price: 1},
			{Code: "C", Price: 3},
			{Code: "D", Price: 1},
			{Code: "E", Price: 5},
		}
	}

	for engine, dbFunc := range dbEngines {
		t.Run(engine, func(t *testing.T) {
			t.Parallel()

			tests := map[string]struct {
				size          int
				cursorValues  []cursorpagination.Cursor
				sort          pagegeneric.Sort
				wantNextTimes int
				want          [][]*TestStruct
			}{
				"size 2": {
					size: 2,
					cursorValues: []cursorpagination.Cursor{
						cursorpagination.Asc("id", nil),
					},
					sort: pagegeneric.Sort{
						pagegeneric.Asc("code"),
					},
					wantNextTimes: 2,
					want: [][]*TestStruct{
						{
							{Code: "A", Price: 2},
							{Code: "B", Price: 1},
						},
						{
							{Code: "C", Price: 3},
							{Code: "D", Price: 1},
						},
						{
							{Code: "E", Price: 5},
						},
					},
				},
			}
			for name, test := range tests {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					db, deferFunc, errStartingContainer := dbFunc(t.Context())
					if errStartingContainer != nil {
						t.Fatalf("failed to start container: %v", errStartingContainer)
					}
					defer deferFunc()
					setupDB(t, db)
					testdata := testData()
					if err := db.CreateInBatches(testdata, len(testdata)).Error; err != nil {
						t.Fatalf("failed to create test data: %v", err)
					}
					// page pagination
					pageRequest := pagepagination.Must(0, test.size, test.sort...)
					testPaginationSequence(t, db, pageRequest, test.want)

					// cursor pagination
					cursorRequest := cursorpagination.Must(test.size, test.cursorValues...)
					testPaginationSequence(t, db, cursorRequest, test.want)
				})
			}
		})
	}
}

func testPaginationSequence(t *testing.T, db *gorm.DB, request any, want [][]*TestStruct) {
	t.Helper()

	hasNext := pagegeneric.NextPossible(true)
	gotTimes := -1
	for hasNext {
		gotTimes++
		var got []*TestStruct
		if err := db.Clauses(request.(clause.Expression)).Find(&got).Error; err != nil {
			t.Fatalf("failed to query page: %v", err)
		}
		compareTestStructs(t, want[gotTimes], got)
		switch r := request.(type) {
		case *pagepagination.Pagination:
			request, hasNext = r.Next()
		case *cursorpagination.Pagination:
			request, hasNext = r.Next()
		default:
			t.Fatalf("unknown pagination type: %T", r)
		}
	}
	if len(want) != gotTimes+1 {
		t.Errorf("expected %d pages, got %d", len(want), gotTimes+1)
	}
}

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
				"delete statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); DELETE FROM test_structs; --"),
					},
				},
				"update statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); UPDATE test_structs SET code = 'hacked'; --"),
					},
				},
				"insert statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); INSERT INTO test_structs (code, price) VALUES ('X', 99); --"),
					},
				},
				"drop table statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); DROP TABLE test_structs; --"),
					},
				},
				"truncate statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); TRUNCATE TABLE test_structs; --"),
					},
				},
				"alter table statement": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("id", "1); ALTER TABLE test_structs ADD COLUMN injected INT; --"),
					},
				},
				"boolean tautology": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("code", "' OR 1=1 --"),
					},
				},
				"union select": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("code", "' UNION SELECT 'HACK', 999 --"),
					},
				},
				"stacked comments": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("code", "A'/**/;DELETE FROM test_structs;--"),
					},
				},
				"multi cursor with malicious values": {
					cursors: []cursorpagination.Cursor{
						cursorpagination.Asc("code", "B'; UPDATE test_structs SET price = 0; --"),
						cursorpagination.Desc("price", "999); DELETE FROM test_structs; --"),
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

					var got int64
					if err := db.Model(&TestStruct{}).Count(&got).Error; err != nil {
						t.Fatalf("failed to count records: %v", err)
					}

					if got != 4 {
						t.Errorf("unexpected count of records: got=%d, want=%d", got, 4)
					}

					wantRows := testData()
					var gotRows []*TestStruct
					if err := db.Order("code ASC").Find(&gotRows).Error; err != nil {
						t.Fatalf("failed to read records: %v", err)
					}

					if diff := cmp.Diff(
						wantRows,
						gotRows,
						cmpopts.IgnoreFields(TestStruct{}, "ID", "CreatedAt", "UpdatedAt", "DeletedAt"),
					); diff != "" {
						t.Errorf("records changed unexpectedly (-want +got):\n%s", diff)
					}
				})
			}
		})
	}
}

func TestParametrizedSQLQueriesCursorPagination(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		payload string
	}{
		"delete statement": {
			payload: "1); DELETE FROM test_structs; --",
		},
		"update statement": {
			payload: "1); UPDATE test_structs SET code = 'hacked'; --",
		},
		"insert statement": {
			payload: "1); INSERT INTO test_structs (code, price) VALUES ('X', 99); --",
		},
		"drop table statement": {
			payload: "1); DROP TABLE test_structs; --",
		},
		"truncate statement": {
			payload: "1); TRUNCATE TABLE test_structs; --",
		},
		"alter table statement": {
			payload: "1); ALTER TABLE test_structs ADD COLUMN injected INT; --",
		},
		"boolean tautology": {
			payload: "' OR 1=1 --",
		},
		"union select": {
			payload: "' UNION SELECT 1 --",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			payload := test.payload
			db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{DryRun: true})
			if err != nil {
				t.Fatalf("failed to open dry-run db: %v", err)
			}

			cursorRequest := cursorpagination.Must(2, cursorpagination.Asc("code", payload))
			tx := db.Clauses(cursorRequest).Model(&TestStruct{}).Find(&[]*TestStruct{})
			if tx.Error != nil {
				t.Fatalf("failed to build query for payload %q: %v", payload, tx.Error)
			}

			querySQL := tx.Statement.SQL.String()
			if strings.Contains(querySQL, payload) {
				t.Fatalf("payload leaked into SQL for %q: %q", payload, querySQL)
			}

			if !strings.Contains(querySQL, "code > ?") {
				t.Fatalf("expected placeholder comparison in SQL, got: %q", querySQL)
			}

			foundPayloadInVars := false
			for _, v := range tx.Statement.Vars {
				if s, ok := v.(string); ok && s == payload {
					foundPayloadInVars = true
					break
				}
			}

			if !foundPayloadInVars {
				t.Fatalf("payload was not passed as bind variable: %q", payload)
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

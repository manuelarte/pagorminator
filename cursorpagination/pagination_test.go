package cursorpagination

import (
	"errors"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestStruct struct {
	gorm.Model

	Code  string
	Price uint
}

func TestNewPageRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		size        int
		cursors     []Cursor
		expectedErr error
	}{
		"negative size": {
			size:        -1,
			cursors:     []Cursor{Asc("id", nil)},
			expectedErr: ErrSizeCantBeNegative,
		},
		"invalid order": {
			size:        10,
			cursors:     []Cursor{{Column: "id"}},
			expectedErr: ErrOrderNotValid,
		},
		"missing order": {
			size:        10,
			cursors:     nil,
			expectedErr: ErrOrderRequired,
		},
		"mixed cursor values": {
			size:        10,
			cursors:     []Cursor{Asc("id", 3), Desc("price", nil)},
			expectedErr: ErrCursorValueNotValid,
		},
		"first page without cursor values": {
			size:    10,
			cursors: []Cursor{Asc("id", nil), Desc("price", nil)},
		},
		"valid request": {
			size:    10,
			cursors: []Cursor{Asc("id", 1), Desc("price", 2)},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := New(test.size, test.cursors...)
			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("expected error %v, got %v", test.expectedErr, err)
			}

			if test.expectedErr != nil {
				return
			}

			if got.GetSize() != test.size {
				t.Errorf("size expected %d, got %d", test.size, got.GetSize())
			}

			gotCursors := got.GetCursors()
			if len(gotCursors) != len(test.cursors) {
				t.Fatalf("cursor count expected %d, got %d", len(test.cursors), len(gotCursors))
			}
		})
	}
}

func TestGetCursorsClone(t *testing.T) {
	t.Parallel()

	pageRequest := Must(10, Asc("id", 1))
	got := pageRequest.GetCursors()
	got[0].Column = "changed"

	again := pageRequest.GetCursors()
	if again[0].Column != "id" {
		t.Fatalf("expected cloned cursors, got %q", again[0].Column)
	}
}

func TestBuildCursorWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cursors  []Cursor
		wantSQL  string
		wantVars []any
	}{
		"single asc": {
			cursors:  []Cursor{Asc("id", 3)},
			wantSQL:  "(id > ?)",
			wantVars: []any{3},
		},
		"single desc": {
			cursors:  []Cursor{Desc("id", 3)},
			wantSQL:  "(id < ?)",
			wantVars: []any{3},
		},
		"multi column mixed directions": {
			cursors:  []Cursor{Asc("code", "A"), Desc("price", 10)},
			wantSQL:  "(code > ?) OR (code = ? AND price < ?)",
			wantVars: []any{"A", "A", 10},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pageRequest := Must(5, test.cursors...)
			gotSQL, gotVars := pageRequest.buildCursorWhere()

			if gotSQL != test.wantSQL {
				t.Fatalf("sql expected %q, got %q", test.wantSQL, gotSQL)
			}

			if len(gotVars) != len(test.wantVars) {
				t.Fatalf("vars size expected %d, got %d", len(test.wantVars), len(gotVars))
			}

			for i := range test.wantVars {
				if gotVars[i] != test.wantVars[i] {
					t.Fatalf("var[%d] expected %v, got %v", i, test.wantVars[i], gotVars[i])
				}
			}
		})
	}
}

func TestCursorPaginationSingleColumn(t *testing.T) {
	t.Parallel()

	db := setupDB(t)
	toMigrate := []*TestStruct{
		{Code: "A", Price: 2},
		{Code: "B", Price: 1},
		{Code: "C", Price: 3},
	}

	if txCreate := db.CreateInBatches(&toMigrate, len(toMigrate)); txCreate.Error != nil {
		t.Fatal(txCreate.Error)
	}

	pageRequest := Must(2, Desc("price", 2))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 1 || products[0].Price != 1 {
		t.Fatalf("unexpected result: %+v", products)
	}
}

func TestCursorPaginationMultiColumnSort(t *testing.T) {
	t.Parallel()

	db := setupDB(t)
	toMigrate := []*TestStruct{
		{Code: "A", Price: 1},
		{Code: "A", Price: 3},
		{Code: "A", Price: 2},
		{Code: "B", Price: 3},
	}

	if txCreate := db.CreateInBatches(&toMigrate, len(toMigrate)); txCreate.Error != nil {
		t.Fatal(txCreate.Error)
	}

	pageRequest := Must(3, Asc("code", "A"), Desc("price", 2))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	expected := []struct {
		code  string
		price uint
	}{
		{code: "A", price: 1},
		{code: "B", price: 3},
	}
	assertProducts(t, expected, products)
}

func TestCursorPaginationUnPaged(t *testing.T) {
	t.Parallel()

	db := setupDB(t)
	toMigrate := []*TestStruct{
		{Code: "A", Price: 1},
		{Code: "B", Price: 2},
		{Code: "C", Price: 3},
	}

	if txCreate := db.CreateInBatches(&toMigrate, len(toMigrate)); txCreate.Error != nil {
		t.Fatal(txCreate.Error)
	}

	pageRequest := Must(0, Asc("id", 0))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 3 {
		t.Fatalf("unexpected result size: %d", len(products))
	}
}

func TestCursorPaginationFirstPageWithoutWhere(t *testing.T) {
	t.Parallel()

	db := setupDB(t)
	toMigrate := []*TestStruct{
		{Code: "A", Price: 1},
		{Code: "B", Price: 3},
		{Code: "C", Price: 2},
	}

	if txCreate := db.CreateInBatches(&toMigrate, len(toMigrate)); txCreate.Error != nil {
		t.Fatal(txCreate.Error)
	}

	pageRequest := Must(2, Desc("price", nil))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 2 || products[0].Price != 3 || products[1].Price != 2 {
		t.Fatalf("unexpected result: %+v", products)
	}
}

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to connect database")
	}

	err = db.AutoMigrate(&TestStruct{})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func assertProducts(t *testing.T, expected []struct {
	code  string
	price uint
}, products []*TestStruct,
) {
	t.Helper()

	if len(products) != len(expected) {
		t.Fatalf("unexpected result size: got %d, want %d", len(products), len(expected))
	}

	for i := range expected {
		if products[i].Code != expected[i].code || products[i].Price != expected[i].price {
			t.Fatalf(
				"unexpected result at %d: got (%s,%d), want (%s,%d)",
				i,
				products[i].Code,
				products[i].Price,
				expected[i].code,
				expected[i].price,
			)
		}
	}
}

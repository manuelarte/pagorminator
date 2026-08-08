package pagorminator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator/cursorpagination"
	"github.com/manuelarte/pagorminator/pagegeneric"
	"github.com/manuelarte/pagorminator/pagepagination"
)

type TestStruct struct {
	gorm.Model

	Code  string
	Price uint
}

type TestProduct struct {
	gorm.Model

	Code  string
	Price TestPrice
}
type TestPrice struct {
	gorm.Model

	Amount        uint
	Currency      string
	TestProductID uint
}

type expectedPagination struct {
	page             int
	size             int
	sort             []pagegeneric.Order
	totalElements    int64
	totalElementsSet bool
}

type expectedCursorPagination struct {
	Size             int
	Cursors          []expectedCursor
	TotalElements    int64
	TotalElementsSet bool
}

type expectedCursor struct {
	Column string
	Value  any
	Order  string
}

func TestNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(1, 1),
			want: &expectedPagination{
				page:             1,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(0, 2),
			want: &expectedPagination{
				page:             0,
				size:             2,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			db.Clauses(test.pageRequest).Find(&products)

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *pagepagination.Pagination
		wantPage       *expectedPagination
		expectedResult []*TestStruct
	}{
		"Paged 1/2 items, sort by id asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1}, {Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(1, 1, pagegeneric.Asc("id")),
			wantPage: &expectedPagination{
				page:             1,
				size:             1,
				sort:             []pagegeneric.Order{pagegeneric.Asc("id")},
				totalElements:    2,
				totalElementsSet: true,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
		},
		"Paged 1/2 items, sort by id desc": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(1, 1, pagegeneric.Desc("id")),
			wantPage: &expectedPagination{
				page:             1,
				size:             1,
				sort:             []pagegeneric.Order{pagegeneric.Desc("id")},
				totalElements:    2,
				totalElementsSet: true,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
			},
		},
		"Paged 1/2 items, sort by code asc, and price desc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 11}, Code: "1", Price: 11},
			},
			pageRequest: pagepagination.Must(0, 5, pagegeneric.Asc("code"), pagegeneric.Desc("price")),
			wantPage: &expectedPagination{
				page:             0,
				size:             5,
				sort:             []pagegeneric.Order{pagegeneric.Asc("code"), pagegeneric.Desc("price")},
				totalElements:    3,
				totalElementsSet: true,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 11}, Code: "1", Price: 11},
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			if tx := db.Clauses(test.pageRequest).Find(&products); tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.wantPage,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Fatalf("diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(
				test.expectedResult,
				products,
				cmpopts.IgnoreFields(TestStruct{}, "Model"),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

// TODO(manuelarte): migrate to tests and do cursor pagination tests too.
func TestWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagepagination.Pagination
		where       string
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price < 100",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price > 100",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    0,
				totalElementsSet: true,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price > 50",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: pagepagination.Must(0, 1),
			where:       "price > 50",
			want: &expectedPagination{
				page:             0,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			if txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate)); txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			if tx := db.Clauses(test.pageRequest).Where(test.where).Find(&products); tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *pagepagination.Pagination
		where          string
		wantPage       *expectedPagination
		expectedResult []*TestStruct
	}{
		"Paged 0 1/2 items, two items filtered out, sort by price asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
			pageRequest: pagepagination.Must(0, 1, pagegeneric.Asc("price")),
			where:       "price > 50",
			wantPage: &expectedPagination{
				page:             0,
				size:             1,
				sort:             []pagegeneric.Order{pagegeneric.Asc("price")},
				totalElements:    2,
				totalElementsSet: true,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
			},
		},
		"Paged 0 1/2 items, two items filtered out, sort by price desc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
			pageRequest: pagepagination.Must(0, 1, pagegeneric.Desc("price")),
			where:       "price > 50",
			wantPage: &expectedPagination{
				page:             0,
				size:             1,
				sort:             []pagegeneric.Order{pagegeneric.Desc("price")},
				totalElements:    2,
				totalElementsSet: true,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			if txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate)); txCreate.Error != nil {
				t.Fatalf("error creating products: %v", txCreate.Error)
			}

			var products []*TestStruct

			if tx := db.Clauses(test.pageRequest).Where(test.where).Find(&products); tx.Error != nil {
				t.Fatalf("error querying products: %v", tx.Error)
			}

			if diff := cmp.Diff(
				test.wantPage,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(
				test.expectedResult,
				products,
				cmpopts.IgnoreFields(TestStruct{}, "Model"),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithPreload(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(0, 1),
			want: &expectedPagination{
				page:             0,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(1, 1),
			want: &expectedPagination{
				page:             1,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			if txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate)); txCreate.Error != nil {
				t.Fatalf("error creating products: %v", txCreate.Error)
			}

			var products []*TestProduct

			if tx := db.Clauses(test.pageRequest).Preload("Price").Find(&products); tx.Error != nil {
				t.Fatalf("error querying products: %v", tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithPreloadAndWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
				{Code: "5", Price: TestPrice{Amount: 5, Currency: "EUR"}},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    4,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
				{Code: "5", Price: TestPrice{Amount: 5, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(0, 2),
			want: &expectedPagination{
				page:             0,
				size:             2,
				totalElements:    4,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Preload("Price").Where("code > 1").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoins(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(0, 1),
			want: &expectedPagination{
				page:             0,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Joins("Price").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoinsWhereClause(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagepagination.Pagination
		where       any
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "1=1",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(0, 1),
			where:       "Price.amount > 1",
			want: &expectedPagination{
				page:             0,
				size:             1,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
			},
			pageRequest: pagepagination.Must(0, 2),
			where:       "Price.amount >= 2",
			want: &expectedPagination{
				page:             0,
				size:             2,
				totalElements:    3,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Joins("Price").Where(test.where).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTable(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(1, 1),
			want: &expectedPagination{
				page:             1,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagepagination.Must(0, 2),
			want: &expectedPagination{
				page:             0,
				size:             2,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var result map[string]any

			tx := db.Clauses(test.pageRequest).Table("test_structs").Find(&result)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTableWithWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagepagination.Pagination
		where       string
		want        *expectedPagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price < 100",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price > 100",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    0,
				totalElementsSet: true,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: pagepagination.UnPaged(),
			where:       "price > 50",
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: pagepagination.Must(0, 1),
			where:       "price > 50",
			want: &expectedPagination{
				page:             0,
				size:             1,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products map[string]any

			tx := db.Clauses(test.pageRequest).Where(test.where).Table("test_structs").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDistinct(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagepagination.Pagination
		want        *expectedPagination
	}{
		"UnPaged two items, same price": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 1},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    1,
				totalElementsSet: true,
			},
		},
		"UnPaged four items, two different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 1},
				{Code: "4", Price: 2},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    2,
				totalElementsSet: true,
			},
		},
		"UnPaged four items, four different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 3},
				{Code: "4", Price: 4},
			},
			pageRequest: pagepagination.UnPaged(),
			want: &expectedPagination{
				page:             0,
				size:             0,
				totalElements:    4,
				totalElementsSet: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products map[string]any

			tx := db.Clauses(test.pageRequest).Distinct("price").Model(&TestStruct{}).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(
				test.want,
				toExpectedPagination(test.pageRequest),
				cmp.AllowUnexported(expectedPagination{}),
			); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPaGorminatorNil(t *testing.T) {
	t.Parallel()

	db := setupDB(t)

	var products []*TestStruct
	db.Clauses(nil).Find(&products)
}

func TestContextCancelledAfterPagorminator(t *testing.T) {
	t.Parallel()

	toMigrate := []*TestStruct{
		{Code: "1", Price: 1},
		{Code: "2", Price: 2},
	}
	pageRequest := pagepagination.UnPaged()
	want := &expectedPagination{
		page:             0,
		size:             0,
		totalElements:    2,
		totalElementsSet: true,
	}

	db := setupDB(t).Debug()

	err := db.CreateInBatches(&toMigrate, len(toMigrate)).Error
	if err != nil {
		t.Fatal(err)
	}

	var products []*TestStruct

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// cancel after pagorminator:count has run
	errRegisteringCallback := db.Callback().Query().After("pagorminator:count").Register("cancel:ctx", func(_ *gorm.DB) {
		cancel()
	})
	if errRegisteringCallback != nil {
		t.Fatalf("can't register callback: %v", errRegisteringCallback)
	}

	errFindingProducts := db.WithContext(ctx).Clauses(pageRequest).Find(&products).Error
	if !errors.Is(errFindingProducts, context.Canceled) {
		t.Fatalf("expecting context cancelled: %v", errFindingProducts)
	}

	if diff := cmp.Diff(want, toExpectedPagination(pageRequest), cmp.AllowUnexported(expectedPagination{})); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
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

	pageRequest := cursorpagination.Must(2, cursorpagination.Desc("price", 2))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 1 || products[0].Price != 1 {
		t.Fatalf("unexpected result: %+v", products)
	}

	wantPage := &expectedCursorPagination{
		Size: 2,
		Cursors: []expectedCursor{
			{Column: "price", Value: 2, Order: "price DESC"},
		},
		TotalElements:    1,
		TotalElementsSet: true,
	}
	if diff := cmp.Diff(wantPage, toExpectedCursorPagination(pageRequest)); diff != "" {
		t.Fatalf("diff (-want +got):\n%s", diff)
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

	pageRequest := cursorpagination.Must(3, cursorpagination.Asc("code", "A"), cursorpagination.Desc("price", 2))

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
	assertStructs(t, expected, products)

	wantPage := &expectedCursorPagination{
		Size: 3,
		Cursors: []expectedCursor{
			{Column: "code", Value: "A", Order: "code ASC"},
			{Column: "price", Value: 2, Order: "price DESC"},
		},
		TotalElements:    2,
		TotalElementsSet: true,
	}
	if diff := cmp.Diff(wantPage, toExpectedCursorPagination(pageRequest)); diff != "" {
		t.Fatalf("diff (-want +got):\n%s", diff)
	}
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

	pageRequest := cursorpagination.Must(0, cursorpagination.Asc("id", 0))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 3 {
		t.Fatalf("unexpected result size: %d", len(products))
	}

	wantPage := &expectedCursorPagination{
		Size: 0,
		Cursors: []expectedCursor{
			{Column: "id", Value: 0, Order: "id ASC"},
		},
		TotalElements:    3,
		TotalElementsSet: true,
	}
	if diff := cmp.Diff(wantPage, toExpectedCursorPagination(pageRequest)); diff != "" {
		t.Fatalf("diff (-want +got):\n%s", diff)
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

	pageRequest := cursorpagination.Must(2, cursorpagination.Desc("price", nil))

	var products []*TestStruct
	if tx := db.Clauses(pageRequest).Find(&products); tx.Error != nil {
		t.Fatal(tx.Error)
	}

	if len(products) != 2 || products[0].Price != 3 || products[1].Price != 2 {
		t.Fatalf("unexpected result: %+v", products)
	}

	wantPage := &expectedCursorPagination{
		Size: 2,
		Cursors: []expectedCursor{
			{Column: "price", Value: nil, Order: "price DESC"},
		},
		TotalElements:    3,
		TotalElementsSet: true,
	}
	if diff := cmp.Diff(wantPage, toExpectedCursorPagination(pageRequest)); diff != "" {
		t.Fatalf("diff (-want +got):\n%s", diff)
	}
}

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(&TestStruct{}, &TestProduct{}, &TestPrice{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Use(PaGorminator{Debug: true})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func toExpectedPagination(actual *pagepagination.Pagination) *expectedPagination {
	if actual == nil {
		return nil
	}

	totalElements, totalElementsSet := actual.GetTotalElements()

	return &expectedPagination{
		page:             actual.GetPage(),
		size:             actual.GetSize(),
		sort:             actual.GetSort(),
		totalElements:    totalElements,
		totalElementsSet: totalElementsSet,
	}
}

func toExpectedCursorPagination(actual *cursorpagination.Pagination) *expectedCursorPagination {
	if actual == nil {
		return nil
	}

	cursors := actual.GetCursors()

	expectedCursors := make([]expectedCursor, len(cursors))
	for i, cursor := range cursors {
		order := ""
		if cursor.GetOrder() != nil {
			order = cursor.GetOrder().GormString()
		}

		expectedCursors[i] = expectedCursor{
			Column: cursor.GetColumn(),
			Value:  cursor.GetValue(),
			Order:  order,
		}
	}

	totalElements, totalElementsSet := actual.GetTotalElements()

	return &expectedCursorPagination{
		Size:             actual.GetSize(),
		Cursors:          expectedCursors,
		TotalElements:    totalElements,
		TotalElementsSet: totalElementsSet,
	}
}

func assertStructs(t *testing.T, expected []struct {
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

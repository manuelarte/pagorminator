package pagorminator

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(1, 1),
			want: &Pagination{
				page:             1,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(0, 2),
			want: &Pagination{
				page:             0,
				size:             2,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *Pagination
		wantPage       *Pagination
		expectedResult []*TestStruct
	}{
		"Paged 1/2 items, sort by id asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1}, {Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(1, 1, Asc("id")),
			wantPage: &Pagination{
				page:             1,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
				sort:             []Order{Asc("id")},
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
		},
		"Paged 1/2 items, sort by id desc": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(1, 1, Desc("id")),
			wantPage: &Pagination{
				page:             1,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
				sort:             []Order{Desc("id")},
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
			pageRequest: MustPageRequest(0, 5, Asc("code"), Desc("price")),
			wantPage: &Pagination{
				page:             0,
				size:             5,
				totalElementsSet: true,
				totalElements:    3,
				sort:             []Order{Asc("code"), Desc("price")},
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

			if diff := cmp.Diff(test.pageRequest, test.wantPage, paginationCmpOpt()); diff != "" {
				t.Fatalf("diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(products, test.expectedResult, cmpopts.IgnoreFields(TestStruct{}, "Model")); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *Pagination
		where       string
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price < 100",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price > 100",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    0,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: UnPaged(),
			where:       "price > 50",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: MustPageRequest(0, 1),
			where:       "price > 50",
			want: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *Pagination
		where          string
		wantPage       *Pagination
		expectedResult []*TestStruct
	}{
		"Paged 0 1/2 items, two items filtered out, sort by price asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
			pageRequest: MustPageRequest(0, 1, Asc("price")),
			where:       "price > 50",
			wantPage: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
				sort:             []Order{Asc("price")},
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
			pageRequest: MustPageRequest(0, 1, Desc("price")),
			where:       "price > 50",
			wantPage: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
				sort:             []Order{Desc("price")},
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

			if diff := cmp.Diff(test.pageRequest, test.wantPage, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(products, test.expectedResult, cmpopts.IgnoreFields(TestStruct{}, "Model")); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithPreload(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			want: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 1, size: 1},
			want: &Pagination{
				page:             1,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithPreloadAndWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
				{Code: "5", Price: TestPrice{Amount: 5, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    4,
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
			pageRequest: &Pagination{page: 0, size: 2},
			want: &Pagination{
				page:             0,
				size:             2,
				totalElementsSet: true,
				totalElements:    4,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoins(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			want: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoinsWhereClause(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *Pagination
		where       any
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			where:       "1=1",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			where:       "Price.amount > 1",
			want: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 2},
			where:       "Price.amount >= 2",
			want: &Pagination{
				page:             0,
				size:             2,
				totalElementsSet: true,
				totalElements:    3,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTable(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(1, 1),
			want: &Pagination{
				page:             1,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: MustPageRequest(0, 2),
			want: &Pagination{
				page:             0,
				size:             2,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTableWithWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *Pagination
		where       string
		want        *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price < 100",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price > 100",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    0,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: UnPaged(),
			where:       "price > 50",
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: MustPageRequest(0, 1),
			where:       "price > 50",
			want: &Pagination{
				page:             0,
				size:             1,
				totalElementsSet: true,
				totalElements:    2,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDistinct(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *Pagination
		want        *Pagination
	}{
		"UnPaged two items, same price": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 1},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    1,
			},
		},
		"UnPaged four items, two different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 1},
				{Code: "4", Price: 2},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    2,
			},
		},
		"UnPaged four items, four different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 3},
				{Code: "4", Price: 4},
			},
			pageRequest: UnPaged(),
			want: &Pagination{
				page:             0,
				size:             0,
				totalElementsSet: true,
				totalElements:    4,
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

			if diff := cmp.Diff(test.pageRequest, test.want, paginationCmpOpt()); diff != "" {
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

func paginationCmpOpt() cmp.Options {
	return cmp.Options{
		cmp.AllowUnexported(Pagination{}),
		cmpopts.IgnoreFields(Pagination{}, "mu"),
	}
}

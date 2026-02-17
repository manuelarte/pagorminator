package pagorminator

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator/pagination"
	"github.com/manuelarte/pagorminator/pagination/sort"
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
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(2)

				return p
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(1, 1),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1)
				p.SetTotalElements(2)

				return p
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(0, 2),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 2)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want := test.wantFn()
			db := setupDB(t)

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			db.Clauses(test.pageRequest).Find(&products)

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *pagination.Pagination
		wantFn         func() *pagination.Pagination
		expectedResult []*TestStruct
	}{
		"Paged 1/2 items, sort by id asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1}, {Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(1, 1, sort.Asc("id")),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1, sort.Asc("id"))
				p.SetTotalElements(2)

				return p
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
		},
		"Paged 1/2 items, sort by id desc": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(1, 1, sort.Desc("id")),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1, sort.Desc("id"))
				p.SetTotalElements(2)

				return p
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			if txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate)); txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			if tx := db.Clauses(test.pageRequest).Find(&products); tx.Error != nil {
				t.Fatalf("error finding products: %v", tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}

			if !cmp.Equal(products, test.expectedResult, cmpopts.IgnoreFields(TestStruct{}, "Model")) {
				t.Errorf("expected result to be %+v, got %+v", test.expectedResult, products)
			}
		})
	}
}

func TestWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagination.Pagination
		where       string
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price < 100",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price > 100",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(0)

				return p
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: pagination.Must(0, 1),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			tx := db.Clauses(test.pageRequest).Where(test.where).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSortWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *pagination.Pagination
		where          string
		wantFn         func() *pagination.Pagination
		expectedResult []*TestStruct
	}{
		"Paged 0 1/2 items, two items filtered out, sort by price asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
			pageRequest: pagination.Must(0, 1, sort.Asc("price")),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1, sort.Asc("price"))
				p.SetTotalElements(2)

				return p
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
			pageRequest: pagination.Must(0, 1, sort.Desc("price")),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1, sort.Desc("price"))
				p.SetTotalElements(2)

				return p
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

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			tx := db.Clauses(test.pageRequest).Where(test.where).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Fatalf("diff (-want +got):\n%s", diff)
			}

			if !cmp.Equal(products, test.expectedResult, cmpopts.IgnoreFields(TestStruct{}, "Model")) {
				t.Fatalf("expected result to be %+v, got %+v", test.expectedResult, products)
			}
		})
	}
}

func TestWithPreload(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagination.Must(0, 1),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				p.SetTotalElements(2)

				return p
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagination.Must(1, 1),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Preload("Price").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithPreloadAndWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
				{Code: "5", Price: TestPrice{Amount: 5, Currency: "EUR"}},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(4)

				return p
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
			pageRequest: pagination.Must(0, 2),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 2)
				p.SetTotalElements(4)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Preload("Price").Where("code > 1").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoins(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagination.Must(0, 1),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Joins("Price").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithJoinsWhereClause(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestProduct
		pageRequest *pagination.Pagination
		where       any
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: pagination.UnPaged(),
			where:       "1=1",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: pagination.Must(0, 1),
			where:       "Price.amount > 1",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
				{Code: "3", Price: TestPrice{Amount: 3, Currency: "EUR"}},
				{Code: "4", Price: TestPrice{Amount: 4, Currency: "EUR"}},
			},
			pageRequest: pagination.Must(0, 2),
			where:       "Price.amount >= 2",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 2)
				p.SetTotalElements(3)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestProduct

			tx := db.Clauses(test.pageRequest).Joins("Price").Where(test.where).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTable(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(2)

				return p
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(1, 1),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1)
				p.SetTotalElements(2)

				return p
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: pagination.Must(0, 2),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 2)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var result map[string]any

			tx := db.Clauses(test.pageRequest).Table("test_structs").Find(&result)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTableWithWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagination.Pagination
		where       string
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price < 100",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price > 100",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(0)

				return p
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: pagination.UnPaged(),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: pagination.Must(0, 1),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				p.SetTotalElements(2)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products map[string]any

			tx := db.Clauses(test.pageRequest).Where(test.where).Table("test_structs").Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDistinct(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate   []*TestStruct
		pageRequest *pagination.Pagination
		wantFn      func() *pagination.Pagination
	}{
		"UnPaged two items, same price": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 1},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(1)

				return p
			},
		},
		"UnPaged four items, two different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 1},
				{Code: "4", Price: 2},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(2)

				return p
			},
		},
		"UnPaged four items, four different prices": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 3},
				{Code: "4", Price: 4},
			},
			pageRequest: pagination.UnPaged(),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 0)
				p.SetTotalElements(4)

				return p
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)

			want := test.wantFn()

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products map[string]any

			tx := db.Clauses(test.pageRequest).Distinct("price").Model(&TestStruct{}).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if diff := cmp.Diff(test.pageRequest, want, paginationCmpOpt()); diff != "" {
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
		cmp.AllowUnexported(pagination.Pagination{}),
		cmpopts.IgnoreFields(pagination.Pagination{}, "mu"),
	}
}

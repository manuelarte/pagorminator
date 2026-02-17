package pagorminator

import (
	"fmt"
	"testing"

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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(2)

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

			// TODO: change for cmp.Equal
			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
			pageRequest: pagination.Must(1, 1, sort.MustOrder("id", sort.ASC)),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1)
				_ = p.SetTotalElements(2)

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
			pageRequest: pagination.Must(1, 1, sort.MustOrder("id", sort.DESC)),
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(1, 1)
				_ = p.SetTotalElements(2)

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

			txCreate := db.CreateInBatches(&test.toMigrate, len(test.toMigrate))
			if txCreate.Error != nil {
				t.Fatal(txCreate.Error)
			}

			var products []*TestStruct

			tx := db.Clauses(test.pageRequest).Find(&products)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
			}

			if !equalsArrays(products, test.expectedResult) {
				t.Fatalf("expected result to be %+v, got %+v", test.expectedResult, products)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(0)

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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
			pageRequest: pagination.Must(0, 1, sort.MustOrder("price", sort.ASC)),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				_ = p.SetTotalElements(2)

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
			pageRequest: pagination.Must(0, 1, sort.MustOrder("price", sort.DESC)),
			where:       "price > 50",
			wantFn: func() *pagination.Pagination {
				p := pagination.Must(0, 1)
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
			}

			if !equalsArrays(products, test.expectedResult) {
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(4)

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
				_ = p.SetTotalElements(4)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(3)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(0)

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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
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
				_ = p.SetTotalElements(1)

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
				_ = p.SetTotalElements(2)

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
				_ = p.SetTotalElements(4)

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

			if !equalPageRequests(test.pageRequest, want) {
				t.Fatalf("expected page to be %+v, got %+v", want, test.pageRequest)
			}
		})
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

func equalPageRequests(p1, p2 *pagination.Pagination) bool {
	// cmp.Equal(p1, p2, cmp.AllowUnexported(pagination.Pagination{})
	return p1.Page() == p2.Page() &&
		p1.Size() == p2.Size() &&
		p1.TotalElements() == p2.TotalElements() &&
		p1.GetTotalPages() == p2.GetTotalPages()
}

func equalsTestStruct(t1, t2 *TestStruct) bool {
	sameID := t1.ID == t2.ID
	sameCode := t1.Code == t2.Code
	samePrice := t1.Price == t2.Price

	return sameID && sameCode && samePrice
}

func equalsArrays(a1, a2 []*TestStruct) bool {
	if len(a1) != len(a2) {
		return false
	}

	for i, item := range a1 {
		if !equalsTestStruct(item, a2[i]) {
			return false
		}
	}

	return true
}

func TestPaGorminator_Nil(t *testing.T) {
	t.Parallel()

	db := setupDB(t)

	var products []*TestStruct
	db.Clauses(nil).Find(&products)
}

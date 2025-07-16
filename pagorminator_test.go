package pagorminator //nolint:testpackage // testing expected page

import (
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

func TestPaGorminator_NoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  *Pagination
		expectedPage *Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 2,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(1, 1),
			expectedPage: &Pagination{
				page:          1,
				size:          1,
				totalElements: 2,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(0, 2),
			expectedPage: &Pagination{
				page:          0,
				size:          2,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestStruct

			db.Clauses(test.pageRequest).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminator_SortNoWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *Pagination
		expectedPage   *Pagination
		expectedResult []*TestStruct
	}{
		"Paged 1/2 items, sort by id asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1}, {Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(1, 1, mustNewOrder("id", ASC)),
			expectedPage: &Pagination{
				page:          1,
				size:          1,
				totalElements: 2,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
			},
		},
		"Paged 1/2 items, sort by id desc": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(1, 1, mustNewOrder("id", DESC)),
			expectedPage: &Pagination{
				page:          1,
				size:          1,
				totalElements: 2,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestStruct

			db.Debug().Clauses(test.pageRequest).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}

			if !equalsArrays(products, test.expectedResult) {
				t.Fatalf("expected result to be %+v, got %+v", test.expectedResult, products)
			}
		})
	}
}

func TestPaGorminator_Where(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  *Pagination
		where        string
		expectedPage *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price < 100",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price > 100",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 0,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: UnPaged(),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: mustPageRequestOf(0, 1),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestStruct

			db.Clauses(test.pageRequest).Where(test.where).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminator_SortWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate      []*TestStruct
		pageRequest    *Pagination
		where          string
		expectedPage   *Pagination
		expectedResult []*TestStruct
	}{
		"Paged 0 1/2 items, two items filtered out, sort by price asc": {
			toMigrate: []*TestStruct{
				{Model: gorm.Model{ID: 1}, Code: "1", Price: 1},
				{Model: gorm.Model{ID: 2}, Code: "2", Price: 2},
				{Model: gorm.Model{ID: 3}, Code: "3", Price: 100},
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
			pageRequest: mustPageRequestOf(0, 1, mustNewOrder("price", ASC)),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
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
			pageRequest: mustPageRequestOf(0, 1, mustNewOrder("price", DESC)),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
			},
			expectedResult: []*TestStruct{
				{Model: gorm.Model{ID: 4}, Code: "4", Price: 200},
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestStruct

			db.Debug().Clauses(test.pageRequest).Where(test.where).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}

			if !equalsArrays(products, test.expectedResult) {
				t.Fatalf("expected result to be %+v, got %+v", test.expectedResult, products)
			}
		})
	}
}

func TestPaGorminatorWithPreload(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestProduct
		pageRequest  *Pagination
		expectedPage *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
			},
		},
		"Paged 2/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 1, size: 1},
			expectedPage: &Pagination{
				page:          1,
				size:          1,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestProduct

			db.Clauses(test.pageRequest).Preload("Price").Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminatorWithPreloadAndWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestProduct
		pageRequest  *Pagination
		expectedPage *Pagination
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
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 4,
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
			expectedPage: &Pagination{
				page:          0,
				size:          2,
				totalElements: 4,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestProduct

			db.Clauses(test.pageRequest).Preload("Price").Where("code > 1").Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminatorWithJoins(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestProduct
		pageRequest  *Pagination
		expectedPage *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestProduct

			db.Clauses(test.pageRequest).Joins("Price").Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminatorWithJoins_WhereClause(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestProduct
		pageRequest  *Pagination
		where        any
		expectedPage *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
			},
			pageRequest: UnPaged(),
			where:       "1=1",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestProduct{
				{Code: "1", Price: TestPrice{Amount: 1, Currency: "EUR"}},
				{Code: "2", Price: TestPrice{Amount: 2, Currency: "EUR"}},
			},
			pageRequest: &Pagination{page: 0, size: 1},
			where:       "Price.amount > 1",
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 1,
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
			expectedPage: &Pagination{
				page:          0,
				size:          2,
				totalElements: 3,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products []*TestProduct

			db.Clauses(test.pageRequest).Joins("Price").Where(test.where).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminator_Table(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  *Pagination
		expectedPage *Pagination
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 2,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(1, 1),
			expectedPage: &Pagination{
				page:          1,
				size:          1,
				totalElements: 2,
			},
		},
		"Paged 0/2 items, size 2": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(0, 2),
			expectedPage: &Pagination{
				page:          0,
				size:          2,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var result map[string]any

			tx := db.Clauses(test.pageRequest).Table("test_structs").Find(&result)
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminator_TableWithWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  *Pagination
		where        string
		expectedPage *Pagination
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price < 100",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price > 100",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 0,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: UnPaged(),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 2},
				{Code: "3", Price: 100},
				{Code: "4", Price: 200},
			},
			pageRequest: mustPageRequestOf(0, 1),
			where:       "price > 50",
			expectedPage: &Pagination{
				page:          0,
				size:          1,
				totalElements: 2,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products map[string]any

			db.Clauses(test.pageRequest).Where(test.where).Table("test_structs").Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaGorminator_Distinct(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  *Pagination
		expectedPage *Pagination
	}{
		"UnPaged two items, same price": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
				{Code: "2", Price: 1},
			},
			pageRequest: UnPaged(),
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 1,
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
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 2,
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
			expectedPage: &Pagination{
				page:          0,
				size:          0,
				totalElements: 4,
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			db := setupDB(t)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			var products map[string]any

			db.Clauses(test.pageRequest).Distinct("price").Model(&TestStruct{}).Find(&products)

			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %+v, got %+v", test.expectedPage, test.pageRequest)
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

func mustPageRequestOf(page, size int, orders ...Order) *Pagination {
	toReturn, _ := PageRequest(page, size, orders...)
	return toReturn
}

func mustNewOrder(property string, direction Direction) Order {
	toReturn, _ := NewOrder(property, direction)
	return toReturn
}

func equalPageRequests(p1, p2 *Pagination) bool {
	return p1.page == p2.page &&
		p1.size == p2.size &&
		p1.totalElements == p2.totalElements &&
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

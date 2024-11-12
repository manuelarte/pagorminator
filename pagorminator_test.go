package pagorminator

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type TestStruct struct {
	gorm.Model
	Code  string
	Price uint
}

func TestPaginationScopeMetadata_NoWhere(t *testing.T) {
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			db := setupDb(t, name)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			// Read
			var products []*TestStruct

			db.Clauses(test.pageRequest).Find(&products) // find product with integer primary key
			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %d, got %d", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaginationScopeMetadata_Where(t *testing.T) {
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
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
				{Code: "1", Price: 100}, {Code: "2", Price: 200},
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
		t.Run(name, func(t *testing.T) {
			db := setupDb(t, name)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			// Read
			var products []*TestStruct

			db.Clauses(test.pageRequest).Where(test.where).Find(&products)
			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %d, got %d", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func setupDb(t *testing.T, name string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to connect database")
	}

	// Migrate the schema
	err = db.AutoMigrate(&TestStruct{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Use(PaGormMinator{})
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func mustPageRequestOf(page, size int) *Pagination {
	toReturn, _ := PageRequest(page, size)
	return toReturn
}

func equalPageRequests(p1, p2 *Pagination) bool {
	return p1.page == p2.page &&
		p1.size == p2.size &&
		p1.totalElements == p2.totalElements &&
		p1.GetTotalPages() == p2.GetTotalPages()
}

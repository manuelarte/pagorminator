package pagorminator

import (
	"fmt"
	"github.com/manuelarte/pagorminator/internal"
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
	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  PageRequest
		expectedPage PageRequest
	}{
		"UnPaged one item": {
			toMigrate: []*TestStruct{
				{Code: "1"},
			},
			pageRequest: UnPaged(),
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          0,
				TotalElements: 1,
				TotalPages:    1,
			},
		},
		"UnPaged several items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: UnPaged(),
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          0,
				TotalElements: 2,
				TotalPages:    1,
			},
		},
		"Paged 1/2 items": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
			},
			pageRequest: mustPageRequestOf(1, 1),
			expectedPage: &internal.PageRequestImpl{
				Page:          1,
				Size:          1,
				TotalElements: 2,
				TotalPages:    2,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			db := setupDb(t, name)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			// Read
			var products []*TestStruct

			db.Scopes(WithPagination(test.pageRequest)).Find(&products) // find product with integer primary key
			if !equalPageRequests(test.pageRequest, test.expectedPage) {
				t.Fatalf("expected page to be %d, got %d", test.expectedPage, test.pageRequest)
			}
		})
	}
}

func TestPaginationScopeMetadata_Where(t *testing.T) {
	tests := map[string]struct {
		toMigrate    []*TestStruct
		pageRequest  PageRequest
		where        string
		expectedPage PageRequest
	}{
		"UnPaged one item, not filtered": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price < 100",
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          0,
				TotalElements: 1,
				TotalPages:    1,
			},
		},
		"UnPaged one item, filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1},
			},
			pageRequest: UnPaged(),
			where:       "price > 100",
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          0,
				TotalElements: 0,
				TotalPages:    1,
			},
		},
		"UnPaged two items, one filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "100", Price: 100},
			},
			pageRequest: UnPaged(),
			where:       "price > 50",
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          0,
				TotalElements: 1,
				TotalPages:    1,
			},
		},
		"Paged four items, two filtered out": {
			toMigrate: []*TestStruct{
				{Code: "1", Price: 1}, {Code: "2", Price: 2},
				{Code: "1", Price: 100}, {Code: "2", Price: 200},
			},
			pageRequest: mustPageRequestOf(0, 1),
			where:       "price > 50",
			expectedPage: &internal.PageRequestImpl{
				Page:          0,
				Size:          1,
				TotalElements: 2,
				TotalPages:    2,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			db := setupDb(t, name)
			db.CreateInBatches(&test.toMigrate, len(test.toMigrate))

			// Read
			var products []*TestStruct

			db.Debug().Scopes(WithPagination(test.pageRequest)).Where(test.where).Find(&products) // find product with integer primary key
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

func mustPageRequestOf(page, size int) PageRequest {
	toReturn, _ := PageRequestOf(page, size)
	return toReturn
}

func equalPageRequests(p1, p2 PageRequest) bool {
	casted1 := p1.(*internal.PageRequestImpl)
	casted2 := p2.(*internal.PageRequestImpl)
	return casted1.Page == casted2.Page &&
		casted1.Size == casted2.Size &&
		casted1.TotalElements == casted2.TotalElements &&
		casted1.TotalPages == casted2.TotalPages
}

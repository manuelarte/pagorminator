package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator/cursorpagination"
)

type Product struct {
	gorm.Model

	Code  string
	Price uint
}

func (p Product) String() string {
	return fmt.Sprintf("Product{ID: %d, Code: %s, Price: %d}", p.ID, p.Code, p.Price)
}

func main() {
	db, err := gorm.Open(sqlite.Open("file:mem?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	_ = db.AutoMigrate(&Product{})
	db.Create(&Product{Code: "A", Price: 5})
	db.Create(&Product{Code: "A", Price: 3})
	db.Create(&Product{Code: "A", Price: 1})
	db.Create(&Product{Code: "B", Price: 5})
	db.Create(&Product{Code: "B", Price: 3})

	firstPage := cursorpagination.Must(
		2,
		cursorpagination.Asc("code", nil),
		cursorpagination.Desc("price", nil),
	)

	var firstPageProducts []*Product
	db.Clauses(firstPage).Find(&firstPageProducts)

	fmt.Printf("First page:\n")
	for _, product := range firstPageProducts {
		fmt.Printf("%s\n", product)
	}

	last := firstPageProducts[len(firstPageProducts)-1]
	nextPage := cursorpagination.Must(
		3,
		cursorpagination.Asc("code", last.Code),
		cursorpagination.Desc("price", last.Price),
	)

	var nextPageProducts []*Product
	db.Clauses(nextPage).Find(&nextPageProducts)

	fmt.Printf("Next page:\n")
	for _, product := range nextPageProducts {
		fmt.Printf("%s\n", product)
	}
}

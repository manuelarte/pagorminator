package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator/cursor"
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
	db.Create(&Product{Code: "A", Price: 100})
	db.Create(&Product{Code: "B", Price: 200})
	db.Create(&Product{Code: "C", Price: 300})
	db.Create(&Product{Code: "D", Price: 400})

	firstPage := cursor.MustPagination(2, cursor.Asc("id", nil))
	var firstPageProducts []*Product
	db.Clauses(firstPage).Find(&firstPageProducts)

	fmt.Printf("First page:\n")
	for _, product := range firstPageProducts {
		fmt.Printf("%s\n", product)
	}

	lastID := firstPageProducts[len(firstPageProducts)-1].ID
	nextPage := cursor.MustPagination(2, cursor.Asc("id", lastID))
	var nextPageProducts []*Product
	db.Clauses(nextPage).Find(&nextPageProducts)

	fmt.Printf("Next page:\n")
	for _, product := range nextPageProducts {
		fmt.Printf("%s\n", product)
	}
}

package main

import (
	"fmt"
	"github.com/manuelarte/pagorminator"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func (p Product) String() string {
	return fmt.Sprintf("Product{Code: %s, Price: %d}", p.Code, p.Price)
}

func main() {
	db, err := gorm.Open(sqlite.Open("file:mem?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Use(pagorminator.PaGormMinator{})
	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "D42", Price: 100})

	// Read
	var products []*Product
	pageRequest, _ := pagorminator.PageRequestOf(0, 1)
	db.Scopes(pagorminator.WithPagination(pageRequest)).First(&products)
	for _, product := range products {
		fmt.Printf("PageRequest: {Page: %d, Size: %d, TotalElements: %d, TotalPages: %d\n",
			pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
		println(product.String())
	}
}

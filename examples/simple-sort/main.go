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

	_ = db.Use(pagorminator.PaGormMinator{})
	_ = db.AutoMigrate(&Product{})
	db.Create(&Product{Code: "D42", Price: 100})
	db.Create(&Product{Code: "E42", Price: 200})
	fmt.Printf("2 products created\n")

	var products []*Product
	pageRequest, _ := pagorminator.PageRequest(0, 1, pagorminator.MustNewOrder("price", pagorminator.DESC))
	db.Clauses(pageRequest).Find(&products)

	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
	for _, product := range products {
		fmt.Printf("%s\n", product)
	}
}

package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator"
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

	_ = db.Use(pagorminator.PaGorminator{})
	_ = db.AutoMigrate(&Product{})

	migrateProducts := []*Product{
		{Code: "1", Price: 1},
		{Code: "10", Price: 10},
		{Code: "20", Price: 20},
		{Code: "21", Price: 21},
	}
	db.CreateInBatches(&migrateProducts, len(migrateProducts))
	fmt.Printf("%d products created\n", len(migrateProducts))

	pageRequest, _ := pagorminator.NewPageRequest(0, 1)
	var products []*Product
	db.Clauses(pageRequest).Where("price > 10").Find(&products)
	fmt.Printf("Query: Products (Page: %d, Size: %d) with '%s'\n", pageRequest.GetPage(), pageRequest.GetSize(), "price > 10")

	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
}

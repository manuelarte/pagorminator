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

	migrateProducts := []*Product{{Code: "1", Price: 1}, {Code: "10", Price: 10},
		{Code: "20", Price: 20}, {Code: "21", Price: 21}}
	db.CreateInBatches(&migrateProducts, len(migrateProducts))

	var products []*Product
	pageRequest, _ := pagorminator.PageRequest(0, 1)
	db.Clauses(pageRequest).Where("price > 10").Find(&products)

	fmt.Printf("PageRequest: {Page: %d, Size: %d, TotalElements: %d, TotalPages: %d\n",
		pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
}

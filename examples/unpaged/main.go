package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator"
	"github.com/manuelarte/pagorminator/pagepagination"
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

	var products []*Product
	pagination := pagepagination.UnPaged()
	db.Clauses(pagination).Find(&products)

	totalElements, _ := pagination.GetTotalElements()
	fmt.Printf("Unpaged(TotalElements: %d, TotalPages: %d)\n",
		totalElements, pagination.GetTotalPages())
}

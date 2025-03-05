package main

import (
	"fmt"
	"strconv"

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
	length := 10
	for i := 0; i < length; i++ {
		db.Create(&Product{Code: strconv.Itoa(i), Price: uint(i)})
	}

	fmt.Printf("%s product created\n", length)

	var products []*Product
	pageRequest, _ := pagorminator.PageRequest(0, 5)
	db.Clauses(pageRequest).Find(&products)

	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}

	pageRequest, _ = pagorminator.PageRequest(1, 5)
	db.Clauses(pageRequest).Find(&products)
	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), pageRequest.GetTotalElements(), pageRequest.GetTotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}
}

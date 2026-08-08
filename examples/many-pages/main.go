package main

import (
	"fmt"
	"strconv"

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
	length := 10
	for i := 0; i < length; i++ {
		db.Create(&Product{Code: strconv.Itoa(i), Price: uint(i)})
	}

	fmt.Printf("%s product created\n", length)

	var products []*Product
	pageRequest, _ := pagepagination.New(0, 5)
	db.Clauses(pageRequest).Find(&products)

	totalElements1, _ := pageRequest.GetTotalElements()
	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), totalElements1, pageRequest.GetTotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}

	pageRequest, _ = pagepagination.New(1, 5)
	db.Clauses(pageRequest).Find(&products)
	totalElements2, _ := pageRequest.GetTotalElements()
	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.GetPage(), pageRequest.GetSize(), totalElements2, pageRequest.GetTotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}
}

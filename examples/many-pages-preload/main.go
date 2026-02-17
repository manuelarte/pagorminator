package main

import (
	"fmt"
	"strconv"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/manuelarte/pagorminator"
	"github.com/manuelarte/pagorminator/pagination"
)

type Product struct {
	gorm.Model

	Code  string
	Price Price
}

type Price struct {
	gorm.Model

	Unit      uint
	Currency  string
	ProductID uint
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
	_ = db.AutoMigrate(&Product{}, &Price{})
	length := 10
	for i := 0; i < length; i++ {
		errCreatingProduct := db.Create(&Product{Code: strconv.Itoa(i), Price: Price{Unit: uint(i), Currency: "EUR"}})
		if errCreatingProduct.Error != nil {
			panic(errCreatingProduct.Error)
		}
	}

	fmt.Printf("%d product created\n", length)

	var products []*Product
	pageRequest, _ := pagination.New(0, 5)
	txErr := db.Debug().Clauses(pageRequest).Preload("Price").Find(&products).Error
	if txErr != nil {
		panic(txErr)
	}

	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.Page(), pageRequest.Size(), pageRequest.TotalElements(), pageRequest.TotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}

	pageRequest, _ = pagination.New(1, 5)
	db.Clauses(pageRequest).Find(&products)
	fmt.Printf("PageRequest result:(Page: %d, Size: %d, TotalElements: %d, TotalPages: %d)\n",
		pageRequest.Page(), pageRequest.Size(), pageRequest.TotalElements(), pageRequest.TotalPages())
	for _, product := range products {
		fmt.Printf("\t Product: %s\n", product)
	}
}

package tests

import "gorm.io/gorm"

type TestStruct struct {
	gorm.Model

	Code  string
	Price uint
}

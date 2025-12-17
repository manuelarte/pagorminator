module examples

go 1.24.0

toolchain go1.24.7

replace github.com/manuelarte/pagorminator v0.1.0 => ../

require (
	github.com/manuelarte/pagorminator v0.1.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	golang.org/x/text v0.30.0 // indirect
)

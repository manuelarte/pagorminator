module examples

go 1.20

replace github.com/manuelarte/pagorminator v0.0.1-rc5 => ../

require (
	github.com/manuelarte/pagorminator v0.0.1-rc5
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.30.0
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.24 // indirect
	golang.org/x/text v0.22.0 // indirect
)

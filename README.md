[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
# 📃 pagorminator

Gorm plugin to add pagination to your select queries

## 😍 How to install it

> go get -u -v github.com/manuelarte/pagorminator

## 🎯 How to use it

```go
DB.Use(pagorminator.PaGormMinator{})
var products []*Products
pageRequest, err := pagorminator.PageRequest(0, 10)
DB.Clauses(pageRequest).First(&products)
```

The plugin will calculate the total amount of elements so then the fields `total amounts` and `total pages` can be used too.

## 🎓 Examples

- [Simple](./examples/simple/main.go)

Simple query with no filters (where clause)

- [Filter](./examples/filter/main.go) 

Using where to filter

## 🔗 Contact

- 📧 manueldoncelmartos@gmail.com

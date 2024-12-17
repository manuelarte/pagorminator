[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/manuelarte/pagorminator)](https://goreportcard.com/report/github.com/manuelarte/pagorminator)
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
The pagination starts with the index 0.

## 🎓 Examples

- [Simple](./examples/simple/main.go)

Simple query with no filters (where clause)

- [Filter](./examples/filter/main.go) 

Using where to filter

- [Unpaged](./examples/unpaged/main.go)

Unpaged query

## 🔗 Contact

- 📧 manueldoncelmartos@gmail.com

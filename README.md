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
The pagination starts at index 0. So if the total pages is 6, then the pagination index goes from 0 to 5.

## 🎓 Examples

- [Simple](./examples/simple/main.go)

Simple query with no filters (where clause)

- [Many Pages](./examples/many-pages/main.go)

Simple query with no filters (where clause), many pages

- [Filter](./examples/filter/main.go) 

Using where to filter

- [Unpaged](./examples/unpaged/main.go)

Unpaged query

- [Many Pages With Preload](./examples/many-pages-preload/main.go)

Example using Preload

## 🔗 Contact

- 📧 manueldoncelmartos@gmail.com

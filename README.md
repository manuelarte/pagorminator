# 📃 PaGorminator

[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/manuelarte/pagorminator)](https://goreportcard.com/report/github.com/manuelarte/pagorminator)
[![Go Reference](https://pkg.go.dev/badge/github.com/manuelarte/pagorminator.svg)](https://pkg.go.dev/github.com/manuelarte/pagorminator)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10813/badge)](https://www.bestpractices.dev/projects/10813)
![version](https://img.shields.io/github/v/release/manuelarte/pagorminator)

Gorm plugin to add **Pagination** to your select queries

<img src="pagorminator_logo.png" alt="logo" width="256" height="256"/>

## ⬇️ How to install it

> go get -u -v github.com/manuelarte/pagorminator

## 🎯 How to use it

```go
var DB *gorm.DB
DB.Use(pagorminator.PaGorminator{})
var products []*Products
// Without sorting
pageRequest, err := pagorminator.PageRequest(0, 10)
// With sorting
pageRequest2, err := pagorminator.PageRequest(0, 10, pagorminator.MustOrder("id", pagorminator.DESC))
DB.Clauses(pageRequest).First(&products)
```

The pagination struct contains the following data:

+ `page`: page number, e.g. `0`
+ `size`: page size, e.g. `10`
+ `sort`: to apply sorting, e.g. `id,asc`

**The plugin will calculate the total amount of elements**.
The pagination instance provides a `GetTotalElements()` and `GetTotalPages()` methods to retrieve the total amount of elements.
The pagination starts at index `0`, e.g., if the total pages is `6`, then the pagination index goes from `0` to `5`.

## Examples

Check the examples in the [./examples](./examples) folder

# 📃 PaGORMinator

[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/manuelarte/pagorminator)](https://goreportcard.com/report/github.com/manuelarte/pagorminator)
[![Go Reference](https://pkg.go.dev/badge/github.com/manuelarte/pagorminator.svg)](https://pkg.go.dev/github.com/manuelarte/pagorminator)
![version](https://img.shields.io/github/v/release/manuelarte/pagorminator)

Gorm plugin to add **Pagination** to your select queries

<img src="pagorminator_logo.png" alt="logo" width="256" height="256"/>

## ⬇️ How to install it

> go get -u -v github.com/manuelarte/pagorminator

## 🎯 How to use it

```go
var DB *gorm.DB
DB.Use(pagorminator.PaGormMinator{})
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
Then the pagination instance  provides a `GetTotalElements()` and `GetTotalPages()` methods to be used.
The pagination starts at index `0`. So if the total pages is `6`, then the pagination index goes from `0` to `5`.

## 🎓 Examples

+ [Simple](./examples/simple/main.go): Simple query with no filters (no WHERE clause).

+ [Simple Sort](./examples/simple-sort/main.go): Simple query with sorting and no filters (no WHERE clause).

+ [Simple Table](./examples/simple-table/main.go): Simple query in which there is no model, but using `Table()`.

+ [Many Pages](./examples/many-pages/main.go): Simple query with no filters (no WHERE clause), many pages.

+ [Filter](./examples/filter/main.go): Using WHERE to filter.

+ [Unpaged](./examples/unpaged/main.go): Unpaged query (pagination with no pagination).

+ [Many Pages With Preload](./examples/many-pages-preload/main.go): Example using Preload.

[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
# 📃 pagorminator

Gorm plugin to add pagination to your select queries

## 😍 How to install it

> go get github.com/manuelarte/pagorminator

## 🎯 How to use it

```go
var products []*Products
// give me the first 10 products
pageRequest := pagorminator.PageRequestOf(0, 10)
db.Scopes(pagorminator.WithPagination(&pageRequest)).Find(&products)
```

The plugin will populate the page request variable will the `total amounts` and `total pages` fields.

## 🎓 Examples

- [Simple](./examples/simple/main.go)

Simple query with no filters (where clause)

- Filtered 

Using where to filter

## Contact

- 📧 manueldoncelmartos@gmail.com

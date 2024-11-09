[![Go](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/go.yml)
# 📃 pagorminator

Gorm plugin to add pagination to your select queries

## 😍 How to install it

> go get github.com/manuelarte/pagorminator

## 🎯 How to use it

```go
var products []*Products
// give me the first 10 products
db.Scopes(pagorminator.WithPagination(pagorminator.PageRequestOf(0, 10))).Find(&products)
```

## 🎓 Examples

- [Simple](./examples/simple/main.go)

Simple query with no filters (where clause)

- Filtered 

Using where to filter

## Contact

- 📧 manueldoncelmartos@gmail.com

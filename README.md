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

```bash
go get -u -v github.com/manuelarte/pagorminator
```

## 🎯 How to use it

### Basic Usage

```go
// Initialize GORM with PaGorminator plugin
db, err := gorm.Open(sqlite.Open("file:mem?mode=memory&cache=shared"), &gorm.Config{})
if err != nil {
    panic("failed to connect database")
}
db.Use(pagorminator.PaGorminator{})

// Create a page request (page 0, size 10)
pageRequest, err := pagorminator.PageRequest(0, 10)
if err != nil {
    // Handle error
}

// Apply pagination to your query
var products []*Product
db.Clauses(pageRequest).Find(&products)

// Access pagination information
fmt.Printf("Total elements: %d\n", pageRequest.GetTotalElements())
fmt.Printf("Total pages: %d\n", pageRequest.GetTotalPages())
```

### Pagination Parameters

The pagination struct contains the following data:

+ `page`: page number, e.g. `0` (zero-based indexing)
+ `size`: page size, e.g. `10`
+ `sort`: to apply sorting, e.g. `id desc`

**The plugin will automatically calculate the total amount of elements**.
The pagination instance provides `GetTotalElements()` and `GetTotalPages()` methods to retrieve the total counts.
The pagination starts at index `0`, e.g., if the total pages is `6`, then the pagination index goes from `0` to `5`.

## Features

### Sorting

You can add sorting to your pagination request:

```go
// Single sort criterion
pageRequest, err := pagorminator.PageRequest(0, 10, 
    pagorminator.MustOrder("id", pagorminator.DESC))

// Multiple sort criteria
pageRequest, err := pagorminator.PageRequest(0, 10, 
    pagorminator.MustOrder("name", pagorminator.ASC),
    pagorminator.MustOrder("id", pagorminator.DESC))
```

### Unpaged Requests

If you want to retrieve all records without pagination:

```go
// Create an unpaged request
unpaged := pagorminator.UnPaged()
db.Clauses(unpaged).Find(&products)
```

#### Debug Mode

You can enable debug mode to see the SQL queries:

```go
// Enable debug mode
db.Use(pagorminator.PaGorminator{Debug: true})
```

## Examples

Check the examples in the [./examples](./examples) folder for more detailed usage patterns.

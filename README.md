# 📃 PaGorminator

[![CI](https://github.com/manuelarte/pagorminator/actions/workflows/ci.yml/badge.svg)](https://github.com/manuelarte/pagorminator/actions/workflows/ci.yml)
![coverage](https://raw.githubusercontent.com/manuelarte/pagorminator/badges/.badges/main/coverage.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/manuelarte/pagorminator.svg)](https://pkg.go.dev/github.com/manuelarte/pagorminator)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/10813/badge)](https://www.bestpractices.dev/projects/10813)
![version](https://img.shields.io/github/v/release/manuelarte/pagorminator)

Gorm plugin to add **Pagination** to your select queries

<img src="pagorminator_logo.png" alt="logo" width="256" height="256"/>

## ⬇️ How to install it

```bash
go get -u -v github.com/manuelarte/pagorminator
```

## 🚀 Features

### Page Pagination

#### Page Pagination Basic Usage

```go
// Initialize GORM with PaGorminator plugin
db, err := gorm.Open(sqlite.Open("file:mem?mode=memory&cache=shared"), &gorm.Config{})
if err != nil {
    panic("failed to connect database")
}
db.Use(pagorminator.PaGorminator{})

// Create a page request (page 0, size 10)
pageRequest, err := pagepagination.New(0, 10)
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

#### Page Pagination Parameters

The pagination needs the following data:

+ `page`: page number, e.g. `0` (zero-based indexing)
+ `size`: page size, e.g. `10`
+ `sort`: to apply sorting, e.g. `id desc`

**The plugin will automatically calculate the total number of elements**.
The pagination instance provides `GetTotalElements()` and `GetTotalPages()` methods to retrieve the total counts.
The pagination starts at index `0`, e.g., if the total pages is `6`, then the pagination index goes from `0` to `5`.

#### Sorting

You can add sorting to your pagination request:

```go
// Single sort criterion
pageRequest, err := pagepagination.New(0, 10, pagorminator.Desc("id"))

// Multiple sort criteria
pageRequest, err := page.NewPagination(
  0,
  10,
  pagorminator.Asc("name"),
  pagorminator.Desc("price")
)
```

##### Unpaged Requests

If you want to retrieve all records without pagination:

```go
// Create an unpaged request
unpaged := pagepagination.UnPaged()
db.Clauses(unpaged).Find(&products)
```

### Cursor Pagination

#### Cursor Pagination Basic Usage

```go
// Initialize GORM with PaGorminator plugin
db, err := gorm.Open(sqlite.Open("file:mem?mode=memory&cache=shared"), &gorm.Config{})
if err != nil {
panic("failed to connect database")
}
db.Use(pagorminator.PaGorminator{})

// first page with cursor ordering
firstPage := cursorpagination.Must(10, cursor.Asc("id"))
db.Clauses(firstPage).Find(&products)

// next page after id 100
nextPage := cursorpagination.Must(10, cursor.Asc("id", 100))
db.Clauses(nextPage).Find(&products)

// multi-column cursor pagination
multiColumnPage := cursorpagination.Must(
  10,
  cursor.Asc("code", "A42"),
  cursor.Desc("price", 100),
)
db.Clauses(multiColumnPage).Find(&products)
```

#### Cursor Pagination Parameters

The pagination needs the following data:

+ `size`: page size, e.g. `10`
+ `cursors`: the cursor values to apply, the column, value and sort direction.

**The plugin will automatically calculate the total number of elements**.
The pagination instance provides `GetTotalElements()` method to retrieve the total counts.

### Debug Mode

You can enable debug mode to see the SQL queries:

```go
// Enable debug mode
db.Use(pagorminator.PaGorminator{Debug: true})
```

## 🎓Examples

Check the examples in the [./examples](./examples) folder for more detailed usage patterns.

## ❓FAQ

### Is it protected against SQL injection?

Yes, or at least as protected as the GORM framework protects. The plugin/library uses GORM's parameterized
queries to prevent SQL injection. There are several tests that cover this scenario,
including the `TestSQLInjection` test in the `pagorminator_test.go` file.

### Does it work for my database?

The plugin is using internal GORM methods to generate the SQL queries. So in theory, it is supported
if there is a GORM dialect for your database. Nevertheless, we have tests in [./tests/pagination.test](./tests/pagination_test.go)
to ensure that the plugin/library works for certain databases.

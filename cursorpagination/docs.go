// Package cursorpagination contains the implementation of cursor-based
// pagination for GORM. It provides a way to paginate through
// database records using cursors, which can be more efficient
// than traditional offset-based pagination, especially for large datasets.
// The package includes the `Pagination` struct,
// which allows you to specify the size of the page and
// the cursors to use for pagination.
package cursorpagination

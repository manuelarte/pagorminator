package sort

import (
	"fmt"
	"strings"
)

var (
	_ Order = new(Asc)
	_ Order = new(Desc)
)

type (
	// Order represents a sort order.
	Order interface {
		gormString() string
	}

	Asc string

	Desc string

	// Sort represents a slice of orders.
	Sort []Order
)

func (a Asc) gormString() string {
	return fmt.Sprintf("%s asc", a)
}

func (d Desc) gormString() string {
	return fmt.Sprintf("%s desc", d)
}

// New sort (slices of [Order]).
func New(orders ...Order) Sort {
	return orders
}

// Unsorted no sorting.
func Unsorted() Sort {
	return Sort{}
}

func (s Sort) String() string {
	orderStrings := make([]string, len(s))
	for i, order := range s {
		orderStrings[i] = order.gormString()
	}

	return strings.Join(orderStrings, ", ")
}

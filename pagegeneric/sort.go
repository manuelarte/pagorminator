package pagegeneric

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
		GormString() string
		order()
	}

	Sort []Order

	// Asc is ascending order.
	Asc string

	// Desc is descending order.
	Desc string
)

func (a Asc) GormString() string {
	return fmt.Sprintf("%s ASC", a)
}

func (a Asc) order() {}

func (d Desc) GormString() string {
	return fmt.Sprintf("%s DESC", d)
}

func (d Desc) order() {}

// NewSort Creates sort (slices of [Order]).
func NewSort(orders ...Order) Sort {
	return orders
}

// Unsorted no sorting.
func Unsorted() Sort {
	return Sort{}
}

func (s Sort) String() string {
	orderStrings := make([]string, len(s))
	for i, order := range s {
		orderStrings[i] = order.GormString()
	}

	return strings.Join(orderStrings, ", ")
}

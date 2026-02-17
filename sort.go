package pagorminator

import (
	"fmt"
	"strings"
)

var (
	_ Order = new(Asc)
	_ Order = new(Desc)
)

type (
	Order interface {
		gormString() string
	}

	Sort []Order

	Asc string

	Desc string
)

func (a Asc) gormString() string {
	return fmt.Sprintf("%s ASC", a)
}

func (d Desc) gormString() string {
	return fmt.Sprintf("%s DESC", d)
}

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
		orderStrings[i] = order.gormString()
	}

	return strings.Join(orderStrings, ", ")
}

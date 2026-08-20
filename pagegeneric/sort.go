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
		Column() string
		GormString() string
		order()
	}

	// Sort represents a collection of Order.
	Sort []Order

	// Asc is ascending order.
	Asc string

	// Desc is descending order.
	Desc string
)

// Column returns the column name of the order.
func (a Asc) Column() string {
	return string(a)
}

// GormString returns the string representation of the order for gorm.
func (a Asc) GormString() string {
	return fmt.Sprintf("%s ASC", a)
}

func (a Asc) order() {}

// Column returns the column name of the order.
func (d Desc) Column() string {
	return string(d)
}

// GormString returns the string representation of the order for gorm.
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

// String returns the string representation of the sort.
func (s Sort) String() string {
	orderStrings := make([]string, len(s))
	for i, order := range s {
		orderStrings[i] = order.GormString()
	}

	return strings.Join(orderStrings, ", ")
}

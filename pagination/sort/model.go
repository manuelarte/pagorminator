package sort

import (
	"fmt"
	"strings"
)

type Direction string

const (
	ASC  Direction = "asc"
	DESC Direction = "desc"
)

type (
	// Order represents a sort order.
	Order struct {
		property  string
		direction Direction
	}

	// Sort represents a slice of orders.
	Sort []Order
)

// NewOrder Creates new order based on a property and a direction.
func NewOrder(property string, direction Direction) (Order, error) {
	if property == "" {
		return Order{}, ErrOrderPropertyIsEmpty
	}

	if direction != "" && direction != ASC && direction != DESC {
		return Order{}, OrderDirectionNotValidError{Direction: direction}
	}

	return Order{
		property:  property,
		direction: direction,
	}, nil
}

// MustOrder Creates a new order based on a property and a direction, or panic.
func MustOrder(property string, direction Direction) Order {
	order, err := NewOrder(property, direction)
	if err != nil {
		panic(err)
	}

	return order
}

func (o Order) String() string {
	if o.direction == "" {
		return o.property
	}

	return fmt.Sprintf("%s %s", o.property, o.direction)
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
		orderStrings[i] = order.String()
	}

	return strings.Join(orderStrings, ", ")
}

package pagorminator

import (
	"fmt"
	"strings"
)

func NewSort(orders ...Order) Sort {
	return orders
}

func Unsorted() Sort {
	return Sort{}
}

type Sort []Order

func (s Sort) String() string {
	orderStrings := make([]string, len(s))
	for i, order := range s {
		orderStrings[i] = order.String()
	}
	return strings.Join(orderStrings, ", ")
}

type Direction string

const (
	ASC  Direction = "asc"
	DESC Direction = "desc"
)

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

func MustNewOrder(property string, direction Direction) Order {
	order, err := NewOrder(property, direction)
	if err != nil {
		panic(err)
	}
	return order
}

type Order struct {
	property  string
	direction Direction
}

func (o Order) String() string {
	if o.direction == "" {
		return o.property
	}
	return fmt.Sprintf("%s %s", o.property, o.direction)
}

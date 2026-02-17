package sort

import (
	"errors"
	"testing"
)

func TestOrderNewOrder(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		property      string
		direction     Direction
		expectedOrder Order
		expectedErr   error
	}{
		"order with valid property and asc direction": {
			property:      "name",
			direction:     "asc",
			expectedOrder: MustOrder("name", ASC),
		},
		"order with valid property and desc direction": {
			property:      "name",
			direction:     "desc",
			expectedOrder: MustOrder("name", DESC),
		},
		"order with valid property and empty direction": {
			property:      "name",
			direction:     "",
			expectedOrder: MustOrder("name", ""),
		},
		"order with valid property and invalid direction (uppercase)": {
			property:    "name",
			direction:   "DESC",
			expectedErr: OrderDirectionNotValidError{Direction: "DESC"},
		},
		"order with valid property and invalid direction (not related)": {
			property:    "name",
			direction:   "hello",
			expectedErr: OrderDirectionNotValidError{Direction: "hello"},
		},
		"order with invalid property": {
			property:    "",
			direction:   "asc",
			expectedErr: ErrOrderPropertyIsEmpty,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			order, err := NewOrder(test.property, test.direction)
			if err != nil {
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("got err %v, expected %v", err, test.expectedErr)
				}
			}

			if order != test.expectedOrder {
				t.Errorf("got order %v, expected %v", order, test.expectedOrder)
			}
		})
	}
}

func TestOrderString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		order    Order
		expected string
	}{
		"order with asc direction": {
			order:    MustOrder("name", ASC),
			expected: "name asc",
		},
		"order with desc direction": {
			order:    MustOrder("name", DESC),
			expected: "name desc",
		},
		"order without direction": {
			order:    MustOrder("name", ""),
			expected: "name",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := test.order.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

func TestSortString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sort     Sort
		expected string
	}{
		"order with asc direction": {
			sort: Sort([]Order{
				MustOrder("name", ASC),
				MustOrder("surname", DESC),
			}),
			expected: "name asc, surname desc",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := test.sort.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

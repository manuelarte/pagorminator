package pagorminator_test

import (
	"errors"
	"testing"

	"github.com/manuelarte/pagorminator"
)

func TestOrder_NewOrder(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		property      string
		direction     pagorminator.Direction
		expectedOrder pagorminator.Order
		expectedErr   error
	}{
		"order with valid property and asc direction": {
			property:      "name",
			direction:     "asc",
			expectedOrder: pagorminator.MustOrder("name", pagorminator.ASC),
		},
		"order with valid property and desc direction": {
			property:      "name",
			direction:     "desc",
			expectedOrder: pagorminator.MustOrder("name", pagorminator.DESC),
		},
		"order with valid property and empty direction": {
			property:      "name",
			direction:     "",
			expectedOrder: pagorminator.MustOrder("name", ""),
		},
		"order with valid property and invalid direction (uppercase)": {
			property:    "name",
			direction:   "DESC",
			expectedErr: pagorminator.OrderDirectionNotValidError{Direction: "DESC"},
		},
		"order with valid property and invalid direction (not related)": {
			property:    "name",
			direction:   "hello",
			expectedErr: pagorminator.OrderDirectionNotValidError{Direction: "hello"},
		},
		"order with invalid property": {
			property:    "",
			direction:   "asc",
			expectedErr: pagorminator.ErrOrderPropertyIsEmpty,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			order, err := pagorminator.NewOrder(test.property, test.direction)
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

func TestOrder_String(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		order    pagorminator.Order
		expected string
	}{
		"order with asc direction": {
			order:    pagorminator.MustOrder("name", pagorminator.ASC),
			expected: "name asc",
		},
		"order with desc direction": {
			order:    pagorminator.MustOrder("name", pagorminator.DESC),
			expected: "name desc",
		},
		"order without direction": {
			order:    pagorminator.MustOrder("name", ""),
			expected: "name",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := test.order.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

func TestSort_String(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		sort     pagorminator.Sort
		expected string
	}{
		"order with asc direction": {
			sort: pagorminator.Sort([]pagorminator.Order{
				pagorminator.MustOrder("name", pagorminator.ASC),
				pagorminator.MustOrder("surname", pagorminator.DESC),
			}),
			expected: "name asc, surname desc",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := test.sort.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

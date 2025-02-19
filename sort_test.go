package pagorminator

import (
	"errors"
	"testing"
)

func TestOrder_NewOrder(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		property      string
		direction     Direction
		expectedOrder Order
		expectedErr   error
	}{
		"order with valid property and asc direction": {
			property:  "name",
			direction: "asc",
			expectedOrder: Order{
				property:  "name",
				direction: ASC,
			},
		},
		"order with valid property and desc direction": {
			property:  "name",
			direction: "desc",
			expectedOrder: Order{
				property:  "name",
				direction: DESC,
			},
		},
		"order with valid property and empty direction": {
			property:  "name",
			direction: "",
			expectedOrder: Order{
				property:  "name",
				direction: "",
			},
		},
		"order with valid property and invalid direction (uppercase)": {
			property:    "name",
			direction:   "DESC",
			expectedErr: ErrOrderDirectionNotValid{Direction: "DESC"},
		},
		"order with valid property and invalid direction (not related)": {
			property:    "name",
			direction:   "hello",
			expectedErr: ErrOrderDirectionNotValid{Direction: "hello"},
		},
		"order with invalid property": {
			property:    "",
			direction:   "asc",
			expectedErr: ErrOrderPropertyIsEmpty,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
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

func TestOrder_String(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		order    Order
		expected string
	}{
		"order with asc direction": {
			order: Order{
				property:  "name",
				direction: ASC,
			},
			expected: "name asc",
		},
		"order with desc direction": {
			order: Order{
				property:  "name",
				direction: DESC,
			},
			expected: "name desc",
		},
		"order without direction": {
			order: Order{
				property:  "name",
				direction: DESC,
			},
			expected: "name",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.order.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

func TestSort_String(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		sort     Sort
		expected string
	}{
		"order with asc direction": {
			sort: Sort([]Order{
				{
					property:  "name",
					direction: ASC,
				},
				{
					property:  "surname",
					direction: DESC,
				},
			}),
			expected: "name asc, surname desc",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if got := test.sort.String(); got != test.expected {
				t.Errorf("got %q, want %q", got, test.expected)
			}
		})
	}
}

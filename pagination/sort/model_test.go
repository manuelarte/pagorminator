package sort

import (
	"testing"
)

func TestOrderString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		order Order
		want  string
	}{
		"order with asc direction": {
			order: Asc("name"),
			want:  "name asc",
		},
		"order with desc direction": {
			order: Desc("name"),
			want:  "name desc",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.order.gormString()
			if got != test.want {
				t.Errorf("test.order.gormString() = %q, want %q", got, test.want)
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
				Asc("name"),
				Desc("surname"),
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

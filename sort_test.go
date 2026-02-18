package pagorminator

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
			want:  "name ASC",
		},
		"order with desc direction": {
			order: Desc("name"),
			want:  "name DESC",
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
		sort Sort
		want string
	}{
		"order with asc direction": {
			sort: Sort([]Order{
				Asc("name"),
				Desc("surname"),
			}),
			want: "name ASC, surname DESC",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.sort.String()
			if got != test.want {
				t.Errorf("test.sort.String() = %q, want %q", got, test.want)
			}
		})
	}
}

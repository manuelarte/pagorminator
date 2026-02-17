package pagination

import (
	"testing"
)

func TestUnPaged(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		page     Page
		size     Size
		expected bool
	}{
		"page 0 size 0": {
			page:     0,
			size:     0,
			expected: true,
		},
		"page zero size not zero": {
			page:     0,
			size:     1,
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			page, err := New(test.page, test.size)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}

			got := page.IsUnPaged()
			if got != test.expected {
				t.Errorf("IsUnPaged() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		totalElements uint64
		size          Size
		want          uint
	}{
		"totalElements lower than size": {
			totalElements: 2,
			size:          4,
			want:          1,
		},
		"totalElements greater and not divisible by size": {
			totalElements: 3,
			size:          2,
			want:          2,
		},
		"totalElements greater and divisible by size": {
			totalElements: 4,
			size:          2,
			want:          2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := calculateTotalPages(test.totalElements, test.size)
			if got != test.want {
				t.Errorf("calculateTotalPages(%d, %d) = %v, want %v", test.totalElements, test.size, got, test.want)
			}
		})
	}
}

func TestSetTotalElements(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		totalElements uint64
		expectedErr   error
	}{
		"positive totalElements": {
			totalElements: 2,
		},
		"0 totalElements": {
			totalElements: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Pagination{}

			p.SetTotalElements(test.totalElements)

			got := p.TotalElements()
			if got != test.totalElements {
				t.Errorf("p.TotalElements() = %v, want %v", got, test.totalElements)
			}
		})
	}
}

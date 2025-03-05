package pagorminator_test

import (
	"math"
	"testing"

	"github.com/manuelarte/pagorminator"
)

func TestPagination_UnPaged(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		page     int
		size     int
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
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			page, err := pagorminator.PageRequest(test.page, test.size)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if page.IsUnPaged() != test.expected {
				t.Errorf("IsUnPaged() expected %v, got %v", test.expected, page.IsUnPaged())
			}
		})
	}
}

func TestPagination_CalculateTotalPages(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		totalElements int64
		size          int
		expected      int
	}{
		"totalElements lower than size": {
			totalElements: 2,
			size:          4,
			expected:      1,
		},
		"totalElements greater and not divisible by size": {
			totalElements: 3,
			size:          2,
			expected:      2,
		},
		"totalElements greater and divisible by size": {
			totalElements: 4,
			size:          2,
			expected:      2,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := calculateTotalPages(test.totalElements, test.size)
			if actual != test.expected {
				t.Errorf("totalPages expected %v, got %v", test.expected, actual)
			}
		})
	}
}

func calculateTotalPages(totalElements int64, size int) int {
	return int(math.Ceil(float64(totalElements) / float64(size)))
}

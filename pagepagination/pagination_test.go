package pagepagination

import (
	"errors"
	"fmt"
	"testing"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

func ExampleNew() {
	page, err := New(1, 10)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Page: %d, Size: %d\n", page.GetPage(), page.GetSize())
	// Output: Page: 1, Size: 10
}

func ExampleMust() {
	page := Must(1, 10)

	fmt.Printf("Page: %d, Size: %d\n", page.GetPage(), page.GetSize())
	// Output: Page: 1, Size: 10
}

func TestUnPaged(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			page, err := New(test.page, test.size)
			if err != nil {
				t.Errorf("NewPagination(%d, %d) = %s, unexpected error", test.page, test.size, err)
			}

			if page.IsUnPaged() != test.expected {
				t.Errorf("IsUnPaged() expected %v, got %v", test.expected, page.IsUnPaged())
			}
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		totalElements int64
		size          int
		want          int
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
				t.Errorf("calculateTotalPages = %v, want %v", got, test.want)
			}
		})
	}
}

func TestSetTotalElements(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		totalElements int64
		expectedErr   error
	}{
		"positive totalElements": {
			totalElements: 2,
		},
		"0 totalElements": {
			totalElements: 0,
		},
		"negative totalElements": {
			totalElements: -1,
			expectedErr:   pagegeneric.TotalElementsNotValidError{TotalElements: -1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Pagination{}

			actualErr := p.SetTotalElements(test.totalElements)
			if !errors.Is(actualErr, test.expectedErr) {
				t.Errorf("expected: %v, got: %v", test.expectedErr, actualErr)
				t.Fail()
			}
		})
	}
}

func TestNext(t *testing.T) {
	t.Parallel()

	t.Run("fails when total elements are not set", func(t *testing.T) {
		t.Parallel()

		p := Must(0, 2, pagegeneric.Asc("id"))

		next, err := p.Next()
		if !errors.Is(err, pagegeneric.ErrTotalElementsNotSet) {
			t.Errorf("expected error %v, got %v", pagegeneric.ErrTotalElementsNotSet, err)
		}

		if next != nil {
			t.Errorf("expected nil next page, got %#v", next)
		}
	})

	t.Run("fails when already at the last page", func(t *testing.T) {
		t.Parallel()

		p := Must(2, 2)
		if err := p.SetTotalElements(5); err != nil {
			t.Errorf("unexpected error setting total elements: %v", err)
		}

		next, err := p.Next()
		if !errors.Is(err, ErrNoNextPage) {
			t.Errorf("expected error %v, got %v", ErrNoNextPage, err)
		}

		if next != nil {
			t.Errorf("expected nil next page, got %#v", next)
		}
	})

	t.Run("returns next page and preserves configuration", func(t *testing.T) {
		t.Parallel()

		p := Must(0, 2, pagegeneric.Asc("id"), pagegeneric.Desc("price"))
		if err := p.SetTotalElements(5); err != nil {
			t.Errorf("unexpected error setting total elements: %v", err)
		}

		next, err := p.Next()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if next.GetPage() != 1 {
			t.Errorf("expected page 1, got %d", next.GetPage())
		}

		if next.GetSize() != 2 {
			t.Errorf("expected size 2, got %d", next.GetSize())
		}

		if !next.IsSort() {
			t.Errorf("expected next page to keep sort")
		}

		if got, want := len(next.GetSort()), 2; got != want {
			t.Errorf("expected %d sort constraints, got %d", want, got)
		}

		if next.IsTotalElementsSet() {
			t.Errorf("expected next page total elements to not be set")
		}

		if got := next.GetTotalElements(); got != 0 {
			t.Errorf("next.GetTotalElements() = %d, expected total elements to not to be set", got)
		}
	})
}

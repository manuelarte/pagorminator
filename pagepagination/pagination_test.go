package pagepagination

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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

	tests := map[string]struct {
		original func() *Pagination
		wantErr  error
	}{
		"fails when total elements are not set": {
			original: func() *Pagination { return Must(0, 2, pagegeneric.Asc("id")) },
			wantErr:  pagegeneric.ErrTotalElementsNotSet,
		},
		"fails when already at the last page": {
			original: func() *Pagination {
				p := Must(2, 2)
				_ = p.SetTotalElements(5)

				return p
			},
			wantErr: ErrNoNextPage,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			next, err := test.original().Next()
			if !errors.Is(err, test.wantErr) {
				t.Errorf("Next() = _, %v, want %v", err, test.wantErr)
			}

			if next != nil {
				t.Errorf("Next() = nil, _, expected nil, got %#v", next)
			}
		})
	}

	t.Run("returns next page and do not preserve configuration", func(t *testing.T) {
		t.Parallel()

		p := Must(0, 2, pagegeneric.Asc("id"), pagegeneric.Desc("price"))
		if err := p.SetTotalElements(5); err != nil {
			t.Errorf("unexpected error setting total elements: %v", err)
		}

		next, err := p.Next()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		want := &Pagination{
			page: 1,
			size: 2,
			sort: pagegeneric.Sort{
				pagegeneric.Asc("id"),
				pagegeneric.Desc("price"),
			},
		}
		if diff := cmp.Diff(
			next,
			want,
			cmp.AllowUnexported(Pagination{}),
			cmpopts.EquateComparable(sync.RWMutex{}),
		); diff != "" {
			t.Errorf("Next() mismatch (-want +got):\n%s", diff)
		}
	})
}

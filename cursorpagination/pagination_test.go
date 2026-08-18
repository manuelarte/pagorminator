package cursorpagination

import (
	"errors"
	"fmt"
	"testing"

	"github.com/manuelarte/pagorminator/pagegeneric"
)

func ExampleNew() {
	cursorPage, err := New(10, Asc("id", nil))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Size: %d, Cursors: %d\n", cursorPage.Size(), len(cursorPage.Cursors()))
	// Output: Size: 10, Cursors: 1
}

func ExampleMust() {
	cursorPage := Must(10, Asc("id", nil))

	fmt.Printf("Size: %d, Cursors: %d\n", cursorPage.Size(), len(cursorPage.Cursors()))
	// Output: Size: 10, Cursors: 1
}

func TestUnPaged(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		size     int
		cursors  []Cursor
		expected bool
	}{
		"size zero cursor nil": {
			size:     0,
			cursors:  nil,
			expected: true,
		},
		"size zero, cursor empty": {
			size:     0,
			cursors:  []Cursor{},
			expected: true,
		},
		"size not zero": {
			size:     1,
			cursors:  []Cursor{Asc("id", nil)},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			page, err := New(test.size, test.cursors...)
			if err != nil {
				t.Fatalf("NewPagination(%d, %v) = %s, unexpected error", test.size, test.cursors, err)
			}

			if page.IsUnPaged() != test.expected {
				t.Errorf("IsUnPaged() expected %v, got %v", test.expected, page.IsUnPaged())
			}
		})
	}
}

func TestNewPageRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		size        int
		cursors     []Cursor
		expectedErr error
	}{
		"negative size": {
			size:        -1,
			cursors:     []Cursor{Asc("id", nil)},
			expectedErr: ErrSizeCantBeNegative,
		},
		"missing order": {
			size:        10,
			cursors:     nil,
			expectedErr: ErrCursorsRequired,
		},
		"mixed cursor values": {
			size:    10,
			cursors: []Cursor{Asc("id", 3), Desc("price", nil)},
			expectedErr: CursorValuesNotValidError{
				CursorsHaveValues: []string{"id"},
				CursorsNilValue:   []string{"price"},
			},
		},
		"first page without cursor values": {
			size:    10,
			cursors: []Cursor{Asc("id", nil), Desc("price", nil)},
		},
		"valid request": {
			size:    10,
			cursors: []Cursor{Asc("id", 1), Desc("price", 2)},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := New(test.size, test.cursors...)
			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("New = %v, expected error %v", err, test.expectedErr)
			}

			if test.expectedErr != nil {
				return
			}

			if got.Size() != test.size {
				t.Errorf("size expected %d, got %d", test.size, got.Size())
			}

			gotCursors := got.Cursors()
			if len(gotCursors) != len(test.cursors) {
				t.Errorf("cursor count expected %d, got %d", len(test.cursors), len(gotCursors))
			}
		})
	}
}

func TestSetTotalElements(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		totalElements int64
		wantErr       error
	}{
		"positive totalElements": {
			totalElements: 2,
		},
		"0 totalElements": {
			totalElements: 0,
		},
		"negative totalElements": {
			totalElements: -1,
			wantErr:       pagegeneric.TotalElementsNotValidError{TotalElements: -1},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Pagination{}

			gotErr := p.SetTotalElements(test.totalElements)
			if !errors.Is(gotErr, test.wantErr) {
				t.Errorf("p.SetTotalElements(%d) = %v, got: %v", test.totalElements, gotErr, test.wantErr)
			}
		})
	}
}

func TestGetCursorsClone(t *testing.T) {
	t.Parallel()

	pageRequest := Must(10, Asc("id", 1))
	got := pageRequest.Cursors()
	got[0] = Asc("changed", 2)

	again := pageRequest.Cursors()
	if again[0].Column() != "id" {
		t.Errorf("expected cloned cursors, got %q", again[0].Column())
	}
}

func TestBuildCursorWhere(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cursors  []Cursor
		wantSQL  string
		wantVars []any
	}{
		"single asc": {
			cursors:  []Cursor{Asc("id", 3)},
			wantSQL:  "(id > ?)",
			wantVars: []any{3},
		},
		"single desc": {
			cursors:  []Cursor{Desc("id", 3)},
			wantSQL:  "(id < ?)",
			wantVars: []any{3},
		},
		"multi column mixed directions": {
			cursors:  []Cursor{Asc("code", "A"), Desc("price", 10)},
			wantSQL:  "(code > ?) OR (code = ? AND price < ?)",
			wantVars: []any{"A", "A", 10},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pageRequest := Must(5, test.cursors...)
			gotSQL, gotVars := pageRequest.buildCursorWhere()

			if gotSQL != test.wantSQL {
				t.Fatalf("sql expected %q, got %q", test.wantSQL, gotSQL)
			}

			if len(gotVars) != len(test.wantVars) {
				t.Fatalf("vars size expected %d, got %d", len(test.wantVars), len(gotVars))
			}

			for i := range test.wantVars {
				if gotVars[i] != test.wantVars[i] {
					t.Errorf("var[%d] = %v, want %v", i, gotVars[i], test.wantVars[i])
				}
			}
		})
	}
}

func TestNext(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		page *Pagination
		want pagegeneric.PrevNextPossible
	}{
		"no latest cursor values": {
			page: &Pagination{
				latestCursorValuesSet: false,
			},
			want: pagegeneric.PreviousCursorValuesNotSet,
		},
		"no next page": {
			page: &Pagination{
				size:                  10,
				totalElementsSet:      true,
				latestCursorValuesSet: true,
				latestLen:             5,
				latestCursorValues:    map[string]any{"id": 1},
			},
			want: pagegeneric.NoNextPage,
		},
		"success": {
			page: &Pagination{
				size:                  10,
				cursors:               []Cursor{Asc("id", 1)},
				totalElementsSet:      true,
				latestCursorValuesSet: true,
				latestLen:             10,
				latestCursorValues:    map[string]any{"id": 1},
			},
			want: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, hasNext := test.page.Next()
			if hasNext != test.want {
				t.Errorf("Next() = _, %v, want %v", hasNext, test.want)
			}
		})
	}
}

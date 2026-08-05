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

	fmt.Printf("Size: %d, Cursors: %d\n", cursorPage.GetSize(), len(cursorPage.GetCursors()))
	// Output: Size: 10, Cursors: 1
}

func ExampleMust() {
	cursorPage := Must(10, Asc("id", nil))

	fmt.Printf("Size: %d, Cursors: %d\n", cursorPage.GetSize(), len(cursorPage.GetCursors()))
	// Output: Size: 10, Cursors: 1
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
		"invalid order": {
			size:        10,
			cursors:     []Cursor{{Column: "id"}},
			expectedErr: ErrOrderNotValid,
		},
		"missing order": {
			size:        10,
			cursors:     nil,
			expectedErr: ErrOrderRequired,
		},
		"mixed cursor values": {
			size:        10,
			cursors:     []Cursor{Asc("id", 3), Desc("price", nil)},
			expectedErr: ErrCursorValueNotValid,
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
				t.Fatalf("expected error %v, got %v", test.expectedErr, err)
			}

			if test.expectedErr != nil {
				return
			}

			if got.GetSize() != test.size {
				t.Errorf("size expected %d, got %d", test.size, got.GetSize())
			}

			gotCursors := got.GetCursors()
			if len(gotCursors) != len(test.cursors) {
				t.Fatalf("cursor count expected %d, got %d", len(test.cursors), len(gotCursors))
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

func TestGetCursorsClone(t *testing.T) {
	t.Parallel()

	pageRequest := Must(10, Asc("id", 1))
	got := pageRequest.GetCursors()
	got[0].Column = "changed"

	again := pageRequest.GetCursors()
	if again[0].Column != "id" {
		t.Fatalf("expected cloned cursors, got %q", again[0].Column)
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
					t.Fatalf("var[%d] expected %v, got %v", i, test.wantVars[i], gotVars[i])
				}
			}
		})
	}
}

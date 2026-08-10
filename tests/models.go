package tests

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
)

type TestStruct struct {
	gorm.Model

	Code  string
	Price uint
}

func compareTestStructs(t *testing.T, want, got []*TestStruct) {
	if diff := cmp.Diff(
		want,
		got,
		cmpopts.IgnoreFields(TestStruct{}, "Model"),
	); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
	}
}

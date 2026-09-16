package conv

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBoolFromPtr(t *testing.T) {
	if got := BoolFromPtr(nil); !got.IsNull() {
		t.Fatalf("nil should become BoolNull, got %v", got)
	}
	tr := true
	if got := BoolFromPtr(&tr); !got.Equal(types.BoolValue(true)) {
		t.Fatalf("true ptr should become BoolValue(true), got %v", got)
	}
	fa := false
	if got := BoolFromPtr(&fa); !got.Equal(types.BoolValue(false)) {
		t.Fatalf("false ptr should become BoolValue(false), got %v", got)
	}
}

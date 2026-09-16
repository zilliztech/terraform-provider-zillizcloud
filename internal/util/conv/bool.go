package conv

import "github.com/hashicorp/terraform-plugin-framework/types"

// BoolFromPtr converts a *bool to a types.Bool, preserving null: a nil pointer
// becomes types.BoolNull() rather than collapsing to a zero value. Use it for
// API fields that may be omitted (null = not applicable / not reported).
func BoolFromPtr(v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*v)
}

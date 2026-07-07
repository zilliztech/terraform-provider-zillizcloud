package planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ignoreChangesString implements a PlanModifier that suppresses all diffs
// for deprecated fields by keeping the prior state value.
type ignoreChangesString struct{}

func (m ignoreChangesString) Description(ctx context.Context) string {
	return "Ignore any configuration changes by keeping the prior state value."
}

func (m ignoreChangesString) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ignoreChangesString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the resource already exists (state is not null/unknown), keep the state value
	// This completely suppresses any diff for this field
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.PlanValue = req.StateValue
		return
	}

	// For new resources: use the config value (or null if not set)
	// This allows the first apply to write a value to state, which is then preserved
	resp.PlanValue = req.ConfigValue
}

// IgnoreChangesString returns a PlanModifier that suppresses all diffs
// by keeping the prior state value. Use this for deprecated fields.
func IgnoreChangesString() planmodifier.String {
	return ignoreChangesString{}
}

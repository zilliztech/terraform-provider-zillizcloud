package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var _ resource.ResourceWithUpgradeState = &ServiceAccountResource{}

func (r *ServiceAccountResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				if req.RawState == nil || len(req.RawState.JSON) == 0 {
					resp.Diagnostics.AddError("Unable to upgrade Service Account state", "Expected JSON state from the previous Service Account schema.")
					return
				}
				// Preserve all other attributes while replacing the former role set with one ID.
				var attributes map[string]json.RawMessage
				if err := json.Unmarshal(req.RawState.JSON, &attributes); err != nil {
					resp.Diagnostics.AddError("Unable to upgrade Service Account state", err.Error())
					return
				}
				if attributes == nil {
					resp.Diagnostics.AddError("Unable to upgrade Service Account state", "Expected an object containing the previous Service Account attributes.")
					return
				}
				var roles []string
				if raw, exists := attributes["org_role_ids"]; exists {
					if err := json.Unmarshal(raw, &roles); err != nil {
						resp.Diagnostics.AddError("Unable to upgrade Service Account roles", err.Error())
						return
					}
				}
				if len(roles) > 1 {
					resp.Diagnostics.AddError(
						"Multiple organization roles in previous state",
						"The current Service Account API supports one organization role. "+
							"Reconcile the account to a single predefined Billing Admin or Org Owner role, "+
							"or project access, and re-import it before applying this configuration.",
					)
					return
				}
				var role *string
				if len(roles) == 1 {
					role = &roles[0]
				}
				encodedRole, err := json.Marshal(role)
				if err != nil {
					resp.Diagnostics.AddError("Unable to upgrade Service Account roles", err.Error())
					return
				}
				delete(attributes, "org_role_ids")
				attributes["org_role_id"] = encodedRole
				encoded, err := json.Marshal(attributes)
				if err != nil {
					resp.Diagnostics.AddError("Unable to upgrade Service Account state", err.Error())
					return
				}
				upgraded := tfprotov6.RawState{JSON: encoded}
				value, err := upgraded.Unmarshal(resp.State.Schema.Type().TerraformType(ctx))
				if err != nil {
					resp.Diagnostics.AddError("Unable to upgrade Service Account state", err.Error())
					return
				}
				resp.State.Raw = value
			},
		},
	}
}

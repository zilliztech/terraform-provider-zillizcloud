package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

func TestServiceAccountUpgradeState(t *testing.T) {
	ctx := context.Background()
	r := &ServiceAccountResource{}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	require.EqualValues(t, 1, schema.Schema.Version)
	for _, tc := range []struct {
		name    string
		roles   []string
		wantErr bool
	}{
		{name: "null roles"},
		{name: "empty roles", roles: []string{}},
		{name: "single role", roles: []string{"r-billing-admin"}},
		{name: "multiple roles", roles: []string{"r-billing-admin", "r-org-owner"}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := map[string]any{
				"id":                  "sa-existing",
				"name":                "existing",
				"description":         "keep this description",
				"status":              "disable",
				"project_scope":       nil,
				"org_role_ids":        tc.roles,
				"project_role_ids":    []string{},
				"project_assignments": map[string][]string{},
				"created_by":          "key-creator",
				"creator_name":        "creator",
				"create_time":         "2026-11-23T00:00:00Z",
			}
			encoded, err := json.Marshal(previous)
			require.NoError(t, err)
			resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schema.Schema}}
			r.UpgradeState(ctx)[0].StateUpgrader(ctx, resource.UpgradeStateRequest{
				RawState: &tfprotov6.RawState{JSON: encoded},
			}, &resp)
			require.Equal(t, tc.wantErr, resp.Diagnostics.HasError(), resp.Diagnostics)
			if tc.wantErr {
				return
			}
			var current ServiceAccountResourceModel
			require.False(t, resp.State.Get(ctx, &current).HasError())
			require.Equal(t, "sa-existing", current.Id.ValueString())
			require.Equal(t, "disable", current.Status.ValueString())
			require.Equal(t, "keep this description", current.Description.ValueString())
			require.Equal(t, "key-creator", current.CreatedBy.ValueString())
			require.Equal(t, len(tc.roles) == 0, current.OrgRoleId.IsNull())
			if len(tc.roles) == 1 {
				require.Equal(t, tc.roles[0], current.OrgRoleId.ValueString())
			}
			delete(previous, "org_role_ids")
			previous["org_role_id"] = nil
			if len(tc.roles) == 1 {
				previous["org_role_id"] = tc.roles[0]
			}
			expected, err := json.Marshal(previous)
			require.NoError(t, err)
			expectedState := tfprotov6.RawState{JSON: expected}
			expectedValue, err := expectedState.Unmarshal(schema.Schema.Type().TerraformType(ctx))
			require.NoError(t, err)
			require.True(t, expectedValue.Equal(resp.State.Raw), "all other state attributes must be preserved")
		})
	}
}

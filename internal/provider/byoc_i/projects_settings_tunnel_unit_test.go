package byoc_op

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestUnitSettingsTunnelClientImageSchema(t *testing.T) {
	var dataSchema datasource.SchemaResponse
	NewBYOCOpProjectSettingsData().Schema(context.Background(), datasource.SchemaRequest{}, &dataSchema)
	require.False(t, dataSchema.Diagnostics.HasError())
	dataConfig := dataSchema.Schema.Attributes["op_config"].(datasourceschema.SingleNestedAttribute)
	dataImage := dataConfig.Attributes["tunnel_client_image_url"].(datasourceschema.StringAttribute)
	require.True(t, dataImage.Computed)
	require.False(t, dataImage.Required)

	var resourceSchema resource.SchemaResponse
	NewBYOCOpProjectSettingsResource().Schema(context.Background(), resource.SchemaRequest{}, &resourceSchema)
	require.False(t, resourceSchema.Diagnostics.HasError())
	resourceConfig := resourceSchema.Schema.Attributes["op_config"].(resourceschema.SingleNestedAttribute)
	resourceImage := resourceConfig.Attributes["tunnel_client_image_url"].(resourceschema.StringAttribute)
	require.True(t, resourceImage.Computed)
	require.False(t, resourceImage.Required)
}

func TestUnitSettingsTunnelClientImageProjection(t *testing.T) {
	for _, tc := range []struct {
		name    string
		image   any
		include bool
		want    string
	}{
		{name: "configured", image: "cloud-tunnel-client:test", include: true, want: "cloud-tunnel-client:test"},
		{name: "missing"},
		{name: "null", include: true},
		{name: "empty", image: "", include: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "project-test", r.URL.Query().Get("projectId"))
				assert.Equal(t, "dataplane-test", r.URL.Query().Get("dataPlaneId"))
				var data map[string]any
				switch r.URL.Path {
				case "/byoc/dataplane/describe":
					data = map[string]any{"projectId": "project-test", "dataPlaneId": "dataplane-test", "status": 99}
				case "/byoc/op/dataplane/setting":
					opConfig := map[string]any{"tunnelToken": "test-token", "agentImageUrl": "cloud-agent:test"}
					if tc.include {
						opConfig["tunnelClientImageUrl"] = tc.image
					}
					data = map[string]any{"opConfig": opConfig}
				default:
					t.Errorf("unexpected request path %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data}))
			}))
			defer server.Close()
			client, err := zilliz.NewClient(zilliz.WithApiKey("test-key"), zilliz.WithBaseUrl(server.URL), zilliz.WithHTTPClient(server.Client()))
			require.NoError(t, err)
			dataStore := byocOpProjectSettingsDataStore{client: client}
			data, err := dataStore.Describe(context.Background(), "project-test", "dataplane-test")
			require.NoError(t, err)
			resourceStore := byocOpProjectSettingsStore{client: client}
			state, err := resourceStore.Describe(context.Background(), "project-test", "dataplane-test")
			require.NoError(t, err)
			for _, config := range []types.Object{data.OpConfig, state.OpConfig} {
				require.Equal(t, types.StringValue(tc.want), config.Attributes()["tunnel_client_image_url"])
				require.Equal(t, types.StringValue("cloud-agent:test"), config.Attributes()["agent_image_url"])
				require.Equal(t, types.StringValue("test-token"), config.Attributes()["token"])
				var model OpConfig
				diagnostics := config.As(context.Background(), &model, basetypes.ObjectAsOptions{})
				require.False(t, diagnostics.HasError(), diagnostics)
				require.Equal(t, tc.want, model.TunnelClientImageUrl.ValueString())
			}
		})
	}
}

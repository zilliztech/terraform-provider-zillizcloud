package byoc_op

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestAliRegistrationCreateAndRefresh(t *testing.T) {
	ctx := context.Background()
	create, err := os.ReadFile("../../../client/testdata/aliyun/create.json")
	require.NoError(t, err)
	describe, err := os.ReadFile("../../../client/testdata/aliyun/describe.json")
	require.NoError(t, err)
	var request zilliz.CreateByocOpProjectRequest
	require.NoError(t, json.Unmarshal(create, &request))
	config, err := aliFromResponse(ctx, request.RegionID, request.AliParam)
	require.NoError(t, err)
	var captured []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /byoc/op/dataplane/create":
			captured, _ = io.ReadAll(r.Body)
			_, _ = w.Write([]byte(`{"code":0,"data":{"projectId":"prj-byoci","dataPlaneId":"vpc-byoci","jobId":"job-1"}}`))
		case "GET /byoc/dataplane/describe":
			if r.URL.Query().Get("projectId") != "prj-byoci" || r.URL.Query().Get("dataPlaneId") != "vpc-byoci" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_, _ = w.Write(describe)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	c, err := zilliz.NewClient(zilliz.WithBaseUrl(server.URL), zilliz.WithApiKey("test"), zilliz.WithCloudRegionId("ali-cn-hangzhou"))
	require.NoError(t, err)
	store := &byocOpProjectStore{client: c}
	model := BYOCOpProjectResourceModel{
		ProjectID: types.StringValue(request.ProjectId), DataPlaneID: types.StringValue(request.DataPlaneId),
		Ali: config, ExtConfig: types.StringPointerValue(request.ExtConfig),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{"create": types.StringType, "delete": types.StringType, "update": types.StringType})},
	}
	require.NoError(t, store.Create(ctx, &model, func(*BYOCOpProjectResourceModel) error { return nil }))
	require.JSONEq(t, string(create), string(captured))
	require.Equal(t, types.StringValue("prj-byoci:vpc-byoci"), model.ID)

	r := NewBYOCOpProjectResource().(*BYOCOpProjectResource)
	r.store = store
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), schemaResponse.Diagnostics)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	// Import starts without Aliyun attributes; Read must reconstruct the complete block.
	model.Ali = nil
	diags := state.Set(ctx, &model)
	require.False(t, diags.HasError(), diags)
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	var refreshed BYOCOpProjectResourceModel
	diags = response.State.Get(ctx, &refreshed)
	require.False(t, diags.HasError(), diags)
	require.Equal(t, config, refreshed.Ali)
	require.Equal(t, types.StringPointerValue(request.ExtConfig), refreshed.ExtConfig)
	require.Equal(t, types.Int64Value(1), refreshed.Status)
}

func TestAliDescribeRejectsMissingConfigAndHandlesNullEndpoint(t *testing.T) {
	ctx := context.Background()
	raw, err := os.ReadFile("../../../client/testdata/aliyun/describe.json")
	require.NoError(t, err)
	var fixture map[string]any
	require.NoError(t, json.Unmarshal(raw, &fixture))
	data := fixture["data"].(map[string]any)
	data["aliConfig"].(map[string]any)["endpointId"] = nil
	data["opConfig"] = nil
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(fixture) }))
	defer server.Close()
	c, err := zilliz.NewClient(zilliz.WithBaseUrl(server.URL), zilliz.WithApiKey("test"), zilliz.WithCloudRegionId("ali-cn-hangzhou"))
	require.NoError(t, err)
	store := &byocOpProjectStore{client: c}
	model, err := store.Describe(ctx, "prj-byoci", "vpc-byoci")
	require.NoError(t, err)
	require.True(t, model.ExtConfig.IsNull())
	require.True(t, model.Ali.EndpointID.IsNull())
	delete(data, "aliConfig")
	_, err = store.Describe(ctx, "prj-byoci", "vpc-byoci")
	require.ErrorContains(t, err, "no aliConfig")
	data["status"] = 3
	_, err = store.Describe(ctx, "prj-byoci", "vpc-byoci")
	require.NoError(t, err)
}

// Reading the canonical opConfig field must not copy an agent token into ext_config
// for existing cloud providers either.
func TestByocDescribeExtConfigAcrossCloudsAndResponseVersions(t *testing.T) {
	for _, cloud := range []string{"aws", "gcp", "az"} {
		for _, field := range []string{"opConfig", "vpcOpConfig"} {
			t.Run(cloud+"/"+field, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": map[string]any{
						"cloudId": cloud, "projectId": "prj-1", "dataPlaneId": "dp-1", "status": 1,
						field: map[string]any{"extConfig": "opaque-extension", "token": "agent-secret"},
					}})
				}))
				defer server.Close()
				c, err := zilliz.NewClient(zilliz.WithBaseUrl(server.URL), zilliz.WithApiKey("test"), zilliz.WithCloudRegionId("ali-cn-hangzhou"))
				require.NoError(t, err)
				model, err := (&byocOpProjectStore{client: c}).Describe(context.Background(), "prj-1", "dp-1")
				require.NoError(t, err)
				require.Equal(t, types.StringValue("opaque-extension"), model.ExtConfig)
			})
		}
	}
}

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestServiceAccountsPagination(t *testing.T) {
	ctx := context.Background()
	for _, keys := range []bool{
		false,
		true,
	} {
		for _, emptyLast := range []bool{
			false,
			true,
		} {
			t.Run(fmt.Sprintf("keys=%t-empty=%t", keys, emptyLast), func(t *testing.T) {
				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					require.Equal(t, fmt.Sprint(requests), r.URL.Query().Get("currentPage"))
					require.Equal(t, "100", r.URL.Query().Get("pageSize"))
					require.Equal(t, "a&b", r.URL.Query().Get("name"))
					if keys {
						require.Equal(t, "/serviceAccounts/sa-test/apiKeys", r.URL.Path)
						require.Equal(t, "key", r.URL.Query().Get("apiKeyId"))
					} else {
						require.Equal(t, "proj-test", r.URL.Query().Get("projectId"))
						require.Equal(t, "active", r.URL.Query().Get("status"))
					}
					data := map[string]any{
						"count":       2,
						"currentPage": requests,
						"pageSize":    100,
					}
					if keys {
						items := []zilliz.ServiceAccountKeyResponse{{
							ApiKeyId:         fmt.Sprintf("key-%d", requests),
							ServiceAccountId: "sa-test",
						}}
						if requests == 2 && emptyLast {
							items = nil
						}
						data["apiKeys"] = items
					} else {
						items := []zilliz.ServiceAccountResponse{{ServiceAccountId: fmt.Sprintf("sa-%d", requests)}}
						if requests == 2 && emptyLast {
							items = nil
						}
						data["serviceAccounts"] = items
					}
					require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
						"code": 0,
						"data": data,
					}))
				}))
				defer server.Close()
				c, err := zilliz.NewClient(zilliz.WithApiKey("test"), zilliz.WithBaseUrl(server.URL))
				require.NoError(t, err)
				var d datasource.DataSource
				var model any
				if keys {
					d = &ServiceAccountKeysDataSource{client: c}
					model = &ServiceAccountKeysDataSourceModel{
						ServiceAccountId: types.StringValue("sa-test"),
						Name:             types.StringValue("a&b"),
						ApiKeyId:         types.StringValue("key"),
					}
				} else {
					d = &ServiceAccountsDataSource{client: c}
					model = &ServiceAccountsDataSourceModel{
						Name:      types.StringValue("a&b"),
						ProjectId: types.StringValue("proj-test"),
						Status:    types.StringValue("active"),
					}
				}
				var schema datasource.SchemaResponse
				d.Schema(ctx, datasource.SchemaRequest{}, &schema)
				config := tfsdk.State{Schema: schema.Schema}
				require.False(t, config.Set(ctx, model).HasError())
				resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
				d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{
					Schema: schema.Schema,
					Raw:    config.Raw,
				}}, &resp)
				require.Equal(t, 2, requests)
				require.Equal(t, emptyLast, resp.Diagnostics.HasError(), resp.Diagnostics)
				if !emptyLast {
					if keys {
						var result ServiceAccountKeysDataSourceModel
						require.False(t, resp.State.Get(ctx, &result).HasError())
						require.Len(t, result.Items, 2)
					} else {
						var result ServiceAccountsDataSourceModel
						require.False(t, resp.State.Get(ctx, &result).HasError())
						require.Len(t, result.Items, 2)
					}
				}
			})
		}
	}
}

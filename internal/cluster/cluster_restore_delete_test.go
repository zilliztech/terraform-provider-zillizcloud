package cluster

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestRestoreDeleteCancelsHTTP(t *testing.T) {
	for _, test := range []struct {
		name    string
		timeout string
		cancel  bool
	}{
		{name: "caller cancellation", timeout: "5s", cancel: true},
		{name: "configured timeout", timeout: "300ms"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			started := make(chan struct{})
			stopped := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodDelete || req.URL.Path != "/v2/clusters/in01-target/drop" {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
					http.NotFound(w, req)
					return
				}
				close(started)
				select {
				case <-req.Context().Done():
					close(stopped)
				case <-time.After(3 * time.Second):
					t.Error("delete HTTP request was not canceled")
				}
			}))
			defer server.Close()
			client, err := zilliz.NewClient(zilliz.WithBaseUrl(server.URL+"/v2"), zilliz.WithApiKey("mock-key"))
			if err != nil {
				t.Fatal(err)
			}
			restore := &ClusterRestoreResource{client: client}
			var schema resource.SchemaResponse
			restore.Schema(ctx, resource.SchemaRequest{}, &schema)
			state := tfsdk.State{Schema: schema.Schema, Raw: tftypes.NewValue(schema.Schema.Type().TerraformType(ctx), nil)}
			for _, attribute := range []struct {
				path  path.Path
				value string
			}{
				{path.Root("id"), "in01-target"},
				{path.Root("timeouts").AtName("delete"), test.timeout},
			} {
				if diags := state.SetAttribute(ctx, attribute.path, attribute.value); diags.HasError() {
					t.Fatalf("prepare state: %v", diags)
				}
			}

			if test.cancel {
				go func() {
					select {
					case <-started:
						cancel()
					case <-ctx.Done():
					}
				}()
			}
			resp := resource.DeleteResponse{State: state}
			restore.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected failed delete diagnostic")
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("HTTP server did not observe request cancellation")
			}
		})
	}
}

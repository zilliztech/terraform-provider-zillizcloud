package byoc_op

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func azureCmekModel(t *testing.T, enabled bool) BYOCOpProjectResourceModel {
	t.Helper()
	identityType := types.ObjectType{AttrTypes: map[string]attr.Type{"client_id": types.StringType, "principal_id": types.StringType, "resource_id": types.StringType}}
	storageIdentities, diags := types.SetValueFrom(context.Background(), identityType, []AzureIdentity{{ClientID: types.StringValue("storage-client"), PrincipalID: types.StringValue("storage-principal"), ResourceID: types.StringValue("storage-resource")}})
	if diags.HasError() {
		t.Fatal(diags)
	}
	model := BYOCOpProjectResourceModel{
		ProjectID: types.StringValue("proj-test"), DataPlaneID: types.StringValue("dataplane-test"), ExtConfig: types.StringNull(),
		Azure: &AzureConfig{
			Region:   types.StringValue("westus3"),
			Network:  AzureNetworkConfig{VNetID: types.StringValue("vnet"), SubnetIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("subnet")}), NSGIDs: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("nsg")}), PrivateEndpointID: types.StringNull()},
			Storage:  AzureStorageConfig{StorageAccountName: types.StringValue("storage"), ContainerName: types.StringValue("container")},
			Identity: AzureIdentityConfig{Storages: storageIdentities, Kubelet: AzureIdentity{ClientID: types.StringValue("kubelet-client"), PrincipalID: types.StringValue("kubelet-principal"), ResourceID: types.StringValue("kubelet-resource")}, Maintenance: AzureIdentity{ClientID: types.StringValue("maintenance-client"), PrincipalID: types.StringValue("maintenance-principal"), ResourceID: types.StringValue("maintenance-resource")}},
		},
	}
	if enabled {
		model.Azure.CSE = &AzureCSEConfig{Identity: AzureIdentity{ClientID: types.StringValue("cse-client"), PrincipalID: types.StringValue("cse-principal"), ResourceID: types.StringValue("cse-resource")}, TenantID: types.StringValue("cse-tenant"), DefaultKeyIdentifier: types.StringValue("https://customer-kms.vault.azure.net/keys/cmek-key")}
	}
	return model
}

func TestAzureCmekCreateRequest(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "cmek"}[enabled], func(t *testing.T) {
			var got zilliz.CreateByocOpProjectRequest
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != http.MethodPost || r.URL.Path != "/byoc/op/dataplane/create" {
					t.Errorf("unexpected route: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Error(err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":0,"data":{"projectId":"proj-test","dataPlaneId":"dataplane-test"}}`))
			}))
			defer server.Close()
			client, err := zilliz.NewClient(zilliz.WithApiKey("test-api-key"), zilliz.WithBaseUrl(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			model := azureCmekModel(t, enabled)
			store := &byocOpProjectStore{client: client}
			if err := store.Create(context.Background(), &model, func(_ *BYOCOpProjectResourceModel) error { return nil }); err != nil {
				t.Fatal(err)
			}
			if !called || got.CloudID != zilliz.Azure || got.RegionID != "westus3" || got.AWSParam != nil || got.GCPParam != nil {
				t.Fatalf("wrong cloud request: %#v", got)
			}
			if got.AzureParam == nil {
				t.Fatal("missing Azure configuration")
			}
			if (got.AzureParam.CSEIdentity != nil) != enabled {
				t.Fatalf("unexpected CSE identity: %#v", got.AzureParam)
			}
			wantStorage := []zilliz.AzureIdentityParam{{ClientID: "storage-client", PrincipalID: "storage-principal", ResourceID: "storage-resource"}}
			if !reflect.DeepEqual(got.AzureParam.StorageIdentities, wantStorage) || got.AzureParam.KubeletIdentity.ClientID != "kubelet-client" || got.AzureParam.MaintenanceIdentity.ClientID != "maintenance-client" {
				t.Fatalf("existing workload identities changed: %#v", got.AzureParam)
			}
			if enabled {
				want := zilliz.AzureIdentityParam{ClientID: "cse-client", PrincipalID: "cse-principal", ResourceID: "cse-resource"}
				if !reflect.DeepEqual(*got.AzureParam.CSEIdentity, want) || got.AzureParam.AzureTenantID != "cse-tenant" || got.AzureParam.DefaultAzureCseKeyIdentifier != "https://customer-kms.vault.azure.net/keys/cmek-key" {
					t.Fatalf("wrong CSE request: %#v", got.AzureParam)
				}
			} else if got.AzureParam.AzureTenantID != "" || got.AzureParam.DefaultAzureCseKeyIdentifier != "" {
				t.Fatal("legacy Azure gained CMEK settings")
			}
		})
	}
}

func TestAzureCmekReadRetainsConfiguredIdentities(t *testing.T) {
	ctx := context.Background()
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "cmek"}[enabled], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/byoc/dataplane/describe" || r.URL.Query().Get("projectId") != "proj-test" || r.URL.Query().Get("dataPlaneId") != "dataplane-test" {
					t.Errorf("unexpected describe route: %s %s", r.Method, r.URL)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":0,"data":{"projectId":"proj-test","dataPlaneId":"dataplane-test","status":1,"cloudId":"azure"}}`))
			}))
			defer server.Close()
			client, err := zilliz.NewClient(zilliz.WithApiKey("test-api-key"), zilliz.WithBaseUrl(server.URL))
			if err != nil {
				t.Fatal(err)
			}
			r := &BYOCOpProjectResource{store: &byocOpProjectStore{client: client}}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			model := azureCmekModel(t, enabled)
			// Match the custom timeout type expected by the resource state schema.
			model.Timeouts.Object = types.ObjectNull(map[string]attr.Type{"create": types.StringType, "update": types.StringType, "delete": types.StringType})
			state := tfsdk.State{Schema: schemaResponse.Schema}
			if diags := state.Set(ctx, &model); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			var actual BYOCOpProjectResourceModel
			if diags := response.State.Get(ctx, &actual); diags.HasError() {
				t.Fatal(diags)
			}
			if !reflect.DeepEqual(actual.Azure, model.Azure) {
				t.Fatalf("Azure identities or CSE were lost on refresh: got %#v, want %#v", actual.Azure, model.Azure)
			}
			if actual.ID.ValueString() != "proj-test:dataplane-test" || actual.Status.ValueInt64() != 1 || actual.AWS != nil || actual.GCP != nil {
				t.Fatalf("lifecycle refresh or cloud isolation failed: %#v", actual)
			}
		})
	}
}

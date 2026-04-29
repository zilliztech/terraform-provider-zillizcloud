# Private Link API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Terraform provider support for the Zilliz Cloud Private Link Management API — 2 data sources (`endpoint_services`, `endpoints`) and 2 resources (`endpoint`, `endpoint_whitelist`).

**Architecture:** Add a client-layer file `client/endpoint.go` with typed request/response structs and `*Client` methods wrapping `c.do(...)`. Add Terraform data sources and resources under `internal/provider/` following the existing patterns from `cloud_regions_data_source.go` and `database_resource.go`. Register in `provider.go`.

**Tech Stack:** Go 1.x, Terraform Plugin Framework, existing Zilliz REST client.

**Spec:** `docs/superpowers/specs/2026-04-23-private-link-api-design.md`

---

## File Structure

**New files:**
- `client/endpoint.go` — types + 5 client methods
- `client/endpoint_test.go` — client unit tests
- `internal/provider/endpoint_services_data_source.go`
- `internal/provider/endpoints_data_source.go`
- `internal/provider/endpoint_resource.go`
- `internal/provider/endpoint_whitelist_resource.go`
- `examples/data-sources/zillizcloud_endpoint_services/data-source.tf`
- `examples/data-sources/zillizcloud_endpoints/data-source.tf`
- `examples/resources/zillizcloud_endpoint/resource.tf`
- `examples/resources/zillizcloud_endpoint_whitelist/resource.tf`

**Modify:**
- `internal/provider/provider.go` — register new resources/data sources

---

## Task 1: Client layer — types and methods

**Files:**
- Create: `client/endpoint.go`

- [ ] **Step 1: Create `client/endpoint.go` with types and methods**

```go
package client

import (
	"fmt"
	"net/url"
)

// EndpointService represents an available private link endpoint service.
type EndpointService struct {
	RegionId          string `json:"regionId"`
	CloudId           string `json:"cloudId"`
	EndpointService   string `json:"endpointService"`
	WhitelistRequired bool   `json:"whitelistRequired"`
}

// Endpoint represents a VPC private link endpoint under a project.
type Endpoint struct {
	RegionId              string  `json:"regionId"`
	CloudId               string  `json:"cloudId"`
	EndpointService       string  `json:"endpointService"`
	EndpointServiceStatus string  `json:"endpointServiceStatus"`
	EndpointId            string  `json:"endpointId"`
	EndpointStatus        string  `json:"endpointStatus"`
	GcpProjectId          *string `json:"gcpProjectId"`
}

// listEndpointServicesData is the inner payload for GET /v2/endpointServices.
type listEndpointServicesData struct {
	EndpointServices []EndpointService `json:"endpointServices"`
	zillizPage
}

// listEndpointsData is the inner payload for GET /v2/projects/{projectId}/endpoints.
type listEndpointsData struct {
	Endpoints []Endpoint `json:"endpoints"`
	zillizPage
}

// CreateEndpointRequest is the body for POST /v2/projects/{projectId}/endpoints.
type CreateEndpointRequest struct {
	RegionId     string `json:"regionId"`
	EndpointId   string `json:"endpointId"`
	GcpProjectId string `json:"gcpProjectId,omitempty"`
}

// CreateEndpointResponse is the response payload for POST /v2/projects/{projectId}/endpoints.
type CreateEndpointResponse struct {
	EndpointId string `json:"endpointId"`
	RegionId   string `json:"regionId"`
}

// AddEndpointWhitelistRequest is the body for POST /v2/projects/{projectId}/endpointWhitelist.
type AddEndpointWhitelistRequest struct {
	RegionId    string `json:"regionId"`
	OuterUserId string `json:"outerUserId"`
}

// ListEndpointServices lists available private link endpoint services for a region.
func (c *Client) ListEndpointServices(regionId string, currentPage, pageSize int) ([]EndpointService, zillizPage, error) {
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := url.Values{}
	q.Set("regionId", regionId)
	q.Set("currentPage", fmt.Sprintf("%d", currentPage))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))

	var response zillizResponse[listEndpointServicesData]
	err := c.do("GET", "endpointServices?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.EndpointServices, response.Data.zillizPage, nil
}

// ListEndpoints lists private link endpoints under a project.
func (c *Client) ListEndpoints(projectId string, currentPage, pageSize int) ([]Endpoint, zillizPage, error) {
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := url.Values{}
	q.Set("currentPage", fmt.Sprintf("%d", currentPage))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))

	var response zillizResponse[listEndpointsData]
	err := c.do("GET", "projects/"+projectId+"/endpoints?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.Endpoints, response.Data.zillizPage, nil
}

// CreateEndpoint creates a private link endpoint under a project.
func (c *Client) CreateEndpoint(projectId string, req *CreateEndpointRequest) (*CreateEndpointResponse, error) {
	var response zillizResponse[CreateEndpointResponse]
	err := c.do("POST", "projects/"+projectId+"/endpoints", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

// DeleteEndpoint deletes a private link endpoint. regionId is required; gcpProjectId is required only for GCP regions.
func (c *Client) DeleteEndpoint(projectId, endpointId, regionId string, gcpProjectId *string) error {
	q := url.Values{}
	q.Set("regionId", regionId)
	if gcpProjectId != nil && *gcpProjectId != "" {
		q.Set("gcpProjectId", *gcpProjectId)
	}
	var response zillizResponse[map[string]any]
	return c.do("DELETE", "projects/"+projectId+"/endpoints/"+endpointId+"?"+q.Encode(), nil, &response)
}

// AddEndpointWhitelist adds an external cloud account to the endpoint whitelist.
func (c *Client) AddEndpointWhitelist(projectId string, req *AddEndpointWhitelistRequest) error {
	var response zillizResponse[string]
	return c.do("POST", "projects/"+projectId+"/endpointWhitelist", req, &response)
}
```

- [ ] **Step 2: Build to verify it compiles**

Run: `go build ./client/...`
Expected: success, no output.

- [ ] **Step 3: Run existing unit tests to verify no regressions**

Run: `make unit-test`
Expected: all existing tests pass.

- [ ] **Step 4: Commit**

```bash
git add client/endpoint.go
git commit -m "feat(client): add private link endpoint API methods"
```

---

## Task 2: Client layer — unit tests with mock HTTP

**Files:**
- Create: `client/endpoint_test.go`

- [ ] **Step 1: Look at the existing mock pattern**

Run: `grep -n "HttpClient" client/*_test.go | head -20`
Expected: see how existing tests inject `HttpClient`. If there's no mock pattern, use a minimal `http.RoundTripper` stub.

- [ ] **Step 2: Write the unit tests**

Create `client/endpoint_test.go`:

```go
package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type mockHTTPClient struct{ do func(*http.Request) (*http.Response, error) }

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) { return m.do(req) }

func newMockClient(t *testing.T, handler func(*http.Request) (*http.Response, error)) *Client {
	t.Helper()
	c, err := NewClient(
		WithApiKey("test-key"),
		WithBaseUrl("https://api.test/v2"),
		WithHTTPClient(&mockHTTPClient{do: handler}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func jsonResponse(t *testing.T, status int, body any) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestUnitListEndpointServices(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if !strings.Contains(req.URL.String(), "/endpointServices") {
			t.Errorf("url=%s", req.URL.String())
		}
		if req.URL.Query().Get("regionId") != "aws-us-west-2" {
			t.Errorf("regionId=%s", req.URL.Query().Get("regionId"))
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"endpointServices": []map[string]any{
					{"regionId": "aws-us-west-2", "cloudId": "aws", "endpointService": "svc-x", "whitelistRequired": false},
				},
				"currentPage": 1, "pageSize": 10, "count": 1,
			},
		}), nil
	})

	svcs, page, err := c.ListEndpointServices("aws-us-west-2", 1, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(svcs) != 1 || svcs[0].EndpointService != "svc-x" {
		t.Errorf("svcs=%+v", svcs)
	}
	if page.Count != 1 {
		t.Errorf("count=%d", page.Count)
	}
}

func TestUnitListEndpoints(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "/projects/proj-1/endpoints") {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{
				"endpoints": []map[string]any{
					{"regionId": "aws-us-west-2", "cloudId": "aws", "endpointService": "svc-x",
						"endpointServiceStatus": "Available", "endpointId": "vpce-abc",
						"endpointStatus": "accepted", "gcpProjectId": nil},
				},
				"currentPage": 1, "pageSize": 10, "count": 1,
			},
		}), nil
	})

	eps, _, err := c.ListEndpoints("proj-1", 1, 10)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(eps) != 1 || eps[0].EndpointId != "vpce-abc" {
		t.Errorf("eps=%+v", eps)
	}
	if eps[0].GcpProjectId != nil {
		t.Errorf("expected gcpProjectId nil, got %v", eps[0].GcpProjectId)
	}
}

func TestUnitCreateEndpoint(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		var body CreateEndpointRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.EndpointId != "vpce-abc" || body.RegionId != "aws-us-west-2" {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, 200, map[string]any{
			"code": 0,
			"data": map[string]any{"endpointId": "vpce-abc", "regionId": "aws-us-west-2"},
		}), nil
	})

	resp, err := c.CreateEndpoint("proj-1", &CreateEndpointRequest{
		RegionId: "aws-us-west-2", EndpointId: "vpce-abc",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.EndpointId != "vpce-abc" {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDeleteEndpoint(t *testing.T) {
	t.Run("no gcpProjectId", func(t *testing.T) {
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.Method != "DELETE" {
				t.Errorf("method=%s", req.Method)
			}
			if req.URL.Query().Get("regionId") != "aws-us-west-2" {
				t.Errorf("regionId missing")
			}
			if _, ok := req.URL.Query()["gcpProjectId"]; ok {
				t.Errorf("gcpProjectId should not be present")
			}
			return jsonResponse(t, 200, map[string]any{
				"code": 0, "data": map[string]any{"endpointId": "vpce-abc"},
			}), nil
		})
		if err := c.DeleteEndpoint("proj-1", "vpce-abc", "aws-us-west-2", nil); err != nil {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("with gcpProjectId", func(t *testing.T) {
		gcp := "my-gcp-proj"
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.URL.Query().Get("gcpProjectId") != "my-gcp-proj" {
				t.Errorf("gcpProjectId=%s", req.URL.Query().Get("gcpProjectId"))
			}
			return jsonResponse(t, 200, map[string]any{
				"code": 0, "data": map[string]any{"endpointId": "vpce-abc"},
			}), nil
		})
		if err := c.DeleteEndpoint("proj-1", "vpce-abc", "gcp-us-west1", &gcp); err != nil {
			t.Fatalf("err: %v", err)
		}
	})
}

func TestUnitAddEndpointWhitelist(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		if !strings.HasSuffix(req.URL.Path, "/projects/proj-1/endpointWhitelist") {
			t.Errorf("path=%s", req.URL.Path)
		}
		var body AddEndpointWhitelistRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.OuterUserId != "user-abc" {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, 200, map[string]any{"code": 0, "data": "success"}), nil
	})

	err := c.AddEndpointWhitelist("proj-1", &AddEndpointWhitelistRequest{
		RegionId: "azure-eastus2", OuterUserId: "user-abc",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}
```

- [ ] **Step 3: Run the new tests**

Run: `go test ./client/ -run TestUnit -v`
Expected: all 5 tests pass.

- [ ] **Step 4: Commit**

```bash
git add client/endpoint_test.go
git commit -m "test(client): add unit tests for private link endpoint API"
```

---

## Task 3: `zillizcloud_endpoint_services` data source

**Files:**
- Create: `internal/provider/endpoint_services_data_source.go`

- [ ] **Step 1: Create the data source file**

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &EndpointServicesDataSource{}

func NewEndpointServicesDataSource() datasource.DataSource {
	return &EndpointServicesDataSource{}
}

type EndpointServicesDataSource struct {
	client *zilliz.Client
}

type EndpointServiceItem struct {
	RegionId          types.String `tfsdk:"region_id"`
	CloudId           types.String `tfsdk:"cloud_id"`
	EndpointService   types.String `tfsdk:"endpoint_service"`
	WhitelistRequired types.Bool   `tfsdk:"whitelist_required"`
}

type EndpointServicesDataSourceModel struct {
	RegionId         types.String          `tfsdk:"region_id"`
	CurrentPage      types.Int64           `tfsdk:"current_page"`
	PageSize         types.Int64           `tfsdk:"page_size"`
	EndpointServices []EndpointServiceItem `tfsdk:"endpoint_services"`
	Count            types.Int64           `tfsdk:"count"`
}

func (d *EndpointServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint_services"
}

func (d *EndpointServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List available private link endpoint services for a region.",
		Attributes: map[string]schema.Attribute{
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
			},
			"current_page": schema.Int64Attribute{
				MarkdownDescription: "Page number (defaults to 1).",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Page size (1-100, defaults to 10).",
				Optional:            true,
			},
			"count": schema.Int64Attribute{
				MarkdownDescription: "Total count of endpoint services.",
				Computed:            true,
			},
			"endpoint_services": schema.ListNestedAttribute{
				MarkdownDescription: "List of endpoint services.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_id":          schema.StringAttribute{Computed: true},
						"cloud_id":           schema.StringAttribute{Computed: true},
						"endpoint_service":   schema.StringAttribute{Computed: true},
						"whitelist_required": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EndpointServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *EndpointServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EndpointServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentPage := 1
	if !state.CurrentPage.IsNull() {
		currentPage = int(state.CurrentPage.ValueInt64())
	}
	pageSize := 10
	if !state.PageSize.IsNull() {
		pageSize = int(state.PageSize.ValueInt64())
	}

	svcs, page, err := d.client.ListEndpointServices(state.RegionId.ValueString(), currentPage, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to ListEndpointServices, got error: %s", err))
		return
	}

	state.EndpointServices = nil
	for _, s := range svcs {
		state.EndpointServices = append(state.EndpointServices, EndpointServiceItem{
			RegionId:          types.StringValue(s.RegionId),
			CloudId:           types.StringValue(s.CloudId),
			EndpointService:   types.StringValue(s.EndpointService),
			WhitelistRequired: types.BoolValue(s.WhitelistRequired),
		})
	}
	state.Count = types.Int64Value(int64(page.Count))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/provider/endpoint_services_data_source.go
git commit -m "feat(provider): add zillizcloud_endpoint_services data source"
```

---

## Task 4: `zillizcloud_endpoints` data source

**Files:**
- Create: `internal/provider/endpoints_data_source.go`

- [ ] **Step 1: Create the data source file**

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ datasource.DataSource = &EndpointsDataSource{}

func NewEndpointsDataSource() datasource.DataSource {
	return &EndpointsDataSource{}
}

type EndpointsDataSource struct {
	client *zilliz.Client
}

type EndpointItem struct {
	RegionId              types.String `tfsdk:"region_id"`
	CloudId               types.String `tfsdk:"cloud_id"`
	EndpointService       types.String `tfsdk:"endpoint_service"`
	EndpointServiceStatus types.String `tfsdk:"endpoint_service_status"`
	EndpointId            types.String `tfsdk:"endpoint_id"`
	EndpointStatus        types.String `tfsdk:"endpoint_status"`
	GcpProjectId          types.String `tfsdk:"gcp_project_id"`
}

type EndpointsDataSourceModel struct {
	ProjectId   types.String   `tfsdk:"project_id"`
	CurrentPage types.Int64    `tfsdk:"current_page"`
	PageSize    types.Int64    `tfsdk:"page_size"`
	Endpoints   []EndpointItem `tfsdk:"endpoints"`
	Count       types.Int64    `tfsdk:"count"`
}

func (d *EndpointsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoints"
}

func (d *EndpointsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List private link endpoints under a project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID.",
				Required:            true,
			},
			"current_page": schema.Int64Attribute{
				MarkdownDescription: "Page number (defaults to 1).",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Page size (1-100, defaults to 10).",
				Optional:            true,
			},
			"count": schema.Int64Attribute{
				MarkdownDescription: "Total count of endpoints.",
				Computed:            true,
			},
			"endpoints": schema.ListNestedAttribute{
				MarkdownDescription: "List of endpoints.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"region_id":               schema.StringAttribute{Computed: true},
						"cloud_id":                schema.StringAttribute{Computed: true},
						"endpoint_service":        schema.StringAttribute{Computed: true},
						"endpoint_service_status": schema.StringAttribute{Computed: true},
						"endpoint_id":             schema.StringAttribute{Computed: true},
						"endpoint_status":         schema.StringAttribute{Computed: true},
						"gcp_project_id":          schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *EndpointsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *EndpointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EndpointsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentPage := 1
	if !state.CurrentPage.IsNull() {
		currentPage = int(state.CurrentPage.ValueInt64())
	}
	pageSize := 10
	if !state.PageSize.IsNull() {
		pageSize = int(state.PageSize.ValueInt64())
	}

	eps, page, err := d.client.ListEndpoints(state.ProjectId.ValueString(), currentPage, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to ListEndpoints, got error: %s", err))
		return
	}

	state.Endpoints = nil
	for _, e := range eps {
		gcp := types.StringNull()
		if e.GcpProjectId != nil {
			gcp = types.StringValue(*e.GcpProjectId)
		}
		state.Endpoints = append(state.Endpoints, EndpointItem{
			RegionId:              types.StringValue(e.RegionId),
			CloudId:               types.StringValue(e.CloudId),
			EndpointService:       types.StringValue(e.EndpointService),
			EndpointServiceStatus: types.StringValue(e.EndpointServiceStatus),
			EndpointId:            types.StringValue(e.EndpointId),
			EndpointStatus:        types.StringValue(e.EndpointStatus),
			GcpProjectId:          gcp,
		})
	}
	state.Count = types.Int64Value(int64(page.Count))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/provider/endpoints_data_source.go
git commit -m "feat(provider): add zillizcloud_endpoints data source"
```

---

## Task 5: `zillizcloud_endpoint` resource

**Files:**
- Create: `internal/provider/endpoint_resource.go`

Behavior recap (from spec):
- Create: POST endpoint, then call ListEndpoints to refresh computed status fields.
- Read: scan pages of ListEndpoints looking for matching `endpoint_id`; remove from state if not found.
- Update: no-op (all user inputs `RequiresReplace`).
- Delete: DELETE endpoint with region_id and optional gcp_project_id.
- No import in v1.

- [ ] **Step 1: Create the resource file**

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &EndpointResource{}
var _ resource.ResourceWithConfigure = &EndpointResource{}

func NewEndpointResource() resource.Resource {
	return &EndpointResource{}
}

type EndpointResource struct {
	client *zilliz.Client
}

type EndpointResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	ProjectId             types.String `tfsdk:"project_id"`
	RegionId              types.String `tfsdk:"region_id"`
	EndpointId            types.String `tfsdk:"endpoint_id"`
	GcpProjectId          types.String `tfsdk:"gcp_project_id"`
	CloudId               types.String `tfsdk:"cloud_id"`
	EndpointService       types.String `tfsdk:"endpoint_service"`
	EndpointServiceStatus types.String `tfsdk:"endpoint_service_status"`
	EndpointStatus        types.String `tfsdk:"endpoint_status"`
}

func (r *EndpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r *EndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a private link endpoint for a Zilliz Cloud project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"endpoint_id": schema.StringAttribute{
				MarkdownDescription: "VPC endpoint ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"gcp_project_id": schema.StringAttribute{
				MarkdownDescription: "GCP project ID (required for GCP regions).",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cloud_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_service": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_service_status": schema.StringAttribute{
				Computed: true,
			},
			"endpoint_status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *EndpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

// findEndpoint scans all pages of ListEndpoints looking for endpointId. Returns nil if not found.
func (r *EndpointResource) findEndpoint(projectId, endpointId string) (*zilliz.Endpoint, error) {
	const pageSize = 100
	page := 1
	for {
		eps, pg, err := r.client.ListEndpoints(projectId, page, pageSize)
		if err != nil {
			return nil, err
		}
		for i := range eps {
			if eps[i].EndpointId == endpointId {
				return &eps[i], nil
			}
		}
		if page*pageSize >= pg.Count || len(eps) == 0 {
			return nil, nil
		}
		page++
	}
}

func (r *EndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &zilliz.CreateEndpointRequest{
		RegionId:   data.RegionId.ValueString(),
		EndpointId: data.EndpointId.ValueString(),
	}
	if !data.GcpProjectId.IsNull() && !data.GcpProjectId.IsUnknown() {
		body.GcpProjectId = data.GcpProjectId.ValueString()
	}

	created, err := r.client.CreateEndpoint(data.ProjectId.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint",
			fmt.Sprintf("projectId=%s endpointId=%s error=%s",
				data.ProjectId.ValueString(), data.EndpointId.ValueString(), err))
		return
	}

	data.Id = types.StringValue(created.EndpointId)

	// Refresh computed status by finding the endpoint.
	ep, err := r.findEndpoint(data.ProjectId.ValueString(), created.EndpointId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read endpoint after create", err.Error())
		return
	}
	if ep != nil {
		data.CloudId = types.StringValue(ep.CloudId)
		data.EndpointService = types.StringValue(ep.EndpointService)
		data.EndpointServiceStatus = types.StringValue(ep.EndpointServiceStatus)
		data.EndpointStatus = types.StringValue(ep.EndpointStatus)
	} else {
		data.CloudId = types.StringNull()
		data.EndpointService = types.StringNull()
		data.EndpointServiceStatus = types.StringNull()
		data.EndpointStatus = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ep, err := r.findEndpoint(state.ProjectId.ValueString(), state.EndpointId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list endpoints", err.Error())
		return
	}
	if ep == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.CloudId = types.StringValue(ep.CloudId)
	state.EndpointService = types.StringValue(ep.EndpointService)
	state.EndpointServiceStatus = types.StringValue(ep.EndpointServiceStatus)
	state.EndpointStatus = types.StringValue(ep.EndpointStatus)
	state.RegionId = types.StringValue(ep.RegionId)
	if ep.GcpProjectId != nil {
		state.GcpProjectId = types.StringValue(*ep.GcpProjectId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *EndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All user-supplied attributes are RequiresReplace; Update is unreachable in practice,
	// but the framework requires the method. Pass plan to state unchanged.
	var plan EndpointResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EndpointResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var gcp *string
	if !state.GcpProjectId.IsNull() && state.GcpProjectId.ValueString() != "" {
		v := state.GcpProjectId.ValueString()
		gcp = &v
	}

	err := r.client.DeleteEndpoint(
		state.ProjectId.ValueString(),
		state.EndpointId.ValueString(),
		state.RegionId.ValueString(),
		gcp,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete endpoint",
			fmt.Sprintf("projectId=%s endpointId=%s error=%s",
				state.ProjectId.ValueString(), state.EndpointId.ValueString(), err))
		return
	}
}

// Keep path import to silence unused-import if compiler complains (remove if not needed).
var _ = path.Root
```

(Note: `path` import is only needed if the framework version requires it; if `go build` errors with "imported and not used: path", remove the `path` import and the `var _ = path.Root` sentinel.)

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success. If compiler complains about unused `path` import, remove `"github.com/hashicorp/terraform-plugin-framework/path"` and the trailing `var _ = path.Root` line, then rebuild.

- [ ] **Step 3: Commit**

```bash
git add internal/provider/endpoint_resource.go
git commit -m "feat(provider): add zillizcloud_endpoint resource"
```

---

## Task 6: `zillizcloud_endpoint_whitelist` resource

**Files:**
- Create: `internal/provider/endpoint_whitelist_resource.go`

Behavior recap: Create-only. Read/Update/Delete are no-ops. ID = project_id. All fields RequiresReplace.

- [ ] **Step 1: Create the resource file**

```go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var _ resource.Resource = &EndpointWhitelistResource{}
var _ resource.ResourceWithConfigure = &EndpointWhitelistResource{}

func NewEndpointWhitelistResource() resource.Resource {
	return &EndpointWhitelistResource{}
}

type EndpointWhitelistResource struct {
	client *zilliz.Client
}

type EndpointWhitelistResourceModel struct {
	Id          types.String `tfsdk:"id"`
	ProjectId   types.String `tfsdk:"project_id"`
	RegionId    types.String `tfsdk:"region_id"`
	OuterUserId types.String `tfsdk:"outer_user_id"`
}

func (r *EndpointWhitelistResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint_whitelist"
}

func (r *EndpointWhitelistResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Adds an external cloud account to the private link endpoint whitelist.

**Limitations:**
- The upstream API only supports adding whitelist entries; there is no list or delete API.
- **Read is a no-op** — drift cannot be detected.
- **Destroy is a no-op** — the whitelist entry is *not* removed from Zilliz Cloud; only from Terraform state.
- Because the resource ID is set to ` + "`project_id`" + `, at most one whitelist resource may be declared per project.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID. Also serves as the resource ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Cloud region ID.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"outer_user_id": schema.StringAttribute{
				MarkdownDescription: "External cloud account identifier (Azure Subscription ID, Alibaba/Tencent/Huawei Account ID).",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *EndpointWhitelistResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*zilliz.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider type",
			fmt.Sprintf("Expected *zilliz.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *EndpointWhitelistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.AddEndpointWhitelist(data.ProjectId.ValueString(), &zilliz.AddEndpointWhitelistRequest{
		RegionId:    data.RegionId.ValueString(),
		OuterUserId: data.OuterUserId.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to add endpoint whitelist",
			fmt.Sprintf("projectId=%s error=%s", data.ProjectId.ValueString(), err))
		return
	}

	data.Id = data.ProjectId
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read is a no-op — the upstream API does not expose a GET for whitelist entries.
func (r *EndpointWhitelistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable — all attributes are RequiresReplace.
func (r *EndpointWhitelistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EndpointWhitelistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op — the upstream API does not expose a DELETE for whitelist entries.
// The entry remains on the server; only Terraform state is cleared.
func (r *EndpointWhitelistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add internal/provider/endpoint_whitelist_resource.go
git commit -m "feat(provider): add zillizcloud_endpoint_whitelist resource"
```

---

## Task 7: Register in provider.go

**Files:**
- Modify: `internal/provider/provider.go`

- [ ] **Step 1: Add two entries to `Resources()` (line ~161)**

In `Resources()` return list, before the final `}`, add:

```go
		NewEndpointResource,
		NewEndpointWhitelistResource,
```

Insert after `NewBackupPolicyResource,` so the block reads:

```go
		NewBackupPolicyResource,
		NewEndpointResource,
		NewEndpointWhitelistResource,
	}
```

- [ ] **Step 2: Add two entries to `DataSources()` (line ~182)**

After `NewPartitionsDataSource,`, add:

```go
		NewEndpointServicesDataSource,
		NewEndpointsDataSource,
```

So the block reads:

```go
		NewPartitionsDataSource,
		NewEndpointServicesDataSource,
		NewEndpointsDataSource,
	}
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 4: Run full unit tests**

Run: `make unit-test`
Expected: all pass.

- [ ] **Step 5: Commit**

```bash
git add internal/provider/provider.go
git commit -m "feat(provider): register private link resources and data sources"
```

---

## Task 8: Examples and generated docs

**Files:**
- Create: `examples/data-sources/zillizcloud_endpoint_services/data-source.tf`
- Create: `examples/data-sources/zillizcloud_endpoints/data-source.tf`
- Create: `examples/resources/zillizcloud_endpoint/resource.tf`
- Create: `examples/resources/zillizcloud_endpoint_whitelist/resource.tf`

- [ ] **Step 1: Create example for endpoint_services data source**

`examples/data-sources/zillizcloud_endpoint_services/data-source.tf`:

```hcl
data "zillizcloud_endpoint_services" "aws_usw2" {
  region_id = "aws-us-west-2"
}

output "available_services" {
  value = data.zillizcloud_endpoint_services.aws_usw2.endpoint_services
}
```

- [ ] **Step 2: Create example for endpoints data source**

`examples/data-sources/zillizcloud_endpoints/data-source.tf`:

```hcl
data "zillizcloud_endpoints" "mine" {
  project_id = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
}

output "endpoints" {
  value = data.zillizcloud_endpoints.mine.endpoints
}
```

- [ ] **Step 3: Create example for endpoint resource**

`examples/resources/zillizcloud_endpoint/resource.tf`:

```hcl
resource "zillizcloud_endpoint" "aws" {
  project_id  = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id   = "aws-us-west-2"
  endpoint_id = "vpce-072eaf2b4a747c24f"
}

# GCP example — gcp_project_id is required
resource "zillizcloud_endpoint" "gcp" {
  project_id     = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id      = "gcp-us-west1"
  endpoint_id    = "my-psc-endpoint"
  gcp_project_id = "my-gcp-project"
}
```

- [ ] **Step 4: Create example for endpoint_whitelist resource**

`examples/resources/zillizcloud_endpoint_whitelist/resource.tf`:

```hcl
# NOTE: Destroy is a no-op - the whitelist entry is not removed from Zilliz Cloud
# on `terraform destroy`. Manual cleanup via console or support is required.
resource "zillizcloud_endpoint_whitelist" "azure" {
  project_id    = "proj-xxxxxxxxxxxxxxxxxxxxxxxx"
  region_id     = "azure-eastus2"
  outer_user_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

- [ ] **Step 5: Generate docs**

Run: `make doc` (or `go generate ./...`)
Expected: new files created in `docs/resources/endpoint.md`, `docs/resources/endpoint_whitelist.md`, `docs/data-sources/endpoint_services.md`, `docs/data-sources/endpoints.md`.

- [ ] **Step 6: Run lint**

Run: `make lint`
Expected: no errors. If there are formatting issues, run `make fmt` then re-run lint.

- [ ] **Step 7: Commit**

```bash
git add examples/ docs/
git commit -m "docs: add examples and generated docs for private link resources"
```

---

## Self-Review Notes

- **Spec coverage:** All 5 API endpoints are covered in Task 1 (client). Both data sources covered in Tasks 3–4. Both resources covered in Tasks 5–6. Registration in Task 7. Examples + docs in Task 8. ✓
- **Placeholder scan:** No TBDs, TODOs, or "implement later". ✓
- **Type consistency:** Method names match between client (Task 1) and callers (Tasks 3–6): `ListEndpointServices`, `ListEndpoints`, `CreateEndpoint`, `DeleteEndpoint`, `AddEndpointWhitelist`. Types: `EndpointService`, `Endpoint`, `CreateEndpointRequest`, `CreateEndpointResponse`, `AddEndpointWhitelistRequest`. ✓
- **Known limitation documented:** `endpoint_whitelist` ID collision and no-op Read/Delete called out in the resource's MarkdownDescription and in the example. ✓

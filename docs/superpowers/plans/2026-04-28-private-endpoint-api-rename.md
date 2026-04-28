# Private Endpoint API Rename Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the recently added PrivateLink endpoint support to the new private-prefixed API paths and Terraform type names, with old Terraform names removed.

**Architecture:** Keep the existing endpoint client and provider implementation shape. Change the HTTP paths in `client/endpoint.go`, change Terraform type metadata in the provider resources/data sources, then regenerate docs from private-prefixed examples. Add focused tests for exact HTTP paths and Terraform type names.

**Tech Stack:** Go, Terraform Plugin Framework, `go test`, `go generate ./...`, `tfplugindocs`.

---

## File Structure

- Modify `client/endpoint.go`: change only the five endpoint HTTP paths to the new upstream private-prefixed paths.
- Modify `client/endpoint_test.go`: assert exact request paths so old paths cannot pass by substring.
- Create `internal/provider/private_endpoint_metadata_test.go`: verify the Terraform type names exposed by provider metadata.
- Modify `internal/provider/endpoint_services_data_source.go`: expose `zillizcloud_private_endpoint_services`.
- Modify `internal/provider/endpoints_data_source.go`: expose `zillizcloud_private_endpoints`.
- Modify `internal/provider/endpoint_resource.go`: expose `zillizcloud_private_endpoint`.
- Modify `internal/provider/endpoint_whitelist_resource.go`: expose `zillizcloud_private_endpoint_whitelist`.
- Move and modify example directories:
  - `examples/data-sources/zillizcloud_endpoint_services` to `examples/data-sources/zillizcloud_private_endpoint_services`
  - `examples/data-sources/zillizcloud_endpoints` to `examples/data-sources/zillizcloud_private_endpoints`
  - `examples/resources/zillizcloud_endpoint` to `examples/resources/zillizcloud_private_endpoint`
  - `examples/resources/zillizcloud_endpoint_whitelist` to `examples/resources/zillizcloud_private_endpoint_whitelist`
- Modify `docs/guides/aws-privatelink.md`, `templates/guides/aws-privatelink.md`, and `new-client-api.md`: replace active user-facing old names and old API paths.
- Regenerate provider docs with `go generate ./...`, then remove stale generated docs for old Terraform type names if the generator leaves them behind.

### Task 1: Client API Paths

**Files:**
- Modify: `client/endpoint_test.go`
- Modify: `client/endpoint.go`

- [ ] **Step 1: Strengthen the client path tests**

Edit `client/endpoint_test.go` so the endpoint tests check exact paths. The relevant assertions should use these blocks:

```go
func TestUnitListEndpointServices(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/privateEndpointServices" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.Query().Get("regionId") != "aws-us-west-2" {
			t.Errorf("regionId=%s", req.URL.Query().Get("regionId"))
		}
		return jsonResponse(t, map[string]any{
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
```

```go
func TestUnitListEndpoints(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v2/projects/proj-1/privateEndpoints" {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, map[string]any{
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
```

In `TestUnitCreateEndpoint`, add this path assertion after the method assertion:

```go
if req.URL.Path != "/v2/projects/proj-1/privateEndpoints" {
	t.Errorf("path=%s", req.URL.Path)
}
```

In both `TestUnitDeleteEndpoint` subtests, add this path assertion inside the mock handler:

```go
if req.URL.Path != "/v2/projects/proj-1/privateEndpoints/vpce-abc" {
	t.Errorf("path=%s", req.URL.Path)
}
```

In `TestUnitAddEndpointWhitelist`, replace the suffix check with:

```go
if req.URL.Path != "/v2/projects/proj-1/privateEndpointWhitelist" {
	t.Errorf("path=%s", req.URL.Path)
}
```

- [ ] **Step 2: Run the focused client tests and verify they fail**

Run:

```bash
go test ./client -run 'TestUnit(ListEndpointServices|ListEndpoints|CreateEndpoint|DeleteEndpoint|AddEndpointWhitelist)' -count=1
```

Expected: FAIL. The failure should show old paths such as `/v2/endpointServices`, `/v2/projects/proj-1/endpoints`, or `/v2/projects/proj-1/endpointWhitelist`.

- [ ] **Step 3: Update `client/endpoint.go` to call the private API paths**

Replace the five `c.do` path strings in `client/endpoint.go` with this implementation:

```go
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
	err := c.do("GET", "privateEndpointServices?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.EndpointServices, response.Data.zillizPage, nil
}

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
	err := c.do("GET", "projects/"+projectId+"/privateEndpoints?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.Endpoints, response.Data.zillizPage, nil
}

func (c *Client) CreateEndpoint(projectId string, req *CreateEndpointRequest) (*CreateEndpointResponse, error) {
	var response zillizResponse[CreateEndpointResponse]
	err := c.do("POST", "projects/"+projectId+"/privateEndpoints", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteEndpoint(projectId, endpointId, regionId string, gcpProjectId *string) error {
	q := url.Values{}
	q.Set("regionId", regionId)
	if gcpProjectId != nil && *gcpProjectId != "" {
		q.Set("gcpProjectId", *gcpProjectId)
	}
	var response zillizResponse[map[string]any]
	return c.do("DELETE", "projects/"+projectId+"/privateEndpoints/"+endpointId+"?"+q.Encode(), nil, &response)
}

func (c *Client) AddEndpointWhitelist(projectId string, req *AddEndpointWhitelistRequest) error {
	var response zillizResponse[string]
	return c.do("POST", "projects/"+projectId+"/privateEndpointWhitelist", req, &response)
}
```

- [ ] **Step 4: Run formatting and focused client tests**

Run:

```bash
gofmt -w client/endpoint.go client/endpoint_test.go
go test ./client -run 'TestUnit(ListEndpointServices|ListEndpoints|CreateEndpoint|DeleteEndpoint|AddEndpointWhitelist)' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the client path change**

Run:

```bash
git add client/endpoint.go client/endpoint_test.go
git commit -m "fix(client): use private endpoint API paths"
```

### Task 2: Terraform Type Names

**Files:**
- Create: `internal/provider/private_endpoint_metadata_test.go`
- Modify: `internal/provider/endpoint_services_data_source.go`
- Modify: `internal/provider/endpoints_data_source.go`
- Modify: `internal/provider/endpoint_resource.go`
- Modify: `internal/provider/endpoint_whitelist_resource.go`

- [ ] **Step 1: Add a failing provider metadata test**

Create `internal/provider/private_endpoint_metadata_test.go`:

```go
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPrivateEndpointTerraformTypeNames(t *testing.T) {
	ctx := context.Background()

	dataSourceCases := []struct {
		name    string
		factory func() datasource.DataSource
		want    string
	}{
		{
			name:    "endpoint services",
			factory: NewEndpointServicesDataSource,
			want:    "zillizcloud_private_endpoint_services",
		},
		{
			name:    "endpoints",
			factory: NewEndpointsDataSource,
			want:    "zillizcloud_private_endpoints",
		},
	}

	for _, tc := range dataSourceCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := tc.factory()
			var resp datasource.MetadataResponse
			ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
			if resp.TypeName != tc.want {
				t.Fatalf("TypeName=%q, want %q", resp.TypeName, tc.want)
			}
		})
	}

	resourceCases := []struct {
		name    string
		factory func() resource.Resource
		want    string
	}{
		{
			name:    "endpoint",
			factory: NewEndpointResource,
			want:    "zillizcloud_private_endpoint",
		},
		{
			name:    "endpoint whitelist",
			factory: NewEndpointWhitelistResource,
			want:    "zillizcloud_private_endpoint_whitelist",
		},
	}

	for _, tc := range resourceCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.factory()
			var resp resource.MetadataResponse
			res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
			if resp.TypeName != tc.want {
				t.Fatalf("TypeName=%q, want %q", resp.TypeName, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run the focused provider metadata test and verify it fails**

Run:

```bash
go test ./internal/provider -run TestPrivateEndpointTerraformTypeNames -count=1
```

Expected: FAIL. The failures should show old type names such as `zillizcloud_endpoint_services`.

- [ ] **Step 3: Update provider metadata suffixes**

In `internal/provider/endpoint_services_data_source.go`, change:

```go
resp.TypeName = req.ProviderTypeName + "_endpoint_services"
```

to:

```go
resp.TypeName = req.ProviderTypeName + "_private_endpoint_services"
```

In `internal/provider/endpoints_data_source.go`, change:

```go
resp.TypeName = req.ProviderTypeName + "_endpoints"
```

to:

```go
resp.TypeName = req.ProviderTypeName + "_private_endpoints"
```

In `internal/provider/endpoint_resource.go`, change:

```go
resp.TypeName = req.ProviderTypeName + "_endpoint"
```

to:

```go
resp.TypeName = req.ProviderTypeName + "_private_endpoint"
```

In `internal/provider/endpoint_whitelist_resource.go`, change:

```go
resp.TypeName = req.ProviderTypeName + "_endpoint_whitelist"
```

to:

```go
resp.TypeName = req.ProviderTypeName + "_private_endpoint_whitelist"
```

- [ ] **Step 4: Run formatting and provider tests**

Run:

```bash
gofmt -w internal/provider/private_endpoint_metadata_test.go internal/provider/endpoint_services_data_source.go internal/provider/endpoints_data_source.go internal/provider/endpoint_resource.go internal/provider/endpoint_whitelist_resource.go
go test ./internal/provider -run TestPrivateEndpointTerraformTypeNames -count=1
go test ./internal/provider/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the provider type name change**

Run:

```bash
git add internal/provider/private_endpoint_metadata_test.go internal/provider/endpoint_services_data_source.go internal/provider/endpoints_data_source.go internal/provider/endpoint_resource.go internal/provider/endpoint_whitelist_resource.go
git commit -m "fix(provider): expose private endpoint terraform types"
```

### Task 3: Examples, API Doc, And Generated Docs

**Files:**
- Move: `examples/data-sources/zillizcloud_endpoint_services` to `examples/data-sources/zillizcloud_private_endpoint_services`
- Move: `examples/data-sources/zillizcloud_endpoints` to `examples/data-sources/zillizcloud_private_endpoints`
- Move: `examples/resources/zillizcloud_endpoint` to `examples/resources/zillizcloud_private_endpoint`
- Move: `examples/resources/zillizcloud_endpoint_whitelist` to `examples/resources/zillizcloud_private_endpoint_whitelist`
- Modify: `examples/guides/aws-privatelink/main.tf`
- Modify: `docs/guides/aws-privatelink.md`
- Modify: `templates/guides/aws-privatelink.md`
- Modify: `new-client-api.md`
- Delete after regeneration if still present: `docs/data-sources/endpoint_services.md`
- Delete after regeneration if still present: `docs/data-sources/endpoints.md`
- Delete after regeneration if still present: `docs/resources/endpoint.md`
- Delete after regeneration if still present: `docs/resources/endpoint_whitelist.md`

- [ ] **Step 1: Move the examples to private-prefixed directories**

Run:

```bash
git mv examples/data-sources/zillizcloud_endpoint_services examples/data-sources/zillizcloud_private_endpoint_services
git mv examples/data-sources/zillizcloud_endpoints examples/data-sources/zillizcloud_private_endpoints
git mv examples/resources/zillizcloud_endpoint examples/resources/zillizcloud_private_endpoint
git mv examples/resources/zillizcloud_endpoint_whitelist examples/resources/zillizcloud_private_endpoint_whitelist
```

- [ ] **Step 2: Rewrite active Terraform names and API paths outside historical specs/plans**

Run this mechanical rewrite on active user-facing files:

```bash
perl -pi -e 's/zillizcloud_endpoint_services/zillizcloud_private_endpoint_services/g; s/zillizcloud_endpoints/zillizcloud_private_endpoints/g; s/zillizcloud_endpoint_whitelist/zillizcloud_private_endpoint_whitelist/g; s/zillizcloud_endpoint/zillizcloud_private_endpoint/g' \
  examples/data-sources/zillizcloud_private_endpoint_services/data-source.tf \
  examples/data-sources/zillizcloud_private_endpoints/data-source.tf \
  examples/resources/zillizcloud_private_endpoint/resource.tf \
  examples/resources/zillizcloud_private_endpoint_whitelist/resource.tf \
  examples/guides/aws-privatelink/main.tf \
  docs/guides/aws-privatelink.md \
  templates/guides/aws-privatelink.md

perl -pi -e 's#/v2/endpointServices#/v2/privateEndpointServices#g; s#/v2/projects/\\{projectId\\}/endpoints#/v2/projects/{projectId}/privateEndpoints#g; s#/v2/projects/\\{projectId\\}/endpointWhitelist#/v2/projects/{projectId}/privateEndpointWhitelist#g' new-client-api.md
```

- [ ] **Step 3: Run docs generation**

Run:

```bash
go generate ./...
```

Expected: PASS. This runs `terraform fmt -recursive ./examples/` and `tfplugindocs`.

- [ ] **Step 4: Remove stale generated docs for old Terraform names if present**

Run:

```bash
rm -f docs/data-sources/endpoint_services.md docs/data-sources/endpoints.md docs/resources/endpoint.md docs/resources/endpoint_whitelist.md
```

Expected: new generated files should exist:

```bash
test -f docs/data-sources/private_endpoint_services.md
test -f docs/data-sources/private_endpoints.md
test -f docs/resources/private_endpoint.md
test -f docs/resources/private_endpoint_whitelist.md
```

- [ ] **Step 5: Verify no active old names or old paths remain**

Run:

```bash
rg -n 'zillizcloud_endpoint_services|zillizcloud_endpoints|zillizcloud_endpoint_whitelist|zillizcloud_endpoint' \
  client internal docs examples templates new-client-api.md \
  -g '!docs/superpowers/**'

rg -n '/v2/endpointServices|/v2/projects/\\{projectId\\}/endpoints|/v2/projects/\\{projectId\\}/endpointWhitelist|\"endpointServices\\?|\"projects/\"\\+projectId\\+\"/endpoints|\"projects/\"\\+projectId\\+\"/endpointWhitelist' \
  client internal docs examples templates new-client-api.md \
  -g '!docs/superpowers/**'
```

Expected: no output from both commands. If the first command reports schema attribute names such as `endpoint_service`, keep them; rerun with this narrower expression instead:

```bash
rg -n 'zillizcloud_endpoint_services|zillizcloud_endpoints|zillizcloud_endpoint_whitelist|zillizcloud_endpoint' \
  docs examples templates \
  -g '!docs/superpowers/**'
```

Expected: no output.

- [ ] **Step 6: Commit examples and docs**

Run:

```bash
git add examples docs templates new-client-api.md
git commit -m "docs: rename endpoint examples to private endpoint"
```

### Task 4: Final Verification

**Files:**
- Verify all files changed by Tasks 1-3.

- [ ] **Step 1: Run focused tests**

Run:

```bash
go test ./client -count=1
go test ./internal/provider/... -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full unit tests if practical**

Run:

```bash
go test ./... -count=1
```

Expected: PASS. If this fails because acceptance-style tests require external services or local credentials, record the failing package and rerun the focused tests from Step 1.

- [ ] **Step 3: Run final docs generation check**

Run:

```bash
go generate ./...
git diff --exit-code
```

Expected: PASS and no diff. If `go generate ./...` changes files, inspect the diff, commit generated changes, then rerun this step.

- [ ] **Step 4: Run final cleanup searches**

Run:

```bash
rg -n 'zillizcloud_endpoint_services|zillizcloud_endpoints|zillizcloud_endpoint_whitelist|zillizcloud_endpoint' \
  docs examples templates \
  -g '!docs/superpowers/**'

rg -n 'endpointServices\\?|projects/\"\\+projectId\\+\"/endpoints|projects/\"\\+projectId\\+\"/endpointWhitelist|/v2/endpointServices|/v2/projects/\\{projectId\\}/endpoints|/v2/projects/\\{projectId\\}/endpointWhitelist' \
  client new-client-api.md \
  -g '!docs/superpowers/**'
```

Expected: no output.

- [ ] **Step 5: Confirm git status**

Run:

```bash
git status --short
```

Expected: no uncommitted files from this implementation. Existing unrelated untracked files such as `.agents/`, `examples/guides/aws-privatelink/.envrc`, or other user-created files may remain and should not be modified or removed.

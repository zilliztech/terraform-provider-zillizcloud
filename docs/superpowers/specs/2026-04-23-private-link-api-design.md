# Private Link API — Terraform Provider Design

**Date:** 2026-04-23
**Scope:** Add Terraform provider support for the new Private Link Management API (5 endpoints).

## Goals

Expose Zilliz Cloud Private Link management through the Terraform provider:

- List available endpoint services (data source)
- List endpoints under a project (data source)
- Manage VPC endpoints lifecycle (resource)
- Add external cloud account to endpoint whitelist (resource, minimal)

## Non-Goals

- Full CRUD for endpoint whitelist — the upstream API only supports POST (Add). List/Get/Delete will be revisited when upstream provides those APIs.
- Import support for the endpoint resource in v1.

## API Summary

See `new-client-api.md` for the authoritative spec. Five endpoints:

| Method | Path | Purpose |
|---|---|---|
| GET | `/v2/endpointServices?regionId=` | List endpoint services |
| GET | `/v2/projects/{projectId}/endpoints` | List endpoints |
| POST | `/v2/projects/{projectId}/endpoints` | Create endpoint |
| DELETE | `/v2/projects/{projectId}/endpoints/{endpointId}` | Delete endpoint (regionId and optional gcpProjectId as **query params**) |
| POST | `/v2/projects/{projectId}/endpointWhitelist` | Add whitelist entry |

## Client Layer

New file: `client/endpoint.go`.

### Types

```go
type EndpointService struct {
    RegionId          string `json:"regionId"`
    CloudId           string `json:"cloudId"`
    EndpointService   string `json:"endpointService"`
    WhitelistRequired bool   `json:"whitelistRequired"`
}

type Endpoint struct {
    RegionId              string  `json:"regionId"`
    CloudId               string  `json:"cloudId"`
    EndpointService       string  `json:"endpointService"`
    EndpointServiceStatus string  `json:"endpointServiceStatus"`
    EndpointId            string  `json:"endpointId"`
    EndpointStatus        string  `json:"endpointStatus"`
    GcpProjectId          *string `json:"gcpProjectId"` // pointer: API returns null for non-GCP
}

type CreateEndpointRequest struct {
    RegionId     string `json:"regionId"`
    EndpointId   string `json:"endpointId"`
    GcpProjectId string `json:"gcpProjectId,omitempty"`
}

type CreateEndpointResponse struct {
    EndpointId string `json:"endpointId"`
    RegionId   string `json:"regionId"`
}

type AddEndpointWhitelistRequest struct {
    RegionId    string `json:"regionId"`
    OuterUserId string `json:"outerUserId"`
}
```

### Methods on `*Client`

```go
ListEndpointServices(regionId string, currentPage, pageSize int) ([]EndpointService, zillizPage, error)
ListEndpoints(projectId string, currentPage, pageSize int) ([]Endpoint, zillizPage, error)
CreateEndpoint(projectId string, req *CreateEndpointRequest) (*CreateEndpointResponse, error)
DeleteEndpoint(projectId, endpointId, regionId string, gcpProjectId *string) error
AddEndpointWhitelist(projectId string, req *AddEndpointWhitelistRequest) error
```

Implementation notes:

- Use the existing `c.do(method, path, body, &response)` pattern.
- For paginated GETs, construct paginated responses with a nested `zillizPage` the same way existing code does (see `client/zilliz.go:288`).
- `DeleteEndpoint` builds query string: `?regionId=<>[&gcpProjectId=<>]`.
- All methods follow existing error-handling conventions in `client/zilliz.go`.

## Data Sources

### `zillizcloud_endpoint_services`

File: `internal/provider/endpoint_services_data_source.go`

| Attribute | Type | Required | Description |
|---|---|---|---|
| `region_id` | string | yes | Cloud region ID |
| `current_page` | int | no (default 1) | Page number |
| `page_size` | int | no (default 10) | Page size (1-100) |
| `endpoint_services` | list[object] | computed | see below |
| `count` | int | computed | Total count |

Nested object: `region_id`, `cloud_id`, `endpoint_service`, `whitelist_required`.

### `zillizcloud_endpoints`

File: `internal/provider/endpoints_data_source.go`

| Attribute | Type | Required | Description |
|---|---|---|---|
| `project_id` | string | yes | Project ID |
| `current_page` | int | no (default 1) | Page number |
| `page_size` | int | no (default 10) | Page size |
| `endpoints` | list[object] | computed | see below |
| `count` | int | computed | Total count |

Nested object: `region_id`, `cloud_id`, `endpoint_service`, `endpoint_service_status`, `endpoint_id`, `endpoint_status`, `gcp_project_id`.

## Resource: `zillizcloud_endpoint`

File: `internal/provider/endpoint_resource.go`

### Schema

| Attribute | Type | Required | Plan modifier |
|---|---|---|---|
| `project_id` | string | yes | RequiresReplace |
| `region_id` | string | yes | RequiresReplace |
| `endpoint_id` | string | yes | RequiresReplace |
| `gcp_project_id` | string | optional | RequiresReplace |
| `id` | string | computed | — |
| `cloud_id` | string | computed | — |
| `endpoint_service` | string | computed | — |
| `endpoint_service_status` | string | computed | — |
| `endpoint_status` | string | computed | — |

### Lifecycle

- **Create:** `CreateEndpoint(projectId, {regionId, endpointId, gcpProjectId})`. After success, set `id = endpoint_id` and call `ListEndpoints` to refresh computed status fields.
- **Read:** `ListEndpoints(projectId, ...)`, scan for matching `endpoint_id`. If not found → remove from state. Otherwise refresh computed fields.
  - If the list is paginated and the endpoint could be on a later page, iterate pages until found or exhausted.
- **Update:** no-op (all user-supplied fields are RequiresReplace).
- **Delete:** `DeleteEndpoint(projectId, endpointId, regionId, gcpProjectId)`. Pass `gcp_project_id` if present.
- **Import:** not supported in v1.
- **GCP validation:** not performed client-side. API error surfaces if `gcp_project_id` is missing for a GCP region (per user decision 2b:B).

## Resource: `zillizcloud_endpoint_whitelist`

File: `internal/provider/endpoint_whitelist_resource.go`

Minimal best-effort resource — the upstream API only supports Add.

### Schema

| Attribute | Type | Required | Plan modifier |
|---|---|---|---|
| `project_id` | string | yes | RequiresReplace |
| `region_id` | string | yes | RequiresReplace |
| `outer_user_id` | string | yes | RequiresReplace |
| `id` | string | computed (= `project_id`) | — |

### Lifecycle

- **Create:** `AddEndpointWhitelist(projectId, {regionId, outerUserId})`. Set `id = project_id`.
- **Read:** no-op. No GET API available.
- **Update:** unreachable (all fields RequiresReplace).
- **Delete:** no-op. Just remove from state.

### Known limitation (documented in generated docs)

Because `id == project_id`, declaring multiple `zillizcloud_endpoint_whitelist` blocks for the same `project_id` produces colliding resource IDs. Users must manage at most one whitelist resource per project. This is acceptable for v1 given the minimal API surface; can be revisited when list/delete APIs are available.

## Provider Registration

In `internal/provider/provider.go`:

- Append `NewEndpointResource`, `NewEndpointWhitelistResource` to `Resources()`.
- Append `NewEndpointServicesDataSource`, `NewEndpointsDataSource` to `DataSources()`.

## Examples

Add:

- `examples/data-sources/zillizcloud_endpoint_services/data-source.tf`
- `examples/data-sources/zillizcloud_endpoints/data-source.tf`
- `examples/resources/zillizcloud_endpoint/resource.tf`
- `examples/resources/zillizcloud_endpoint_whitelist/resource.tf`

Run `go generate ./...` to produce `docs/resources/endpoint.md`, etc.

## Testing

- Unit tests for client methods using `client/testdata/` fixtures, following the pattern in `client/cluster_test.go` / `client/project_test.go`.
- Acceptance tests optional for v1 (creating real VPC endpoints requires cloud-side setup). Provide at minimum a compile-checking test for each resource/data-source.

## Out of Scope

- Automatic retry semantics beyond what `client/retry` already provides.
- Whitelist list/delete — blocked on upstream API.
- Endpoint import — deferred.

# Private Endpoint API Rename Design

**Date:** 2026-04-28
**Scope:** Update the recently added PrivateLink endpoint support after the upstream API paths were renamed with a `private` prefix.

## Goals

- Switch the client to the new upstream API paths:
  - `GET /v2/privateEndpointServices`
  - `GET /v2/projects/{projectId}/privateEndpoints`
  - `POST /v2/projects/{projectId}/privateEndpoints`
  - `DELETE /v2/projects/{projectId}/privateEndpoints/{endpointId}`
  - `POST /v2/projects/{projectId}/privateEndpointWhitelist`
- Rename the Terraform-facing resources and data sources to include `private`.
- Remove the old Terraform-facing names entirely. No deprecated aliases are kept.
- Keep the current endpoint lifecycle behavior, schema attributes, pagination, and whitelist limitations unchanged.

## Non-Goals

- Do not add compatibility aliases for `zillizcloud_endpoint_services`, `zillizcloud_endpoints`, `zillizcloud_endpoint`, or `zillizcloud_endpoint_whitelist`.
- Do not redesign whitelist lifecycle behavior. The API still only supports POST, so read and delete remain no-ops.
- Do not add import support.
- Do not perform a broad internal Go rename unless needed for the external behavior.

## Chosen Approach

Use a breaking external rename with a minimal internal code change.

Terraform users must move to:

| Old name | New name |
|---|---|
| `zillizcloud_endpoint_services` | `zillizcloud_private_endpoint_services` |
| `zillizcloud_endpoints` | `zillizcloud_private_endpoints` |
| `zillizcloud_endpoint` | `zillizcloud_private_endpoint` |
| `zillizcloud_endpoint_whitelist` | `zillizcloud_private_endpoint_whitelist` |

The old Terraform type names are removed from the provider surface. This is acceptable because the feature is recent and the requested behavior is a clean private-prefixed Terraform surface.

## Client API

Keep the existing Go client types and method names in `client/endpoint.go`:

- `EndpointService`
- `Endpoint`
- `CreateEndpointRequest`
- `CreateEndpointResponse`
- `AddEndpointWhitelistRequest`
- `ListEndpointServices`
- `ListEndpoints`
- `CreateEndpoint`
- `DeleteEndpoint`
- `AddEndpointWhitelist`

Only their request paths change:

| Method | New client path |
|---|---|
| `ListEndpointServices` | `privateEndpointServices?regionId=...&currentPage=...&pageSize=...` |
| `ListEndpoints` | `projects/{projectId}/privateEndpoints?currentPage=...&pageSize=...` |
| `CreateEndpoint` | `projects/{projectId}/privateEndpoints` |
| `DeleteEndpoint` | `projects/{projectId}/privateEndpoints/{endpointId}?regionId=...` plus optional `gcpProjectId` |
| `AddEndpointWhitelist` | `projects/{projectId}/privateEndpointWhitelist` |

Pagination, query construction, response decoding, and error handling continue to use the current `c.do` pattern.

## Provider Surface

The provider constructors may keep their current Go names, but their `Metadata` methods must return the new Terraform type names:

| Component | File | Metadata type suffix |
|---|---|---|
| Endpoint services data source | `internal/provider/endpoint_services_data_source.go` | `_private_endpoint_services` |
| Endpoints data source | `internal/provider/endpoints_data_source.go` | `_private_endpoints` |
| Endpoint resource | `internal/provider/endpoint_resource.go` | `_private_endpoint` |
| Endpoint whitelist resource | `internal/provider/endpoint_whitelist_resource.go` | `_private_endpoint_whitelist` |

`internal/provider/provider.go` should continue registering the endpoint constructors, but the exposed Terraform schema must only contain the new private-prefixed names because the metadata suffixes changed.

## Lifecycle Behavior

`zillizcloud_private_endpoint` keeps the existing lifecycle:

- Create calls `CreateEndpoint`, sets `id` to the endpoint id, then refreshes computed fields by listing endpoints.
- Read lists endpoints and removes state if the endpoint is absent.
- Update remains unreachable because user-supplied fields require replacement.
- Delete calls `DeleteEndpoint` with `regionId` and optional `gcpProjectId`.

`zillizcloud_private_endpoint_whitelist` keeps the existing minimal lifecycle:

- Create calls `AddEndpointWhitelist`.
- Read is a no-op because there is no GET whitelist API.
- Delete is a no-op because there is no DELETE whitelist API.
- `id` remains `project_id`, so at most one whitelist resource should be declared per project.

## Documentation And Examples

Update current docs, examples, templates, and guide snippets that reference the Terraform names:

- `zillizcloud_endpoint_services` becomes `zillizcloud_private_endpoint_services`
- `zillizcloud_endpoints` becomes `zillizcloud_private_endpoints`
- `zillizcloud_endpoint` becomes `zillizcloud_private_endpoint`
- `zillizcloud_endpoint_whitelist` becomes `zillizcloud_private_endpoint_whitelist`

Generated provider docs should move from the old endpoint filenames to private-prefixed filenames when generation runs. Historical design and plan files under `docs/superpowers` may keep old names because they document previous decisions, not current user-facing documentation.

Update `new-client-api.md` if it remains in the repo, because it currently describes the old API paths.

## Testing

Update `client/endpoint_test.go` so each test asserts the exact new request path:

- `/v2/privateEndpointServices`
- `/v2/projects/proj-1/privateEndpoints`
- `/v2/projects/proj-1/privateEndpoints/vpce-abc`
- `/v2/projects/proj-1/privateEndpointWhitelist`

Avoid substring checks that could pass against the old endpoint names.

Run the available focused tests:

- `go test ./client`
- `go test ./internal/provider/...`

Run the repo's documentation generation command if available, then inspect generated docs and examples for current names.

## Verification

Before completing implementation:

- `rg` confirms old Terraform names are gone from active docs, examples, templates, and provider metadata.
- `rg` confirms old API paths are gone from `client/endpoint.go`, client tests, and `new-client-api.md`.
- Historical specs and plans under `docs/superpowers` are excluded from that cleanup unless explicitly updated for context.

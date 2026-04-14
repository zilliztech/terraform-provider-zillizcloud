# Terraform Provider: Tiered NodeGroup Support

## Goal

Support `vmNodeGroups` array in Create and Describe APIs, enabling tiered (and future custom) node groups while maintaining backward compatibility with the old flat-field API (`searchVm`, `coreVm`, etc.).

## Current State

### Data Source (`zillizcloud_byoc_i_project_settings`)
- `node_quotas` schema: hardcoded `SingleNestedAttribute` with keys `core`, `index`, `search`, `fundamental`
- `buildNodeQuotas()` filters API response `[]NodeQuota` by name → returns null object if not found
- **Problem**: no `tiered` key in schema; API may now return tiered in `NodeQuotas`

### Resource (`zillizcloud_byoc_i_project`)
- `CreateByocOpProjectRequest` has flat fields: `SearchVM`, `CoreVM`, `FundamentalVM`
- **Problem**: no `vmNodeGroups` field; no way to pass tiered config

### API Client
- `NodeQuota` struct already supports arbitrary `Name` field
- `GetByocOpProjectSettingsResponse.NodeQuotas` is `[]NodeQuota` (array, not hardcoded)
- **Problem**: client types don't have `VmNodeGroups` on create request

## Design

### Principle: Forward Compatible, Backward Safe

1. **Data source output**: Change `node_quotas` from `SingleNestedAttribute` (fixed keys) to `MapNestedAttribute` (dynamic keys). This way any node group name the API returns (core, search, tiered, future ones) is automatically exposed.
2. **Create request**: Add `VmNodeGroups` field. When populated, the API uses the new array format. Old flat fields are still sent for backward compat with older control-api versions.
3. **Describe response**: Add `VmNodeGroups` to `DescribeByocOpProjectResponse` for completeness, but the data source already reads from `NodeQuotas` in settings API which is already an array.

### Changes

#### A. Client: `byoc_op_project.go`

```go
// Add to CreateByocOpProjectRequest:
VmNodeGroups []VmNodeGroup `json:"vmNodeGroups,omitempty"`

// New type:
type VmNodeGroup struct {
    Name string `json:"name"`
    Type string `json:"type"`  // instance type e.g. "m6i.2xlarge"
    Min  int    `json:"min"`
    Max  int    `json:"max"`
}
```

#### B. Data Source: `projects_settings_data.go`

Change `node_quotas` schema from `SingleNestedAttribute` to `MapNestedAttribute`:

```go
// Before:
"node_quotas": schema.SingleNestedAttribute{
    Attributes: map[string]schema.Attribute{
        "core":        nodeSchema,
        "index":       nodeSchema,
        "search":      nodeSchema,
        "fundamental": nodeSchema,
    },
}

// After:
"node_quotas": schema.MapNestedAttribute{
    Computed: true,
    NestedObject: schema.NestedAttributeObject{
        Attributes: nodeSchemaAttributes,  // disk_size, min_size, max_size, etc.
    },
}
```

Update `Describe()` in `projects_settings_data.go` and `projects_settings_store.go`:
- Instead of calling `buildNodeQuotas()` 4 times for fixed names, iterate all `response.NodeQuotas` and build a map.

#### C. Data Source Model: `projects_settings_model.go`

```go
// Before:
NodeQuotas types.Object `tfsdk:"node_quotas"`

// After:
NodeQuotas types.Map `tfsdk:"node_quotas"`
```

#### D. Resource Store: `projects_resource_store.go`

In `Create()`, build `VmNodeGroups` from the instances config and include in request:

```go
// Build vmNodeGroups from the instances config
var vmNodeGroups []zilliz.VmNodeGroup
// ... populate from resource model ...
request.VmNodeGroups = vmNodeGroups
```

#### E. Terraform Examples: `data.tf`

The `k8s_node_groups` local already does `for name, ng in node_quotas`. With `MapNestedAttribute`, it becomes a native map — the existing for-each just works.

### Backward Compatibility

| Scenario | Behavior |
|----------|----------|
| Old API (no tiered in NodeQuotas) | `node_quotas` map won't have "tiered" key → `enable_tiered = false` → no nodegroup created |
| New API (tiered in NodeQuotas, max>0) | `node_quotas` map has "tiered" key → `enable_tiered = true` → nodegroup created |
| Old Terraform config (no vmNodeGroups) | `VmNodeGroups` is nil/empty → API falls back to flat fields |
| New Terraform config (has vmNodeGroups) | `VmNodeGroups` sent → API uses array format |
| Existing state with SingleNested node_quotas | **BREAKING**: `SingleNestedAttribute` → `MapNestedAttribute` changes state schema. Users need `terraform state replace-provider` or state migration. Mitigation: add state upgrade in data source. |

### State Migration Note

Changing `node_quotas` from `SingleNestedAttribute` to `MapNestedAttribute` is a schema change. For data sources, this is less impactful than resources (data sources don't have persistent state that needs migration — they're re-read every plan). But the **terraform-zilliz-examples** code that references `data.zillizcloud_byoc_i_project_settings.this.node_quotas.core` would need to change to `data.zillizcloud_byoc_i_project_settings.this.node_quotas["core"]`.

### Validation

The `k8s_node_groups` variable validation in the EKS module requires `max_size > 0`. With optional groups (search/tiered having max=0 defaults), this validation will fail. Fix: relax to `max_size >= 0` (already handled in terraform-zilliz-examples PR).

## File Changes

| File | Change |
|------|--------|
| `client/byoc_op_project.go` | Add `VmNodeGroups` field + `VmNodeGroup` type |
| `internal/provider/byoc_i/projects_settings_data.go` | `node_quotas` schema → `MapNestedAttribute`; Describe() builds dynamic map |
| `internal/provider/byoc_i/projects_settings_store.go` | Describe() builds dynamic map from NodeQuotas array |
| `internal/provider/byoc_i/projects_settings_model.go` | `NodeQuotas` field type: `types.Object` → `types.Map` |
| `internal/provider/byoc_i/projects_resource_store.go` | Create() populates `VmNodeGroups` in request |

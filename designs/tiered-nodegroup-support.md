# Terraform Provider: Tiered NodeGroup Support

## Goal

Expose the `tiered` node group quota returned by the BYOC-I settings API through the `zillizcloud_byoc_i_project_settings` data source and resource, so that terraform-zilliz-examples can provision a tiered storage node group when the backend has tiered enabled for the project.

## Problem

- `node_quotas` schema was a hardcoded `SingleNestedAttribute` with keys `core`, `index`, `search`, `fundamental` — no `tiered` key, so even when the API returned a tiered entry, users could not read it from Terraform.
- `buildNodeQuotas()` already filters the API response `[]NodeQuota` by name and returns a null object if the name is missing, so it can handle the "tiered not enabled" case without any new logic.

## Design

### Principle: Minimal Schema Change, Backward Safe

Add `tiered` as a new nested attribute inside the existing `node_quotas` `SingleNestedAttribute`. The `buildNodeQuotas()` helper already returns `types.ObjectNull` when the API response does not include the given node group name, so `node_quotas.tiered` is automatically null when tiered storage is not enabled — no extra null-handling needed.

An earlier draft proposed switching `node_quotas` to `MapNestedAttribute` (dynamic keys). That was rejected because it changes the state schema for all existing users and forces examples to switch from attribute access (`.core`) to map index access (`["core"]`). The fixed-key approach keeps all existing references working.

A second draft introduced `tiered_node_quota` as a separate top-level field. That was also rejected: it splits one logical concept (node group quotas from the API) across two schema fields and complicates the examples' merge logic. Keeping everything inside `node_quotas` is cleaner.

### Changes

#### A. Schema: `projects_settings_data.go` and `projects_settings_resource.go`

Add `tiered` as a fifth key inside the existing `node_quotas` `SingleNestedAttribute`:

```go
"node_quotas": schema.SingleNestedAttribute{
    Computed: true,
    Attributes: map[string]schema.Attribute{
        "core":        nodeSchema,
        "index":       nodeSchema,
        "search":      nodeSchema,
        "fundamental": nodeSchema,
        "tiered": func() schema.SingleNestedAttribute {
            s := nodeSchema
            s.MarkdownDescription = "Tiered storage node group quota. Null when tiered storage is not enabled."
            return s
        }(),
    },
},
```

#### B. Store: `projects_settings_store.go` and `projects_settings_data.go`

`Describe()` calls `buildNodeQuotas("tiered", response.NodeQuotas)` alongside the existing four and includes it in the `types.ObjectValue` call:

```go
tiered, err := buildNodeQuotas("tiered", response.NodeQuotas)
if err != nil {
    return data, err
}

NodeQuotas, diag := types.ObjectValue(nodeQuotasGenerateAttrTypes, map[string]attr.Value{
    "core":        core,
    "index":       index,
    "search":      search,
    "fundamental": fundamental,
    "tiered":      tiered,
})
```

`nodeQuotasGenerateAttrTypes` gains the corresponding `"tiered"` entry.

`buildNodeQuotas()` is unchanged — it already returns `types.ObjectNull(...)` when the named node group is missing from the API response, which gives `node_quotas.tiered = null` for old API responses.

#### C. Model: `projects_resource_model.go`

Add `Tiered NodeQuota \`tfsdk:"tiered"\`` to the `NodeQuotas` struct. No new top-level field on `BYOCOpProjectSettingsResourceModel` or `BYOCOpProjectSettingsDataModel`.

### Backward Compatibility

| Scenario | Behavior |
|----------|----------|
| Old API (no tiered in NodeQuotas array) | `node_quotas.tiered = null` (via `buildNodeQuotas` returning `ObjectNull`) |
| New API (tiered in NodeQuotas, max>0) | `node_quotas.tiered` populated with quota values |
| Existing examples referencing `node_quotas.core` | Still works — fixed-key schema preserved |
| Examples iterating `for name, ng in node_quotas` | Still works — iterates 5 keys instead of 4, with tiered null-handling in HCL |

### Consumer Side (terraform-zilliz-examples)

Examples reference `data.zillizcloud_byoc_i_project_settings.this.node_quotas` as before. To detect whether tiered is enabled:

```hcl
enable_tiered = (
  local.k8s_node_groups["tiered"] != null
  && local.k8s_node_groups["tiered"].max_size > 0
)
```

The `max_size > 0` guard handles both the null case (tiered not returned by API) and the disabled case (returned but max=0).

## File Changes

| File | Change |
|------|--------|
| `internal/provider/byoc_i/projects_settings_data.go` | Add `tiered` key to `node_quotas` schema; call `buildNodeQuotas("tiered", ...)` and include in `ObjectValue` |
| `internal/provider/byoc_i/projects_settings_resource.go` | Add `tiered` key to `node_quotas` schema |
| `internal/provider/byoc_i/projects_settings_store.go` | Call `buildNodeQuotas("tiered", ...)` and include in `ObjectValue`; add `"tiered"` to `nodeQuotasGenerateAttrTypes` |
| `internal/provider/byoc_i/projects_resource_model.go` | Add `Tiered NodeQuota` field to `NodeQuotas` struct |
| `docs/data-sources/byoc_i_project_settings.md` | Regenerated via `go generate` |
| `docs/resources/byoc_i_project_settings.md` | Regenerated via `go generate` |

# Terraform Provider: Tiered NodeGroup Support

## Goal

Expose the `tiered` node group quota returned by the BYOC-I settings API through the `zillizcloud_byoc_i_project_settings` data source and resource, so that terraform-zilliz-examples can provision a tiered storage node group when the backend has tiered enabled for the project.

## Problem

- `node_quotas` schema is a `SingleNestedAttribute` with fixed keys `core`, `index`, `search`, `fundamental`. There is no `tiered` key, so even when the API returns a tiered entry in its `NodeQuotas` array, users cannot read it from Terraform.
- `buildNodeQuotas()` already filters the API response `[]NodeQuota` by name and returns a null object if the name is missing, so the "tiered not enabled" case is easy to represent — the question is only where to hang the new attribute.

## Design

### Principle: Minimal Schema Change, Forward Compatible

Add `tiered_node_quota` as a **separate top-level computed attribute** next to `node_quotas`. Do **not** touch the `node_quotas` schema.

Rejected alternatives and why:

1. **Add `tiered` as a fifth key inside `node_quotas` (`SingleNestedAttribute`)**
   With a fixed-key nested object, the `tiered` attribute would *always* be present in state, set to null when the API does not return it. HCL callers that iterate `for name, ng in node_quotas` would suddenly see a 5th iteration with a null value and have to filter it out. Every existing example — including the master branch that users have already deployed — would start surfacing `node_quotas.tiered = null` and need changes to iterate safely. That breaks the "forward compatible" goal.

2. **Switch `node_quotas` to `MapNestedAttribute` (dynamic keys)**
   Clean in theory: the API returns whatever node groups it wants, and Terraform exposes them as a map. But this changes the state schema type (`SingleNestedAttribute` → `MapNestedAttribute`), which is a breaking state-schema change. Existing HCL references like `node_quotas.core` would need to become `node_quotas["core"]`, and every deployed state would need migration. Hard "no" for a point release.

3. **Add `tiered` as a separate top-level field** (chosen)
   `node_quotas` stays byte-identical in schema and state. Users who don't know about tiered see no change at all — their HCL keeps working, `terraform plan` shows no drift, and state format is unchanged. Users who want tiered opt in by reading a new top-level field `tiered_node_quota`.

### Changes

#### A. Schema: `projects_settings_data.go` and `projects_settings_resource.go`

Add a new top-level attribute `tiered_node_quota` alongside `node_quotas`. Reuse the same `nodeSchema` (disk_size / min_size / max_size / desired_size / instance_types / capacity_type) with a custom markdown description:

```go
"tiered_node_quota": func() schema.SingleNestedAttribute {
    s := nodeSchema
    s.MarkdownDescription = "Tiered storage node group quota. Null when tiered storage is not enabled."
    return s
}(),
```

`node_quotas` is **unchanged**.

#### B. Store: `projects_settings_store.go` and `projects_settings_data.go`

In `Describe()`, after building `NodeQuotas` for the four existing keys, call `buildNodeQuotas("tiered", response.NodeQuotas)` once more and assign the result to the new top-level model field:

```go
tiered, err := buildNodeQuotas("tiered", response.NodeQuotas)
if err != nil {
    return data, err
}
data.TieredNodeQuota = tiered
```

`buildNodeQuotas()` is unchanged — it already returns `types.ObjectNull(...)` when the named node group is missing from the API response, which gives `tiered_node_quota = null` for old API responses.

#### C. Model: `projects_resource_model.go`

Add `TieredNodeQuota types.Object` to both `BYOCOpProjectSettingsResourceModel` and `BYOCOpProjectSettingsDataModel`. The `refresh()` helper on the data model copies it through. `NodeQuotas` struct is unchanged (still 4 keys).

#### D. Resource Create/Read: `projects_settings_resource.go`

After the post-create / post-read `Describe()`, sync the new field alongside the existing ones:

```go
data.OpConfig = model.OpConfig
data.NodeQuotas = model.NodeQuotas
data.TieredNodeQuota = model.TieredNodeQuota
```

### Backward Compatibility

| Scenario | Behavior |
|----------|----------|
| Old API (no tiered in NodeQuotas array) | `tiered_node_quota = null` (via `buildNodeQuotas` returning `ObjectNull`); no change to `node_quotas` |
| New API (tiered in NodeQuotas, max>0) | `tiered_node_quota` populated with quota values |
| Existing examples referencing `node_quotas.core` etc. | Still works — `node_quotas` schema preserved |
| Existing examples iterating `for name, ng in node_quotas` | Still works — iterates the same 4 keys, no new null entry |
| Existing state from older provider | No state migration required — `node_quotas` unchanged; `tiered_node_quota` is a new Computed attribute populated on next refresh |

### Consumer Side (terraform-zilliz-examples)

Examples that want tiered merge the new top-level field into the node-group map conditionally:

```hcl
_tiered_from_api = data.zillizcloud_byoc_i_project_settings.this.tiered_node_quota != null ? {
  tiered = data.zillizcloud_byoc_i_project_settings.this.tiered_node_quota
} : {}

k8s_node_groups = {
  for name, ng in merge(
    local._optional_ng_defaults,
    data.zillizcloud_byoc_i_project_settings.this.node_quotas,
    local._tiered_from_api,
  ) : name => merge(ng, {
    disk_size = ng.disk_size > 0 ? ng.disk_size : 100  # tiered (i4i) returns 0 — floor to EBS root size
  })
}

enable_tiered = (
  data.zillizcloud_byoc_i_project_settings.this.tiered_node_quota != null
  && local.k8s_node_groups["tiered"].max_size > 0
)
```

## File Changes

| File | Change |
|------|--------|
| `internal/provider/byoc_i/projects_settings_data.go` | Add `tiered_node_quota` top-level schema attribute; `Describe()` populates `data.TieredNodeQuota` |
| `internal/provider/byoc_i/projects_settings_resource.go` | Add `tiered_node_quota` top-level schema attribute; sync field in Create/Read |
| `internal/provider/byoc_i/projects_settings_store.go` | `Describe()` calls `buildNodeQuotas("tiered", ...)` and assigns to `data.TieredNodeQuota` |
| `internal/provider/byoc_i/projects_resource_model.go` | Add `TieredNodeQuota types.Object` field to both resource and data models |
| `docs/data-sources/byoc_i_project_settings.md` | Regenerated via `go generate` |
| `docs/resources/byoc_i_project_settings.md` | Regenerated via `go generate` |

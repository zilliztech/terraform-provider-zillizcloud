# Terraform Provider: API Key Resource

## Goal

Complete the `zillizcloud_api_key` resource and add a `zillizcloud_api_keys` data source, so that Customized API keys are fully manageable from Terraform. This is a prerequisite for the "admin key bootstraps project key" pattern in BYOC Terraform workflows.

**Design stance**: the Terraform layer must (a) match the live public API exactly, and (b) absorb API evolution the API itself doesn't plan for — unknown future roles, new scoping dimensions, response field additions, and error-code drift must not require a provider release or break existing state.

## Background

Zilliz Cloud has two types of API keys:

| Type | Bound To | Lifecycle | Risk |
|------|----------|-----------|------|
| **Personal** | Individual user account | Auto-created on registration, **auto-deleted when user leaves org** | Person leaves → Terraform breaks |
| **Customized** | Organization | Manually created, independent of any user account | Person-independent, survives team changes |

For production BYOC deployments, Customized API keys scoped to specific projects are the recommended practice.

### Motivating Use Case

```
1. Org Owner creates a bootstrap admin API key (Console UI, one-time)
2. Terraform Phase 1 (admin key): create project + create project-scoped API key
3. Terraform Phase 2 (project key): provision clusters, databases, collections
```

> **Terraform limitation**: Provider configuration is static — resolved at plan time, cannot reference resource outputs. Therefore Phase 1 and Phase 2 must be separate `terraform apply` runs (or use a CI wrapper that extracts the key between runs). This is a fundamental Terraform constraint, not a Zilliz-specific issue.

> **Key rotation caveat**: a project-scoped key **cannot manage API keys at all** (error `96041`, see Permission Model). Rotating the Phase 2 key always requires going back to Phase 1 with the admin key. CI pipelines must keep the admin key available for rotation, not just bootstrap.

## Live API Contract (Verified 2026-07-22)

All endpoints tested end-to-end against `api.cloud.zilliz.com` with an Org Owner key.

### Endpoint Inventory

| Method | Path | Function | Verified |
|--------|------|----------|:--------:|
| `POST` | `/v2/apiKeys` | Create customized key | ✅ |
| `GET` | `/v2/apiKeys` | List keys (paginated) | ✅ |
| `GET` | `/v2/apiKeys/{apiKeyId}` | Describe single key | ✅ |
| `PUT` | `/v2/apiKeys/{apiKeyId}` | Update key | ✅ |
| `DELETE` | `/v2/apiKeys/{apiKeyId}` | Delete key | ✅ |

> **⚠️ Path is camelCase `/v2/apiKeys`**, NOT kebab-case. The internal source (`OpenApiKeyController.java`) maps to `/v2/api-keys`, but the **public API gateway routes to `/v2/apiKeys`**. Tested: kebab-case returns `"no Route matched"`. The merged client path is correct.

### Permission Model

| Caller Key Role | Can List/Describe | Can Create/Update/Delete |
|-----------------|:-:|:-:|
| Org Owner key | ✅ | ✅ (except cannot create another Owner key) |
| Member + Project Admin key | ❌ `code: 96041` "Only API keys with Org Owner permission can access this endpoint" | ❌ Same |

**Key finding**: API key management requires an **Org Owner** scoped key. A project-scoped key cannot manage API keys at all — not even list them.

### Create Request (Verified)

```json
{
  "name": "terraform-test-key",
  "orgRole": "Member",
  "projects": [{
    "projectId": "proj-<project-id>",
    "role": "Read-Only",
    "allCluster": true,
    "allVolume": true
  }]
}
```

**Schema uses human-readable role names**, not role IDs:
- `orgRole`: `"Owner"` | `"Member"` | `"Billing-Admin"`
- `projects[].role`: `"Admin"` | `"Read-Write"` | `"Read-Only"`

**Cluster/Volume scoping** is supported:
- `allCluster: true` — all clusters in the project
- `allCluster: false` + `clusterIds: ["in01-xxx"]` — specific clusters only
- Same pattern for `allVolume` / `volumeIds`

**Creating an Owner key is blocked** (error code `46001`):
```json
{
  "code": 46001,
  "message": "Creating API key with Org Owner permission via API is not allowed. Please use the Zilliz Cloud Console for this operation."
}
```

### Create Response (Verified)

```json
{
  "code": 0,
  "data": {
    "apiKeyId": "key-<created-key-id>",
    "apiKey": "<64-hex-secret — returned only once>"
  }
}
```

> `apiKey` value is **only returned at creation time**. Subsequent GET calls do not include it.

### Describe / List Response (Verified)

```json
{
  "code": 0,
  "data": {
    "apiKeyId": "key-<created-key-id>",
    "name": "terraform-test-key",
    "description": "",
    "creatorName": "<org-user-name>",
    "createdBy": "key-<creator-key-id>",
    "orgRole": "Member",
    "projects": [
      {
        "projectId": "proj-<project-id>",
        "projectName": "<project-name>",
        "role": "Read-Only",
        "allCluster": true,
        "clusters": [],
        "allVolume": true,
        "volumes": []
      }
    ],
    "createTime": "2026-07-22T18:22:47Z"
  }
}
```

**Notable fields**:
- `createdBy`: when created by another API key (not a user), this shows the **creator key's ID** (e.g., `key-<creator-key-id>`), not an email
- `orgRole`: single string, not a list
- `projects`: flat array with `role` as string name, plus cluster/volume scoping
- `description`: returned as `""` when unset (never omitted) — matters for Terraform null-vs-empty handling, see Schema

### Update Request/Response (Verified)

Same schema as Create. Successfully tested changing `name`, `description`, and upgrading `role` from `Read-Only` to `Read-Write` in-place.

### Delete Response (Verified)

Returns the deleted key's full details (same schema as Describe). Subsequent GET returns `code: 404`.

### List Response (Verified)

```json
{
  "code": 0,
  "data": {
    "count": 56,
    "currentPage": 1,
    "pageSize": 10,
    "apiKeys": [...]
  }
}
```

Paginated, **default `pageSize=10`**. Returns all Customized API keys in the org (both Owner and Member keys).

## Source Code vs Live API Discrepancies

The internal source code (`cloud-control-api`) uses a **different schema** than the public API gateway exposes:

| Aspect | Source Code (`OpenApiKeyController.java`) | Live Public API |
|--------|----------------------------------------|-----------------|
| **Path** | `/v2/api-keys` (kebab-case) | `/v2/apiKeys` (camelCase) |
| **Org role** | `orgRoleIds: ["role-xxx"]` (list of IDs) | `orgRole: "Member"` (single name string) |
| **Project roles** | `projectRoles: [{projectId, roleIds}]` (ID-based) | `projects: [{projectId, role, allCluster, ...}]` (name-based) |
| **Response org role** | `orgRoles: [{bindingId, roleId, roleName, roleType}]` (list) | `orgRole: "Member"` (single string) |
| **Response projects** | `projectRoles: [{projectId, roles: [...]}]` (nested) | `projects: [{projectId, projectName, role, allCluster, clusters, allVolume, volumes}]` (flat) |

**Conclusion**: The public gateway has a **translation layer** (`cloud-service/ApiKeyController.java`) between the public API and the internal `cloud-control-api`. The Terraform provider client targets the **public API schema**, not the internal schema.

## Current State: Already Merged in `master` (PR #201)

> ⚠️ Earlier drafts of this doc treated `origin/feat/api-key` as unmerged work. That is stale.
> **PR #201** (commits `ba14766` + `3eba148`, 2026-04-03) merged the resource into `master`.

What `master` already has:

| Item | Location | Status |
|------|----------|:-:|
| HTTP client, 5 methods, camelCase `apiKeys` path | `client/apikey.go` | ✅ |
| Resource `zillizcloud_api_key` (CRUD + Import) | `internal/provider/apikey_resource.go` (~428 lines), registered in `provider.go` | ✅ |
| `creatorEmail` → `createdBy` rename | commit `3eba148` | ✅ |
| Read 404 → `RemoveResource` (external-delete drift) | `Read()` | ✅ |
| Delete tolerates 404 (key already removed externally) | `Delete()` | ✅ |
| `ValidateConfig`: Member requires `project_access`; `all_cluster=true` conflicts with `cluster_ids` | `ValidateConfig()` | ✅ |
| `ModifyPlan`: auto-set `all_cluster=false` when `cluster_ids` provided | `ModifyPlan()` | ✅ |
| Import (`key_value` unavailable) | `ImportState()` | ✅ |
| Acceptance tests (full CRUD lifecycle) | `apikey_resource_test.go` | ✅ |

**Actual remaining gaps** — the real scope of this design:

| # | Gap | Detail |
|---|-----|--------|
| G1 | `description` missing end-to-end | Not in client structs, not in schema, not populated in Read/Update |
| G2 | `all_volume` / `volume_ids` absent from schema | **Deliberately removed** in `ba14766` because the API response didn't return them at the time. Live test (2026-07-22) proves the API now returns `allVolume` / `volumes` — the API evolved. Re-adding is a *restore* with state-compat implications (see Schema) |
| G3 | No `zillizcloud_api_keys` data source | — |
| G4 | `ListApiKeys` does **not paginate** | Current impl is a single `GET apiKeys` with no paging params. With default `pageSize=10` and 56 keys in the test org, a data source built on it silently returns 10 keys |
| G5 | Forward-compat hardening | Role validators lock today's enum; not-found code drift (`80001` vs `404`); update send-semantics — see next section |
| G6 | Multi-project ordering | `populateProjectAccess` fills state in API response order; API order is not guaranteed to match config order → spurious diffs with >1 project |

## Forward Compatibility Principles

The API team designs for today's Console; Terraform state lives for years. The provider must absorb the following classes of API evolution **without a provider release and without breaking existing state**:

### FC1. Unknown role names must pass through

Do **not** hard-fail validation on role strings outside today's enum (`Member` / `Billing-Admin`, project `Admin` / `Read-Write` / `Read-Only`). If the API adds a role (e.g. `Auditor`), an old provider binary must still be able to:
- send it on Create/Update (server is the source of truth for validity), and
- read it back into state without diagnostics.

Implementation: replace `stringvalidator.OneOf(...)` with a **warning-only** custom validator ("unrecognized role, passing through to API") — keep only the hard client-side block for `orgRole = "Owner"` on create/update, because the server rejects it with `46001` and we can give a better message ("use the Console"). `"Owner"` remains accepted for **import** (Console-created Owner keys are importable/readable).

### FC2. New response fields must be inert

Go's `encoding/json` ignores unknown fields — new response fields (a future `expireTime`, `lastUsedTime`, …) won't break decoding. Rule: **never** switch the client to strict decoding (`DisallowUnknownFields`). New fields get exposed in a later provider release as `Computed` attributes; until then they are silently ignored.

### FC3. Scoping dimensions are a recurring pattern

`allCluster`/`clusterIds` and `allVolume`/`volumeIds` follow one pattern: `all<X> bool` + `<x>Ids []string` + response `<x>s [{id, name}]`. A future dimension (e.g. pipelines) will follow the same shape. Keep each dimension **independent and additive** in both the client struct and the schema — no shared "scope" abstraction that would need reshaping when a third dimension lands. G2's restore is the template for how a new dimension gets added later (including the state-upgrade recipe).

### FC4. Not-found detection must accept code drift

History: internal error code for not-found was `80001` (per `ba14766` commit message); the live gateway today returns `404`. Treat **both** as not-found in `Read` (→ `RemoveResource`) and `Delete` (→ success). Gateway/backend version skew between prod, CN, and UAT makes single-code matching fragile.

### FC5. Update always sends the full desired state

Never rely on server-side merge semantics for PUT. The provider always sends `name`, `description`, `orgRole`, and the complete `projects` array from the plan. This sidesteps two unverified behaviors (see Open Questions): whether omitting `projects` clears or preserves access, and whether `"description": ""` clears the field. Full-state PUT is deterministic regardless of how the server answers those questions. Concretely: **no `omitempty` on Update request fields** (an `omitempty` on `description` makes "clear the description" silently impossible).

### FC6. Pagination without assumptions

The list client loops `currentPage` until `len(collected) >= count` (with a hard safety cap, e.g. 100 pages), using an explicit `pageSize=100` request. If the server clamps `pageSize` to a lower max, the loop still terminates correctly because it trusts the *returned* `count`/page contents, not the requested page size. Do not assume the max page size — it is unverified.

### FC7. Friendly permission errors

All five operations fail with `96041` when the provider is configured with a non-Owner key. Wrap this error once in the client layer: *"API key management requires an Org Owner API key; the configured key is project-scoped. See docs/guides/api-key-bootstrap."* Otherwise users hit a bare gateway message at apply time.

## Design

### Principle: Match Public API Exactly, Cushion What It Doesn't Promise

The client layer mirrors the **live public API contract** (name-based roles, camelCase path, flat response). The resource layer adds the durability the API doesn't promise: state compatibility across schema versions, order-insensitivity, null/empty normalization, error-code tolerance.

### File Structure (actual repo names)

```
client/
  └── apikey.go                     # EXISTS — add Description, fix pagination, wrap 96041

internal/provider/
  ├── apikey_resource.go            # EXISTS — add description, restore volume fields,
  │                                 #   schema v0→v1 upgrader, order-stable populate
  ├── apikey_resource_test.go       # EXISTS — extend (description drift, multi-project, volume)
  └── apikeys_data_source.go        # NEW — naming follows projects_data_source.go convention

docs/resources/apikey.md            # regenerate
docs/data-sources/apikeys.md        # NEW
examples/resources/zillizcloud_api_key/resource.tf   # update
```

> Note: the resource file is `apikey_resource.go` (no underscore between "api" and "key") — earlier drafts of this doc used `api_key_resource.go`, which doesn't exist.

### Client Layer: `client/apikey.go` (delta)

```go
// --- Request types (match live public API) ---

type ApiKeyProjectAccess struct {
    ProjectId  string   `json:"projectId"`
    Role       string   `json:"role,omitempty"`       // pass-through; server validates
    AllCluster *bool    `json:"allCluster,omitempty"`
    ClusterIds []string `json:"clusterIds,omitempty"`
    AllVolume  *bool    `json:"allVolume,omitempty"`
    VolumeIds  []string `json:"volumeIds,omitempty"`
}

type CreateApiKeyRequest struct {
    Name        string                `json:"name"`
    Description string                `json:"description,omitempty"`  // omitempty OK on create
    OrgRole     string                `json:"orgRole"`
    Projects    []ApiKeyProjectAccess `json:"projects,omitempty"`
}

// FC5: Update always sends full desired state — NO omitempty on scalar fields,
// so clearing description is expressible and server-side merge semantics are moot.
type UpdateApiKeyRequest struct {
    Name        string                `json:"name"`
    Description string                `json:"description"`
    OrgRole     string                `json:"orgRole"`
    Projects    []ApiKeyProjectAccess `json:"projects"`
}

// --- Response types: add Description to ApiKeyResponse ---

type ApiKeyResponse struct {
    ApiKeyId    string                  `json:"apiKeyId"`
    Name        string                  `json:"name"`
    Description string                  `json:"description"`   // NEW — "" when unset
    CreatorName string                  `json:"creatorName"`
    CreatedBy   string                  `json:"createdBy"`     // email or creator key ID
    OrgRole     string                  `json:"orgRole"`
    Projects    []ApiKeyProjectResponse `json:"projects"`
    CreateTime  string                  `json:"createTime"`
}

// --- ListApiKeys: FC6 pagination loop (current impl fetches page 1 only) ---

func (c *Client) ListApiKeys() ([]ApiKeyResponse, error) {
    var all []ApiKeyResponse
    for page := 1; page <= maxListPages; page++ {
        var resp zillizResponse[ApiKeyListResponse]
        err := c.do("GET", fmt.Sprintf("apiKeys?currentPage=%d&pageSize=100", page), nil, &resp)
        if err != nil {
            return nil, err
        }
        all = append(all, resp.Data.ApiKeys...)
        if len(resp.Data.ApiKeys) == 0 || len(all) >= resp.Data.Count {
            break
        }
    }
    return all, nil
}
```

Not-found handling (FC4) stays in the resource layer but matches on `code == 404 || code == 80001`.

### Resource Schema: `zillizcloud_api_key`

```hcl
resource "zillizcloud_api_key" "project_key" {
  name        = "byoc-terraform-key"
  description = "Scoped key for BYOC project automation"
  role        = "Member"

  project_access = [{
    project_id  = zillizcloud_project.prod.id
    role        = "Admin"
    all_cluster = true
    all_volume  = true
  }]
}

output "api_key_value" {
  value     = zillizcloud_api_key.project_key.key_value
  sensitive = true
}
```

#### Schema Attributes (delta on existing schema)

| Attribute | Type | Required | Plan Modifier / Default | Notes |
|-----------|------|----------|---------------|-------|
| `id` | String | Computed | `UseStateForUnknown` | = apiKeyId *(exists)* |
| `name` | String | Required | — | *(exists)* |
| `description` | String | Optional + Computed | `Default: ""` | **NEW.** Defaulting to `""` (not null) matches the API, which always returns `description: ""` when unset — avoids null-vs-empty perpetual diffs on read/import |
| `role` | String | Required | — | *(exists)*; validator loosened per FC1 — hard-block only `"Owner"` on create/update, warn-only otherwise |
| `project_access` | ListNested | Optional | — | *(exists)*; populate becomes order-stable, see below |
| `key_value` | String | Computed, Sensitive | `UseStateForUnknown` | *(exists)* Only set on Create, never refreshed |
| `creator_name` / `created_by` / `create_time` | String | Computed | `UseStateForUnknown` | *(exists)* |

**Nested `project_access` (delta):**

| Attribute | Type | Required | Default | Notes |
|-----------|------|----------|---------|-------|
| `project_id` | String | Required | — | *(exists)* |
| `role` | String | Optional | `"Read-Only"` | *(exists)*; validator warn-only per FC1 |
| `all_cluster` | Bool | Optional + Computed | `true` | *(exists)* |
| `cluster_ids` | List(String) | Optional | — | *(exists)* |
| `all_volume` | Bool | Optional + Computed | `true` | **RESTORED** (G2) |
| `volume_ids` | List(String) | Optional | — | **RESTORED** (G2); mirror the cluster validations: conflicts with `all_volume = true` (ValidateConfig) and auto-corrects `all_volume = false` (ModifyPlan) |

#### State Compatibility for Restored Volume Fields (G2)

Existing deployments have state written by the current schema, which has **no** `all_volume` / `volume_ids`. Adding attributes makes the framework read them as null → with `Default(true)` the first plan after upgrade would show `all_volume: null → true` noise (or worse, an unwanted update call).

Mitigation: bump resource **schema `Version` 0 → 1** with a `StateUpgrader` that fills `all_volume = true`, `volume_ids = null` for every `project_access` entry (that is exactly what the current code sends hardcoded — `allVolume: true` — so the upgraded state matches reality). Result: upgrade is a silent no-op plan.

> This 0→1 upgrader is also the template for any future scoping dimension (FC3).

#### Order-Stable `populateProjectAccess` (G6)

The API does not guarantee `projects[]` order matches the request. Current code copies API order into state → with ≥2 projects a refresh can reorder the list and produce a spurious diff.

Fix without breaking state type (stay `ListNestedAttribute` — switching to Set is a state-incompatible type change, not worth it):

1. Build a `projectId → response item` map from the API response
2. Iterate the **existing state/plan order**, emitting matched items in that order
3. Append any response items not present in prior state (out-of-band additions) at the end, sorted by `projectId` for determinism

Same normalization applies to `cluster_ids` / `volume_ids` inner lists: preserve config order for known IDs, append unknown ones sorted.

#### CRUD Behavior

| Operation | Behavior |
|-----------|----------|
| **Create** | `POST /v2/apiKeys` → store `apiKeyId` as `id`, `apiKey` as `key_value` |
| **Read** | `GET /v2/apiKeys/{id}` → refresh all fields except `key_value`; not-found (`404` **or** `80001`) → `RemoveResource` *(404 path exists; add 80001)* |
| **Update** | `PUT /v2/apiKeys/{id}` full desired state (FC5). Key value unchanged by update — verified |
| **Delete** | `DELETE /v2/apiKeys/{id}`; not-found tolerated *(exists; add 80001)* |
| **Import** | By `apiKeyId`; `key_value` empty and unrecoverable → warning diagnostic *(exists)* |

#### `key_value` Handling

The API key secret is only returned once at creation. *(All implemented in master; restated for completeness.)*

1. On Create: store in state as `Sensitive` attribute
2. On Read: skip this field (preserve state value)
3. On Import: field will be empty — warning diagnostic
4. On Plan: `UseStateForUnknown` prevents spurious diff

This matches the `aws_iam_access_key.secret` pattern. **Caveat**: `Sensitive` only redacts CLI output — the plaintext key lives in the state file. Document that users should use an encrypted state backend; when the ecosystem settles, evaluate Terraform write-only/ephemeral attributes as a follow-up (currently not applicable to computed outputs).

### Data Source: `zillizcloud_api_keys`

```hcl
data "zillizcloud_api_keys" "all" {}

output "key_names" {
  value = [for k in data.zillizcloud_api_keys.all.api_keys : k.name]
}
```

- Returns **all** Customized API keys in the org (depends on G4 pagination fix — without it, silently truncates at 10)
- Attributes mirror `ApiKeyResponse` (including `description`) minus `key_value`
- Plain computed list in API-returned order (no order normalization needed for a data source)
- Only usable with an Org Owner key — a project-scoped key gets `96041` (surface the FC7 wrapped error)

### Provider Registration

`NewApiKeyResource` is already registered. Add `NewApiKeysDataSource` to `DataSources()`.

## Resolved Questions

### Q1: Role ID vs Name → **Name-based (RESOLVED)**

Live API uses human-readable role names (`"Member"`, `"Admin"`, `"Read-Only"`), not role IDs. The gateway translates names to IDs internally. No ID resolution needed in the provider.

### Q2: Cluster/Volume Scoping → **Supported (RESOLVED)**

Live API supports `allCluster`, `clusterIds`, `allVolume`, `volumeIds`. Verified in List response: keys with `allCluster: false` show specific `clusters: [{clusterId, clusterName}]`. Note this is **newer than PR #201** — the response didn't include volume fields in April, which is why the schema dropped them (G2).

### Q3: Owner Key Creation → **Blocked by API (RESOLVED)**

`orgRole: "Owner"` returns error `46001`. Provider validates this client-side with a clear error message referencing the Console (Owner still accepted for import of Console-created keys).

### Q4: Permission Requirements → **Org Owner Only (RESOLVED)**

All API key management operations require an Org Owner key. A project-scoped key gets `96041`. Documented prominently + FC7 error wrapping.

## Open Questions (untested — do not assume)

| # | Question | Why it matters | Current mitigation |
|---|----------|----------------|--------------------|
| O1 | Does `PUT` with `projects` omitted **clear** or **preserve** project access? | Server-merge ambiguity | Moot under FC5 (always send full array) — but worth a probe test before relying on it anywhere else |
| O2 | Does `PUT` with `"description": ""` clear the field? | Clearing description from Terraform must work | FC5 sends it always; add an acceptance test asserting post-clear reads back `""` |
| O3 | `Billing-Admin` — **zero test coverage** | Create with `orgRole: "Billing-Admin"`, and in-place `Member ↔ Billing-Admin` transitions, never exercised in the 2026-07-22 transcript | Add to acceptance tests before documenting as supported |
| O4 | Max `pageSize` for List | Pagination efficiency | FC6 loop is correct regardless; probe once and record |
| O5 | Does not-found still return `80001` on any environment (CN / UAT)? | FC4 dual-code matching | Harmless to match both; verify CN gateway when convenient |

## Implementation Plan (revised — resource already exists)

| Phase | Scope | Files | Notes |
|-------|-------|-------|-------|
| **1. Client delta** | `Description` in structs; Update full-state (drop omitempty); pagination loop; 96041 wrap; 80001+404 not-found helper | `client/apikey.go` | Small, no breaking change |
| **2. Resource delta** | `description` attr (Default `""`); restore `all_volume`/`volume_ids` + validations; schema v0→v1 `StateUpgrader`; order-stable populate; FC1 validator loosening | `internal/provider/apikey_resource.go` | The state upgrader is the risky part — test with a real v0 state fixture |
| **3. Data source** | `zillizcloud_api_keys` | `internal/provider/apikeys_data_source.go` | Depends on Phase 1 pagination |
| **4. Tests** | Extend acceptance: description set/clear/drift, multi-project order stability, volume scoping, import, Billing-Admin probe (O3), v0 state upgrade | `apikey_resource_test.go`, new DS test | Use real API for acceptance |
| **5. Docs** | Regenerate resource doc; new data source doc; bootstrap/rotation guide (two-phase pattern + rotation caveat) | `docs/`, `examples/` | |

## Appendix: Live Test Transcript (2026-07-22)

> **Credentials note**: tests were run with a **temporary** Org Owner key (internal Vault: `secret/zilliz/terraform-apikey-test`, scheduled for deletion after the feature ships). All IDs and secrets below are redacted placeholders — never put real key material in this repo.

### Test 1: List (camelCase path) ✅
```bash
GET /v2/apiKeys?currentPage=1&pageSize=10
→ 200, count=56, 10 keys returned
```

### Test 2: List (kebab-case path) ❌
```bash
GET /v2/api-keys?currentPage=1&pageSize=10
→ "no Route matched with those values"
```

### Test 3: Create (name-based roles) ✅
```bash
POST /v2/apiKeys {"name":"terraform-test-key","orgRole":"Member","projects":[{"projectId":"proj-<project-id>","role":"Read-Only","allCluster":true,"allVolume":true}]}
→ 200, apiKeyId=key-<created-key-id>, apiKey=<64-hex-secret>
```

### Test 4: Describe ✅
```bash
GET /v2/apiKeys/key-<created-key-id>
→ 200, full details including createdBy=key-<creator-key-id> (creator key ID, not email)
```

### Test 5: Update ✅
```bash
PUT /v2/apiKeys/key-<created-key-id> {"name":"terraform-test-key-updated","description":"Updated via API test","orgRole":"Member","projects":[{"projectId":"proj-<project-id>","role":"Read-Write","allCluster":true,"allVolume":true}]}
→ 200, name/description/role all updated in-place
```

### Test 6: Auth with created key (project-scoped → blocked) ✅
```bash
GET /v2/apiKeys (using project-scoped key)
→ code=96041, "Only API keys with Org Owner permission can access this endpoint"
```

### Test 7: Delete ✅
```bash
DELETE /v2/apiKeys/key-<created-key-id>
→ 200, returns deleted key details
```

### Test 8: Verify delete ✅
```bash
GET /v2/apiKeys/key-<created-key-id>
→ code=404, "API key not found"
```

### Test 9: Create Owner key (blocked) ✅
```bash
POST /v2/apiKeys {"name":"owner-test","orgRole":"Owner"}
→ code=46001, "Creating API key with Org Owner permission via API is not allowed"
```

## References

- Public API gateway: `api.cloud.zilliz.com/v2/apiKeys`
- Internal backend: `cloud-control-api/.../OpenApiKeyController.java` (different schema — gateway translates)
- Merged implementation: PR #201 (`ba14766`, `3eba148`) — `client/apikey.go` + `internal/provider/apikey_resource.go`
- API key docs: `zdoc/docs/tutorials/security/authentication.md`
- Design pattern reference: `designs/tiered-nodegroup-support.md` (same repo)

---
page_title: "Managing Service Accounts"
subcategory: ""
description: |-
  Manage Service Accounts, permissions, and API keys with Terraform.
---

# Managing Service Accounts

Use `zillizcloud_service_account` to manage an account and its permissions, and
`zillizcloud_service_account_api_key` to manage each credential belonging to it.
Reference the account ID from the key resource so Terraform deletes keys before
it deletes the account. Keys created outside Terraform must also be deleted
before the account can be deleted.

## Account scope and permissions

An organization account requires exactly one of `org_role_id` or
`project_assignments`. `org_role_id` is a single predefined Billing Admin or Org
Owner role ID, not a list or a role name. An Org Owner credential is required to
create organization accounts or replace their access. Use `project_assignments = {}`
to create an account without access.

```terraform
resource "zillizcloud_service_account" "ingestion" {
  name = "ingestion-service"
  project_assignments = {
    "proj-a" = ["r-project-writer"]
    "proj-b" = []
  }
}

resource "zillizcloud_service_account_api_key" "ingestion" {
  service_account_id = zillizcloud_service_account.ingestion.id
  name               = "ingestion-key"
  expire_days        = 90
}
```

The account resource owns its complete permission configuration. Removing a
project from `project_assignments` removes its assignment and roles. Keeping the
project with `[]` clears only its roles. To clear all access, omit `org_role_id`
and set `project_assignments = {}`. Adding or removing projects, changing the
organization role, or switching between organization and project access uses
`PUT /access` to replace the complete configuration in one request and requires
an Org Owner credential. Changing between an organization role and project access
removes all previous access.

When an organization account has no organization role and its assigned projects
stay the same, changes to project roles use `PATCH /projectAssignments/{projectId}`
for each changed project. Unchanged projects are not written. These updates require
`service_account.modify` and `project_role.grant` in each changed project, or Org
Owner permissions. Multiple project updates are sequential; if a later update
fails, the provider refreshes the actual permissions so the next apply can finish
the remaining changes. Do not manage these same grants through another tool.

For a project-local account, set `project_scope` and optionally
`project_role_ids`. It always belongs to that project and cannot be assigned to
other projects or granted organization roles. Changing `project_scope` replaces
the account. Role changes use `PATCH /projectAssignments/{projectId}` and leave
the owning project assignment intact, including when the role set is empty.

## Disable and re-enable

`status = "disable"` makes all associated keys unusable. To edit permissions or
metadata on a disabled account, set `status = "active"` in the same apply or
apply that status change first. The provider enables the account before editing
it and disables it after any requested edits. It never temporarily enables an
account whose desired status is still `disable`.

Create or update keys only while their account is active. Apply an account's
status change separately from key changes when disabling it. Re-enabling an
account does not renew expired keys.

## Credentials and import

New keys require `expire_days` between 1 and 730. Changing this value replaces
the key; it does not extend the existing credential. For rotation, consider
`lifecycle { create_before_destroy = true }` and update credential consumers
before retiring the old key.

The key's `secret` is sensitive and returned only when it is created. Terraform
stores it in state. Restrict access to your state backend. Refresh and metadata
updates preserve that value; data sources and import cannot recover it.

```shell
terraform import zillizcloud_service_account.ingestion sa-example
terraform import zillizcloud_service_account_api_key.ingestion sa-example/key-example
```

For an imported account, match its scope, permissions, and status in configuration.
Empty project permissions are stored as `[]` or `{}`; `org_role_id` is null when
no organization role is granted. An organization account with no permissions
requires `project_assignments = {}`. For a project-local account with no roles, omit
`project_role_ids` or set it to `[]`. A matching configuration produces no changes,
including when the account is disabled; importing does not enable the account.

For an imported key, set `expire_days` to its original lifetime. Terraform derives
that lifetime from the creation and expiration timestamps when it is an exact
number of 24-hour days. A different lifetime, a legacy unlimited lifetime, or a
non-integer-day lifetime results in replacement. The imported secret remains
null until the key is replaced.

The legacy `zillizcloud_api_key` resource remains available. Its state cannot be
moved directly to the new resource types because their schemas and identities
differ. Existing service accounts and keys can be imported using the commands
above after removing their old Terraform ownership without destroying them.

## Updating an earlier SA configuration

The previous `org_role_ids` field is now `org_role_id`. Replace a one-element
set with its single ID string. Replace `org_role_ids = []` with
`project_assignments = {}`. Data sources also expose the singular `org_role_id`.
The provider upgrades existing account state automatically when it contains zero
or one organization role. Multiple roles require reconciling the account to the
current API's supported access model and re-importing it; no role is silently
chosen or dropped. API key resources and their stored secrets are unchanged.

## Queries

Use `zillizcloud_service_account` and `zillizcloud_service_account_api_key` data
sources to look up individual records. The plural `zillizcloud_service_accounts`
and `zillizcloud_service_account_api_keys` data sources follow pagination and
return matching records in `items`. Name and ID filters are substring matches;
when both are supplied they are combined with AND. Queries return metadata,
never secrets.

## Transient failures

Reads, complete access replacements, project role updates, safely repeatable
metadata/status updates, assignments, unassignments, and deletes retry transient network failures and HTTP 429, 500, 502, 503, and 504
responses. Requests use the provider's shared backoff implementation, with at
most four attempts and a one-minute total deadline per operation. Cancellation
stops both an in-flight request and its retry wait. Permission, validation, and
not-found errors are returned immediately.

Creating an account or API key is not automatically retried because the API does
not provide an idempotency key. The account resource sends metadata and status
changes separately; combined metadata/status API requests are not retried because
metadata cannot be edited after disabling an account.

Retries can recover a transient key synchronization failure after a database
commit. They do not guarantee recovery after the retry limit is exhausted or the
Terraform process exits; reliable completion still depends on the backend's
synchronization mechanism.

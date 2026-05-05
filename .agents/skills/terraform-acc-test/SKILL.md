---
name: terraform-acc-test
description: Run Terraform provider acceptance tests from a local Go provider repository. Use when the user asks Codex to run Terraform acc tests, acceptance tests, TF_ACC tests, or a command like `TF_ACC=1 go test`, especially when they mention loading `.env.test` or `.evn.test`, locating a test case package, or running `TestAccOnDemandClusterResource`.
---

# Terraform Acc Test

## Overview

Run Terraform provider acceptance tests with the repository's test environment loaded first. Prefer the bundled runner because it locates the requested test function with `rg` before choosing the Go package.

## Workflow

1. Check the current repository and confirm the requested test pattern.
2. Locate the test function before running Go tests:
   `rg -n "func[[:space:]]+TestName[[:space:]]*\\(" --glob '*_test.go' .`
3. Use the containing directory as the Go package unless the user explicitly provided `--package`.
4. Load `.env.test` from the repo root. If only `.evn.test` exists, use it and treat the spelling as intentional.
5. Run the test with `TF_ACC=1`, `go test -count=1`, verbose output, and a long timeout.
6. Keep the process open until it exits. Report the located package, command, pass/fail result, and the relevant failure lines if it fails.

## Quick Start

From the repository root:

```bash
.agents/skills/terraform-acc-test/scripts/run_acc_test.sh
```

Defaults:

- Package: auto-located from the requested test function with `rg`
- Test pattern: `TestAccOnDemandClusterResource`
- Timeout: `120m`
- Count: `1`

Override values:

```bash
.agents/skills/terraform-acc-test/scripts/run_acc_test.sh \
  --run TestAccOnDemandClusterResource \
  --timeout 120m \
  --count 1
```

Use `--package` only when auto-location is ambiguous or when running a broader package pattern:

```bash
.agents/skills/terraform-acc-test/scripts/run_acc_test.sh \
  --package ./internal/on_demand_cluster/... \
  --run TestAccOnDemandClusterResource
```

Pass extra `go test` flags after `--`:

```bash
.agents/skills/terraform-acc-test/scripts/run_acc_test.sh -- --failfast
```

## Direct Command

Use this if the user provides an exact command and wants no runner script:

```bash
test_file="$(rg -l --glob '*_test.go' 'func[[:space:]]+TestAccOnDemandClusterResource[[:space:]]*\(' .)"
test_package="./$(dirname "${test_file#./}")"
set -a
source .env.test
set +a
TF_ACC=1 go test -count=1 "$test_package" -run TestAccOnDemandClusterResource -v -timeout 120m
```

If `.env.test` is missing but `.evn.test` exists, replace `source .env.test` with `source .evn.test`.

## Notes

- Acceptance tests may create or delete real cloud resources. Run them only when the user asks for acceptance tests or `TF_ACC=1`.
- Do not print secrets from `.env.test` or `.evn.test`.
- If the worktree is dirty, do not revert unrelated changes before running tests.
- If `rg` finds no matching test function, stop and report that the test could not be located instead of running a stale default package.
- For long-running tests, use a TTY session and poll until completion.

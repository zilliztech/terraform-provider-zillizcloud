# Project Endpoints Documentation

This document describes the standard project endpoints added to the mock server.

## Overview

The mock server now supports standard project operations that mirror the Zilliz Cloud API for managing projects. These endpoints complement the existing BYOC (Bring Your Own Cloud) project endpoints.

## Endpoints

### 1. Create Project
**POST** `/v2/projects`

Creates a new standard project.

**Request Body:**
```json
{
  "projectName": "my-project",
  "plan": "Standard"
}
```

**Fields:**
- `projectName` (required): Name of the project
- `plan` (optional): Plan type (e.g., "Standard", "Enterprise", "Business Critical", "BYOC"). Defaults to "Enterprise" if not specified.

**Response:**
```json
{
  "code": 0,
  "data": "proj-abc123def456789012345"
}
```

Returns the newly created project ID.

---

### 2. List Projects
**GET** `/v2/projects`

Lists all projects in the account.

**Response:**
```json
{
  "code": 0,
  "data": [
    {
      "projectName": "Default Project",
      "projectId": "proj-ebc5ac7f430702aec8c57b",
      "instanceCount": 0,
      "createTimeMilli": 1703745469000,
      "plan": "Enterprise"
    },
    {
      "projectName": "my-project",
      "projectId": "proj-abc123def456789012345",
      "instanceCount": 0,
      "createTimeMilli": 1730000000000,
      "plan": "Standard"
    }
  ]
}
```

---

### 3. Get Project by ID
**GET** `/v2/projects/:projectId`

Retrieves a specific project by its ID.

**URL Parameters:**
- `projectId`: The project ID

**Response:**
```json
{
  "code": 0,
  "data": {
    "projectName": "my-project",
    "projectId": "proj-abc123def456789012345",
    "instanceCount": 0,
    "createTimeMilli": 1730000000000,
    "plan": "Standard"
  }
}
```

**Error Response (404):**
```json
{
  "code": 404,
  "message": "Project with ID proj-xyz not found"
}
```

---

### 4. Upgrade Project Plan
**PATCH** `/v2/projects/:projectId/plan`

Upgrades the plan for an existing project.

**URL Parameters:**
- `projectId`: The project ID

**Request Body:**
```json
{
  "plan": "Enterprise"
}
```

**Response:**
```json
{
  "code": 0,
  "data": "Project proj-abc123def456789012345 plan upgraded to Enterprise"
}
```

---

### 5. Delete Project
**DELETE** `/v2/projects/:projectId`

Deletes a project.

**URL Parameters:**
- `projectId`: The project ID

**Response:**
```json
{
  "code": 0,
  "data": "Project proj-abc123def456789012345 deleted successfully"
}
```

**Error Response (404):**
```json
{
  "code": 404,
  "message": "Project with ID proj-xyz not found"
}
```

---

## Data Model

### Project Structure

```go
type Project struct {
    ProjectName     string `json:"projectName"`     // Name of the project
    ProjectId       string `json:"projectId"`       // Unique identifier (format: proj-{uuid})
    InstanceCount   int64  `json:"instanceCount"`   // Number of cluster instances
    CreateTimeMilli int64  `json:"createTimeMilli"` // Creation timestamp in milliseconds
    Plan            string `json:"plan"`            // Plan type (Standard, Enterprise, etc.)
}
```

## Testing

A test script is provided to verify all project endpoints:

```bash
# Start the mock server
cd mock-server
go run main.go

# In another terminal, run the test script
./test/project_test.sh
```

The test script performs the following operations:
1. Lists all projects (initial state with default project)
2. Creates a new project with Standard plan
3. Gets the created project by ID
4. Lists all projects (should now have 2)
5. Upgrades the project plan to Enterprise
6. Verifies the plan upgrade
7. Creates a project with default plan (should default to Enterprise)
8. Verifies the default plan
9. Deletes a project
10. Attempts to get the deleted project (should return 404)
11. Lists all projects (deleted project should not appear)

## Default Data

The mock server initializes with one default project:

```json
{
  "projectName": "Default Project",
  "projectId": "proj-ebc5ac7f430702aec8c57b",
  "instanceCount": 0,
  "createTimeMilli": 1703745469000,
  "plan": "Enterprise"
}
```

## Implementation Details

### Files Modified/Created

1. **Created: `pkg/byoc_project/project.go`**
   - Contains all handler functions for project operations
   - Implements thread-safe operations using the existing `safeStore`

2. **Modified: `pkg/byoc_project/model.go`**
   - Updated `Project` struct with `Plan` and `CreateTimeMilli` fields
   - Added `CreateProjectRequest` and `UpgradeProjectPlanRequest` types

3. **Modified: `pkg/byoc_project/store.go`**
   - Updated default project initialization with `Plan` field
   - Added `Delete()` method to `safeStore` for project deletion

4. **Modified: `pkg/byoc_project/cluster.go`**
   - Removed duplicate `GetProjects()` function (moved to project.go)

5. **Modified: `main.go`**
   - Added routes for standard project endpoints under `/v2/projects`

6. **Created: `test/project_test.sh`**
   - Comprehensive test script for all project endpoints

## Notes

- Project IDs are generated using the format `proj-{uuid}` where uuid is a shortened UUID (22 characters)
- The mock server stores all data in memory - data is lost when the server restarts
- All operations are thread-safe using `sync.Map` under the hood
- Error responses follow the standard format with `code` and `message` fields
- Success responses follow the format with `code: 0` and `data` field

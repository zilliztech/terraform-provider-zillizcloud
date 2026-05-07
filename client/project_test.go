package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClient_ListProjects(t *testing.T) {
	c, requests := testProjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Fatalf("path = %s, want /projects", r.URL.Path)
		}
		writeProjectResponse(t, w, []Project{{
			ProjectId:     "proj-1",
			ProjectName:   "Project One",
			InstanceCount: 2,
			CreateTime:    "2026-05-07T08:00:00Z",
			Plan:          "Standard",
			OrgType:       "SAAS",
		}})
	})

	projects, err := c.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects error: %v", err)
	}
	if len(*requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(*requests))
	}
	if len(projects) != 1 {
		t.Fatalf("project count = %d, want 1", len(projects))
	}
	if projects[0].ProjectId != "proj-1" || projects[0].ProjectName != "Project One" || projects[0].Plan != "Standard" || projects[0].OrgType != "SAAS" {
		t.Fatalf("decoded project = %#v", projects[0])
	}
}

func TestClient_CreateProject(t *testing.T) {
	c, _ := testProjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Fatalf("path = %s, want /projects", r.URL.Path)
		}
		var got CreateProjectRequest
		decodeProjectRequest(t, r, &got)
		want := CreateProjectRequest{
			ProjectName: "test-project",
			Plan:        "Enterprise",
			Regions:     []string{"aws-us-east-1", "gcp-us-west1"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %#v, want %#v", got, want)
		}
		writeProjectResponse(t, w, "proj-created")
	})

	projectID, err := c.CreateProject(&CreateProjectRequest{
		ProjectName: "test-project",
		Plan:        "Enterprise",
		Regions:     []string{"aws-us-east-1", "gcp-us-west1"},
	})
	if err != nil {
		t.Fatalf("CreateProject error: %v", err)
	}
	if projectID == nil || *projectID != "proj-created" {
		t.Fatalf("projectID = %v, want proj-created", projectID)
	}
}

func TestClient_GetProjectById(t *testing.T) {
	c, _ := testProjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/projects/proj-1" {
			t.Fatalf("path = %s, want /projects/proj-1", r.URL.Path)
		}
		writeProjectResponse(t, w, Project{
			ProjectId:     "proj-1",
			ProjectName:   "Project One",
			InstanceCount: 3,
			CreateTime:    "2026-05-07T08:00:00Z",
			Plan:          "BusinessCritical",
			OrgType:       "SAAS",
		})
	})

	project, err := c.GetProjectById("proj-1")
	if err != nil {
		t.Fatalf("GetProjectById error: %v", err)
	}
	if project == nil {
		t.Fatal("project is nil")
	}
	if project.ProjectId != "proj-1" || project.ProjectName != "Project One" || project.Plan != "BusinessCritical" || project.InstanceCount != 3 {
		t.Fatalf("decoded project = %#v", project)
	}
}

func TestClient_UpgradeProjectPlanRequest(t *testing.T) {
	c, _ := testProjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/projects/proj-1/plan" {
			t.Fatalf("path = %s, want /projects/proj-1/plan", r.URL.Path)
		}
		var got UpgradeProjectPlanRequest
		decodeProjectRequest(t, r, &got)
		if got.Plan != "Enterprise" {
			t.Fatalf("plan = %s, want Enterprise", got.Plan)
		}
		writeProjectResponse(t, w, "proj-1")
	})

	projectID, err := c.UpgradeProjectPlan("proj-1", "Enterprise")
	if err != nil {
		t.Fatalf("UpgradeProjectPlan error: %v", err)
	}
	if projectID == nil || *projectID != "proj-1" {
		t.Fatalf("projectID = %v, want proj-1", projectID)
	}
}

func TestClient_AddProjectRegions(t *testing.T) {
	c, _ := testProjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/projects/proj-1/regions" {
			t.Fatalf("path = %s, want /projects/proj-1/regions", r.URL.Path)
		}
		var got AddProjectRegionsRequest
		decodeProjectRequest(t, r, &got)
		want := AddProjectRegionsRequest{Regions: []string{"gcp-us-west1"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("body = %#v, want %#v", got, want)
		}
		writeProjectResponse(t, w, []string{"aws-us-east-1", "gcp-us-west1"})
	})

	regions, err := c.AddProjectRegions("proj-1", []string{"gcp-us-west1"})
	if err != nil {
		t.Fatalf("AddProjectRegions error: %v", err)
	}
	want := []string{"aws-us-east-1", "gcp-us-west1"}
	if !reflect.DeepEqual(regions, want) {
		t.Fatalf("regions = %#v, want %#v", regions, want)
	}
}

func testProjectClient(t *testing.T, handler http.HandlerFunc) (*Client, *[]*http.Request) {
	t.Helper()

	requests := make([]*http.Request, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Clone(r.Context()))
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	c, err := NewClient(
		WithApiKey("test-api-key"),
		WithBaseUrl(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	return c, &requests
}

func decodeProjectRequest(t *testing.T, r *http.Request, target any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

func writeProjectResponse(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"code": 0,
		"data": data,
	}); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

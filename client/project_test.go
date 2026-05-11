package client

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestClient_ListProject(t *testing.T) {

	t.Run("ListProjects", func(t *testing.T) {
		c, teardown := zillizClient[[]Project](t)
		defer teardown()

		got, err := c.ListProjects()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		if len(got) == 0 {
			t.Errorf("want = %v, got = %v", len(got), got)
		}

	})

	t.Run("get project list with wrong api key", func(t *testing.T) {
		var tmp string
		if apiKey != "" {
			tmp = apiKey
		}
		defer func() {
			apiKey = tmp
		}()

		apiKey = "gibberish_api_key"
		c, teardown := zillizClient[[]Project](t)
		defer teardown()

		_, err := c.ListProjects()
		var apierr = Error{
			Code: 21119,
		}
		if !errors.Is(err, apierr) {
			t.Errorf("want = %v, got = %v", apierr, err)
		}
	})

}

func TestProject_UnmarshalRegionFields(t *testing.T) {
	var response zillizResponse[[]Project]
	err := json.Unmarshal([]byte(`{"code":0,"data":[{"projectId":"proj-001","projectName":"test-project","instanceCount":0,"createTimeMilli":1762324757000,"plan":"Enterprise","regionIds":["aws-us-east-1","gcp-us-west1"]}]}`), &response)
	if err != nil {
		t.Fatalf("failed to unmarshal project response: %v", err)
	}

	if len(response.Data) != 1 {
		t.Fatalf("want 1 project, got %d", len(response.Data))
	}

	project := response.Data[0]
	if len(project.RegionIds) != 2 || project.RegionIds[0] != "aws-us-east-1" || project.RegionIds[1] != "gcp-us-west1" {
		t.Fatalf("unexpected region ids: %v", project.RegionIds)
	}
}

func TestClient_CreateProject(t *testing.T) {

	t.Run("CreateProject", func(t *testing.T) {
		c, teardown := zillizClient[string](t)
		defer teardown()

		// Note: This test might fail if you don't have permission to create projects
		// or if project with this name already exists
		req := &CreateProjectRequest{
			ProjectName: "test-project",
			Plan:        "Standard",
		}

		resp, err := c.CreateProject(req)
		if err != nil {
			t.Fatalf("failed to create project: %v", err)
			t.Skipf("failed to create project (this may be expected if API doesn't support project creation): %v", err)
			return
		}

		if resp == nil {
			t.Fatal("expected non-nil response")
			return
		}

		if *resp == "" {
			t.Error("expected non-empty ProjectId")
		}

		t.Logf("Created project with ID: %s", *resp)
	})

}

func TestClient_GetProjectById(t *testing.T) {

	t.Run("GetProjectById", func(t *testing.T) {
		c, teardown := zillizClient[Project](t)
		defer teardown()

		project, err := c.GetProjectById("proj-d4477e7af43bbeb44594d9")
		if err != nil {
			t.Fatalf("failed to get project: %v", err)
		}

		if project == nil {
			t.Fatal("expected non-nil project")
			return
		}

		if project.ProjectName != "test-project" {
			t.Errorf("want = %s, got = %s", "test-project", project.ProjectName)
		}
	})
}

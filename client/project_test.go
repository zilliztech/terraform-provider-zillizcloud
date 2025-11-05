package client

import (
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
		}

		if project.ProjectName != "test-project" {
			t.Errorf("want = %s, got = %s", "test-project", project.ProjectName)
		}
	})
}

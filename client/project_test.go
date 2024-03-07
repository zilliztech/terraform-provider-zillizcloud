package client

import (
	"flag"
	"testing"
)

var (
	apiKey string
)

func init() {
	flag.StringVar(&apiKey, "key", "", "Your TEST secret key for the zilliz cloud API. If present, integration tests will be run using this key.")
}

func TestClient_ListProjects(t *testing.T) {
	if apiKey == "" {
		t.Skip("No API key provided")
	}

	type checkFn func(*testing.T, []Project, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoErr := func() checkFn {
		return func(t *testing.T, _ []Project, err error) {
			if err != nil {
				t.Fatalf("err = %v; want nil", err)
			}
		}
	}

	hasErrCode := func(code int) checkFn {
		return func(t *testing.T, _ []Project, err error) {
			se, ok := err.(Error)
			if !ok {
				t.Fatalf("err isn't a Error")
			}
			if se.Code != code {
				t.Errorf("err.Code = %d; want %d", se.Code, code)
			}
		}
	}

	hasProject := func(Name string) checkFn {
		return func(t *testing.T, p []Project, err error) {
			for _, project := range p {
				if project.ProjectName == Name {
					return
				}
			}
			t.Errorf("project not found: %s", Name)
		}
	}

	type fields struct {
		CloudRegionId string
		apiKey        string
	}
	tests := []struct {
		name   string
		fields fields
		checks []checkFn
	}{
		{
			"postive 1",
			fields{CloudRegionId: "gcp-us-west1", apiKey: apiKey},
			check(
				hasNoErr(),
				hasProject("Default Project")),
		},
		{
			"none exist region",
			fields{CloudRegionId: "gcp-us-west1", apiKey: "fake"},
			check(hasErrCode(80001)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient(
				tt.fields.apiKey,
				tt.fields.CloudRegionId,
			)

			got, err := c.ListProjects()
			for _, check := range tt.checks {
				check(t, got, err)
			}
		})
	}
}

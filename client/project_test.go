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

		want := "Default Project"

		if len(got) == 0 || got[0].ProjectName != want {
			t.Errorf("want = %s, got = %v", want, got)
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

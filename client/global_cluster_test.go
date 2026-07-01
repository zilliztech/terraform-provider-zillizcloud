package client

import (
	"net"
	"net/http"
	"testing"
)

func TestUnitDeleteCluster(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-secondary" {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-secondary", "prompt": "deleted"},
		}), nil
	})

	resp, err := c.DeleteCluster("glo-1", "in01-secondary")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.GlobalClusterId != "glo-1" || resp.ClusterId != "in01-secondary" || resp.Prompt != "deleted" {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDeleteClusterRequiresIDs(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, &net.OpError{Op: "unexpected request"}
	})

	if _, err := c.DeleteCluster("", "in01-secondary"); err == nil {
		t.Fatal("expected global cluster ID error")
	}
	if _, err := c.DeleteCluster("glo-1", ""); err == nil {
		t.Fatal("expected cluster ID error")
	}
}

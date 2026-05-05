package client

import (
	"encoding/json"
	"net"
	"net/http"
	"testing"
)

func TestUnitCreateOnDemandCluster(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/clusters/createOnDemandCluster" {
			t.Errorf("path=%s", req.URL.Path)
		}

		var body CreateOnDemandClusterRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.ProjectID != "proj-1" || body.RegionID != "aws-us-west-2" || body.CUSize != 8 || body.ClusterName != "query-dev" {
			t.Errorf("body=%+v", body)
		}
		if body.AutoSuspend == nil || *body.AutoSuspend != 1800 {
			t.Errorf("autoSuspend=%v", body.AutoSuspend)
		}
		if body.MaxQueryNodeCU == nil || *body.MaxQueryNodeCU != 8 {
			t.Errorf("maxQueryNodeCU=%v", body.MaxQueryNodeCU)
		}
		if body.MaxQueryNodeReplicas == nil || *body.MaxQueryNodeReplicas != 4 {
			t.Errorf("maxQueryNodeReplicas=%v", body.MaxQueryNodeReplicas)
		}

		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"clusterId": "in07-qc-1", "prompt": "Creating"},
		}), nil
	})

	maxCU := 8
	maxReplicas := 4
	autoSuspend := 1800
	resp, err := c.CreateOnDemandCluster(&CreateOnDemandClusterRequest{
		ProjectID:            "proj-1",
		RegionID:             "aws-us-west-2",
		CUSize:               8,
		AutoSuspend:          &autoSuspend,
		MaxQueryNodeCU:       &maxCU,
		MaxQueryNodeReplicas: &maxReplicas,
		ClusterName:          "query-dev",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.ClusterId != "in07-qc-1" || resp.Prompt != "Creating" {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDescribeOnDemandCluster(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/clusters/onDemandClusters/in07-qc-1" {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{
				"clusterId":   "in07-qc-1",
				"status":      "RUNNING",
				"regionId":    "aws-us-west-2",
				"autoSuspend": 1800,
			},
		}), nil
	})

	resp, err := c.DescribeOnDemandCluster("in07-qc-1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.ClusterID != "in07-qc-1" || resp.Status != "RUNNING" {
		t.Errorf("resp=%+v", resp)
	}
	if resp.AutoSuspend == nil || *resp.AutoSuspend != 1800 {
		t.Errorf("autoSuspend=%v", resp.AutoSuspend)
	}
}

func TestUnitListOnDemandClusters(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/clusters/onDemandClusters" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.Query().Get("projectId") != "proj-1" {
			t.Errorf("projectId=%s", req.URL.Query().Get("projectId"))
		}
		if req.URL.Query().Get("regionId") != "aws-us-west-2" {
			t.Errorf("regionId=%s", req.URL.Query().Get("regionId"))
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{
				"onDemandClusters": []map[string]any{
					{"clusterId": "in07-qc-1", "regionId": "aws-us-west-2", "status": "RUNNING"},
				},
				"count": 1,
			},
		}), nil
	})

	resp, err := c.ListOnDemandClusters("proj-1", "aws-us-west-2")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.Count != 1 || len(resp.OnDemandClusters) != 1 {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDeleteOnDemandCluster(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "DELETE" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/clusters/onDemandClusters/in07-qc-1" {
			t.Errorf("path=%s", req.URL.Path)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"clusterId": "in07-qc-1", "prompt": "Deleting"},
		}), nil
	})

	resp, err := c.DeleteOnDemandCluster("in07-qc-1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.ClusterID != "in07-qc-1" || resp.Prompt != "Deleting" {
		t.Errorf("resp=%+v", resp)
	}
}

func TestUnitDescribeOnDemandClusterRequiresID(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		return nil, &net.OpError{Op: "unexpected request"}
	})

	if _, err := c.DescribeOnDemandCluster(""); err == nil {
		t.Fatal("expected error")
	}
}

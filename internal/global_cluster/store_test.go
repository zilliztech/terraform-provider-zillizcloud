package global_cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type storeMockHTTPClient struct {
	t    *testing.T
	call int
	do   func(call int, req *http.Request, body []byte) (*http.Response, error)
}

func (m *storeMockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.t.Helper()
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			m.t.Fatalf("ReadAll request body: %v", err)
		}
	}
	m.call++
	return m.do(m.call, req, body)
}

func newStoreTestClient(t *testing.T, do func(call int, req *http.Request, body []byte) (*http.Response, error)) *zilliz.Client {
	t.Helper()
	client, err := zilliz.NewClient(
		zilliz.WithApiKey("test-key"),
		zilliz.WithBaseUrl("https://api.test/v2"),
		zilliz.WithHTTPClient(&storeMockHTTPClient{t: t, do: do}),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func storeJSONResponse(t *testing.T, statusCode int, body any) *http.Response {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal response: %v", err)
	}
	return &http.Response{StatusCode: statusCode, Body: io.NopCloser(bytes.NewReader(payload)), Header: http.Header{"Content-Type": []string{"application/json"}}}
}

func TestStoreCreateAllowsEmptyMembersForBackendValidation(t *testing.T) {
	store := NewGlobalClusterStore(newStoreTestClient(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 || req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/create" {
			t.Fatalf("unexpected request %d %s %s", call, req.Method, req.URL.Path)
		}
		var createReq zilliz.CreateGlobalClusterParams
		if err := json.Unmarshal(body, &createReq); err != nil {
			t.Fatalf("Unmarshal create request: %v", err)
		}
		if createReq.PrimaryCluster.ClusterName != "" || len(createReq.SecondaryClusters) != 0 {
			t.Fatalf("unexpected member params: %+v", createReq)
		}
		return storeJSONResponse(t, http.StatusBadRequest, map[string]any{"code": 400, "message": "primaryCluster clusterName and regionId are required"}), nil
	}))

	_, err := store.Create(context.Background(), CreateGlobalClusterCommand{GlobalClusterName: "global-a", ProjectID: "proj-1", CUType: "Performance-optimized", CUSize: 4})
	if err == nil {
		t.Fatalf("expected backend validation error")
	}
}

func TestStoreCreateSendsAutoscalingAndMemberReplicas(t *testing.T) {
	store := NewGlobalClusterStore(newStoreTestClient(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 || req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/create" {
			t.Fatalf("unexpected request %d %s %s", call, req.Method, req.URL.Path)
		}
		var createReq zilliz.CreateGlobalClusterParams
		if err := json.Unmarshal(body, &createReq); err != nil {
			t.Fatalf("Unmarshal create request: %v", err)
		}
		if createReq.CuSize != 0 {
			t.Fatalf("cuSize must be omitted for CU autoscaling, got %d", createReq.CuSize)
		}
		if createReq.Autoscaling == nil || createReq.Autoscaling.CU == nil || createReq.Autoscaling.Replica == nil {
			t.Fatalf("autoscaling policies missing: %+v", createReq.Autoscaling)
		}
		if *createReq.Autoscaling.CU.Min != 4 || *createReq.Autoscaling.CU.Max != 8 {
			t.Fatalf("unexpected CU autoscaling: %+v", createReq.Autoscaling.CU)
		}
		if *createReq.Autoscaling.Replica.Min != 1 || *createReq.Autoscaling.Replica.Max != 3 {
			t.Fatalf("unexpected replica autoscaling: %+v", createReq.Autoscaling.Replica)
		}
		if createReq.PrimaryCluster.Replica == nil || *createReq.PrimaryCluster.Replica != 2 {
			t.Fatalf("unexpected primary replica: %+v", createReq.PrimaryCluster)
		}
		if len(createReq.SecondaryClusters) != 1 || createReq.SecondaryClusters[0].Replica == nil || *createReq.SecondaryClusters[0].Replica != 1 {
			t.Fatalf("unexpected secondary replicas: %+v", createReq.SecondaryClusters)
		}
		return storeJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{
			"globalClusterId": "glo-1", "username": "db_admin", "password": "secret", "jobId": "job-1",
		}}), nil
	}))

	cuMin, cuMax := int64(4), int64(8)
	replicaMin, replicaMax := int64(1), int64(3)
	primaryReplica, secondaryReplica := int64(2), int64(1)
	created, err := store.Create(context.Background(), CreateGlobalClusterCommand{
		GlobalClusterName: "global-a",
		ProjectID:         "proj-1",
		CUType:            "Performance-optimized",
		Autoscaling: GlobalClusterAutoscaling{
			CU:      &GlobalClusterAutoscalingPolicy{Min: &cuMin, Max: &cuMax},
			Replica: &GlobalClusterAutoscalingPolicy{Min: &replicaMin, Max: &replicaMax},
		},
		Members: []GlobalClusterMemberSpec{
			{ClusterName: "primary-a", RegionID: "aws-us-west-2", Replica: &primaryReplica},
			{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1", Replica: &secondaryReplica},
		},
	})
	if err != nil {
		t.Fatalf("Create err: %v", err)
	}
	if created.GlobalClusterID != "glo-1" || created.JobID != "job-1" {
		t.Fatalf("unexpected result: %+v", created)
	}
}

func TestStoreDescribeReturnsDomainModel(t *testing.T) {
	store := NewGlobalClusterStore(newStoreTestClient(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		if call != 1 || req.Method != http.MethodGet || req.URL.Path != "/v2/globalClusters/glo-1" {
			t.Fatalf("unexpected request %d %s %s", call, req.Method, req.URL.Path)
		}
		return storeJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{
			"globalClusterId": "glo-1", "globalClusterName": "global-a", "projectId": "proj-1",
			"regionIds": []string{"aws-us-west-2"}, "cuType": "Performance-optimized", "cuSize": 4,
			"connectAddress": "https://glo-1.global-cluster.vectordb.zillizcloud.com", "createTime": "2026-06-04T10:00:00Z",
			"autoscaling": map[string]any{
				"cu":      map[string]any{"min": 4, "max": 8},
				"replica": map[string]any{"min": 1, "max": 3},
			},
			"clusters": []map[string]any{{"clusterId": "in01-primary", "clusterName": "primary-a", "regionId": "aws-us-west-2", "role": "PRIMARY", "status": "RUNNING", "replica": 2}},
		}}), nil
	}))

	got, err := store.Describe(context.Background(), "glo-1")
	if err != nil {
		t.Fatalf("Describe err: %v", err)
	}
	if got.GlobalClusterID != "glo-1" || len(got.Clusters) != 1 || got.Clusters[0].ClusterID != "in01-primary" {
		t.Fatalf("unexpected cluster: %+v", got)
	}
	if got.Autoscaling.CU == nil || *got.Autoscaling.CU.Max != 8 || got.Autoscaling.Replica == nil || *got.Autoscaling.Replica.Max != 3 {
		t.Fatalf("unexpected autoscaling: %+v", got.Autoscaling)
	}
	if got.Clusters[0].Replica != 2 {
		t.Fatalf("unexpected member replica: %+v", got.Clusters[0])
	}
}

func TestStoreAddAndDeleteSecondaries(t *testing.T) {
	store := NewGlobalClusterStore(newStoreTestClient(t, func(call int, req *http.Request, body []byte) (*http.Response, error) {
		switch call {
		case 1:
			if req.Method != http.MethodPost || req.URL.Path != "/v2/globalClusters/glo-1/secondaryClusters" {
				t.Fatalf("add request %s %s", req.Method, req.URL.Path)
			}
			var addReq zilliz.AddSecondaryClustersParams
			if err := json.Unmarshal(body, &addReq); err != nil {
				t.Fatalf("Unmarshal add request: %v", err)
			}
			if len(addReq.SecondaryClusters) != 1 || addReq.SecondaryClusters[0].ClusterName != "secondary-au" || addReq.SecondaryClusters[0].Replica != nil {
				t.Fatalf("unexpected add request: %+v", addReq.SecondaryClusters)
			}
			return storeJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"jobId": "job-add-1"}}), nil
		case 2:
			if req.Method != http.MethodDelete || req.URL.Path != "/v2/globalClusters/glo-1/clusters/in01-secondary" {
				t.Fatalf("delete request %s %s", req.Method, req.URL.Path)
			}
			return storeJSONResponse(t, http.StatusOK, map[string]any{"code": 0, "data": map[string]any{"globalClusterId": "glo-1", "clusterId": "in01-secondary", "prompt": "deleted"}}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", call)
		}
	}))

	replica := int64(2)
	if err := store.AddSecondaryClusters(context.Background(), "glo-1", []GlobalClusterMemberSpec{{ClusterName: "secondary-au", RegionID: "aws-ap-southeast-2", Replica: &replica}}); err != nil {
		t.Fatalf("AddSecondaryClusters err: %v", err)
	}
	if err := store.DeleteCluster(context.Background(), "glo-1", "in01-secondary"); err != nil {
		t.Fatalf("DeleteCluster err: %v", err)
	}
}

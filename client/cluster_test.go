package client

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zilliztech/terraform-provider-zillizcloud/internal/util/conv"
)

func TestClient_Cluster(t *testing.T) {
	var clusterId string
	var projectID string
	if update {
		pollInterval = 60
	}

	checkPlan := func(plan string) func(resp *Cluster) bool {
		return func(resp *Cluster) bool {
			return resp.Plan == plan
		}
	}

	checkCUSize := func(cuSize int64) func(resp *Cluster) bool {
		return func(resp *Cluster) bool {
			return resp.CuSize == cuSize
		}
	}

	checkCUType := func(cuType string) func(resp *Cluster) bool {
		return func(resp *Cluster) bool {
			return resp.CuType == cuType
		}
	}

	c, teardown := zillizClient[Clusters](t)
	defer teardown()

	getProject := func() string {

		projects, err := c.ListProjects()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		var want = "Default Project"

		if len(projects) == 0 || projects[0].ProjectName != want {
			t.Errorf("want = %s, got = %v", want, projects)
		}

		return projects[0].ProjectId
	}

	projectID = getProject()
	cuSize := 1
	params := CreateClusterParams{
		ProjectId:   projectID,
		Plan:        conv.StringPtr("Standard"),
		ClusterName: "a-standard-type-cluster",
		CUSize:      &cuSize,
		CUType:      "Performance-optimized",
		RegionId:    "gcp-us-west1",
	}

	t.Run("CreateCluster", func(t *testing.T) {
		c, teardown := zillizClient[Clusters](t)
		defer teardown()

		resp, err := c.CreateDedicatedCluster(params)
		if err != nil {
			t.Fatalf("failed to create cluster: %v", err)
		}

		if resp.ClusterId == "" {
			t.Fatalf("failed to create cluster: %v", resp)
		}

		clusterId = resp.ClusterId
	})

	t.Run("duplicately create cluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		_, err := c.CreateDedicatedCluster(params)

		var e = Error{
			Code: 40013,
		}

		if !errors.Is(err, e) {
			t.Fatalf("want = %v, but got = %v", e, err)
		}

	})

	t.Run("DescribeCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		// checkfn:=make([]func(resp *Cluster) bool,0)
		checkfn := []func(resp *Cluster) bool{
			checkPlan("Standard"),
			checkCUSize(1),
			checkCUType("Performance-optimized"),
		}

		ctx, cancelfn := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancelfn()
		got := pollClusterStatus(ctx, t, c, clusterId, "RUNNING", pollInterval)

		for _, fn := range checkfn {
			if !fn(got) {
				t.Errorf("check failed")
			}
		}

	})

	t.Run("scale up cluster", func(t *testing.T) {
		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		checkfn := []func(resp *Cluster) bool{
			checkCUSize(2),
		}
		_, err := c.ModifyCluster(clusterId, &ModifyClusterParams{

			CuSize: 2,
		})
		if err != nil {
			t.Fatalf("failed to describe cluster: %v", err)
		}

		ctx, cancelfn := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancelfn()
		got := pollClusterStatus(ctx, t, c, clusterId, "RUNNING", pollInterval)
		for _, fn := range checkfn {
			if !fn(got) {
				t.Errorf("check failed")
			}
		}

	})

	t.Run("DeleteCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		got, err := c.DropCluster(clusterId)
		if err != nil {
			t.Fatalf("failed to delete cluster: %v", err)
		}

		if got == nil || *got != clusterId {
			t.Fatalf("want = %s, got = %v", clusterId, got)
		}
	})
}

func TestCreateClusterParamsFixedCUJSON(t *testing.T) {
	cuSize := 4
	params := CreateClusterParams{
		ProjectId:   "proj",
		Plan:        conv.StringPtr("Enterprise"),
		ClusterName: "fixed-cu",
		CUSize:      &cuSize,
		CUType:      "Performance-optimized",
		RegionId:    "aws-us-west-2",
	}

	payload, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if got["cuSize"] != float64(4) {
		t.Fatalf("cuSize = %v, want 4", got["cuSize"])
	}
	if _, ok := got["autoscaling"]; ok {
		t.Fatalf("autoscaling should be omitted for fixed CU payload: %s", string(payload))
	}
}

func TestCreateClusterParamsAutoscalingJSON(t *testing.T) {
	minCU := 4
	maxCU := 8
	replica := 2
	params := CreateClusterParams{
		ProjectId:   "proj",
		Plan:        conv.StringPtr("Enterprise"),
		ClusterName: "dynamic-cu",
		CUType:      "Performance-optimized",
		RegionId:    "aws-us-west-2",
		Replica:     &replica,
		Autoscaling: &AutoscalingConfig{
			CU: &AutoscalingPolicy{
				Min: &minCU,
				Max: &maxCU,
			},
		},
	}

	payload, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if _, ok := got["cuSize"]; ok {
		t.Fatalf("cuSize should be omitted for autoscaling payload: %s", string(payload))
	}
	if got["replica"] != float64(2) {
		t.Fatalf("replica = %v, want 2", got["replica"])
	}
	autoscaling, ok := got["autoscaling"].(map[string]any)
	if !ok {
		t.Fatalf("autoscaling = %T, want map[string]any", got["autoscaling"])
	}
	cu, ok := autoscaling["cu"].(map[string]any)
	if !ok {
		t.Fatalf("autoscaling.cu = %T, want map[string]any", autoscaling["cu"])
	}
	if cu["min"] != float64(4) || cu["max"] != float64(8) {
		t.Fatalf("autoscaling.cu = %v, want min=4 max=8", cu)
	}
	if _, ok := autoscaling["replica"]; ok {
		t.Fatalf("autoscaling.replica should be omitted for create payload: %s", string(payload))
	}
}

func TestClient_ServerlessCluster(t *testing.T) {
	var clusterId string
	var projectID string
	if update {
		pollInterval = 60
	}

	checkClusterId := func(clusterId string) func(resp *Cluster) bool {
		return func(resp *Cluster) bool {
			return resp.ClusterId == clusterId
		}
	}

	c, teardown := zillizClient[Clusters](t)
	defer teardown()

	getProject := func() string {

		projects, err := c.ListProjects()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		var want = "Default Project"

		if len(projects) == 0 || projects[0].ProjectName != want {
			t.Errorf("want = %s, got = %v", want, projects)
		}

		return projects[0].ProjectId
	}

	projectID = getProject()
	params := CreateServerlessClusterParams{
		ProjectId:   projectID,
		ClusterName: "a-starter-type-cluster",
		RegionId:    "gcp-us-west1",
	}

	t.Run("create serverless cluster", func(t *testing.T) {
		c, teardown := zillizClient[Clusters](t)
		defer teardown()

		resp, err := c.CreateServerlessCluster(params)
		if err != nil {
			t.Fatalf("failed to create cluster: %v", err)
		}

		if resp.ClusterId == "" {
			t.Fatalf("failed to create cluster: %v", resp)
		}

		clusterId = resp.ClusterId
	})

	t.Run("duplicately create cluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()

		_, err := c.CreateServerlessCluster(params)
		var e = Error{
			Code: 40013,
		}

		if !errors.Is(err, e) {
			t.Fatalf("want = %v, but got = %v", e, err)
		}

	})

	t.Run("DescribeCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		// checkfn:=make([]func(resp *Cluster) bool,0)
		checkfn := []func(resp *Cluster) bool{
			checkClusterId(clusterId),
		}

		ctx, cancelfn := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancelfn()
		got := pollClusterStatus(ctx, t, c, clusterId, "RUNNING", pollInterval)

		for _, fn := range checkfn {
			if !fn(got) {
				t.Errorf("check failed")
			}
		}

	})

	t.Run("DeleteCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		got, err := c.DropCluster(clusterId)
		if err != nil {
			t.Fatalf("failed to delete cluster: %v", err)
		}

		if got == nil || *got != clusterId {
			t.Fatalf("want = %s, got = %v", clusterId, got)
		}
	})
}

func TestClient_FreeCluster(t *testing.T) {
	var clusterId string
	var projectID string
	if update {
		pollInterval = 60
	}

	checkClusterId := func(clusterId string) func(resp *Cluster) bool {
		return func(resp *Cluster) bool {
			return resp.ClusterId == clusterId
		}
	}

	c, teardown := zillizClient[Clusters](t)
	defer teardown()

	getProject := func() string {

		projects, err := c.ListProjects()
		if err != nil {
			t.Fatalf("failed to list projects: %v", err)
		}

		var want = "Default Project"

		if len(projects) == 0 || projects[0].ProjectName != want {
			t.Errorf("want = %s, got = %v", want, projects)
		}

		return projects[0].ProjectId
	}

	projectID = getProject()
	params := CreateServerlessClusterParams{
		ProjectId:   projectID,
		ClusterName: "a-starter-type-cluster",
		RegionId:    "gcp-us-west1",
	}

	t.Run("create free cluster", func(t *testing.T) {
		c, teardown := zillizClient[Clusters](t)
		defer teardown()

		resp, err := c.CreateFreeCluster(params)
		if err != nil {
			t.Fatalf("failed to create cluster: %v", err)
		}

		if resp.ClusterId == "" {
			t.Fatalf("failed to create cluster: %v", resp)
		}

		clusterId = resp.ClusterId
	})

	t.Run("duplicately create cluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()

		_, err := c.CreateFreeCluster(params)
		var e = Error{
			Code: 40013,
		}

		if !errors.Is(err, e) {
			t.Fatalf("want = %v, but got = %v", e, err)
		}

	})

	t.Run("DescribeCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		// checkfn:=make([]func(resp *Cluster) bool,0)
		checkfn := []func(resp *Cluster) bool{
			checkClusterId(clusterId),
		}

		ctx, cancelfn := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancelfn()
		got := pollClusterStatus(ctx, t, c, clusterId, "RUNNING", pollInterval)

		for _, fn := range checkfn {
			if !fn(got) {
				t.Errorf("check failed")
			}
		}

	})

	t.Run("DeleteCluster", func(t *testing.T) {

		c, teardown := zillizClient[Clusters](t)
		defer teardown()
		got, err := c.DropCluster(clusterId)
		if err != nil {
			t.Fatalf("failed to delete cluster: %v", err)
		}

		if got == nil || *got != clusterId {
			t.Fatalf("want = %s, got = %v", clusterId, got)
		}
	})
}

// nolint:unparam
func pollClusterStatus(ctx context.Context, t *testing.T, c *Client, clusterId string, status string, pollingInterval int) *Cluster {

	var (
		got Cluster
		err error
	)
	interval := time.Duration(pollingInterval) * time.Second
	// ctx, _ := context.WithTimeout(ctx, 5*time.Minute)

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout")
			return nil

		case <-time.After(interval):
			t.Logf("[%s] polling cluster status...", time.Now().Format("2006-01-02 15:04:05"))

			got, err = c.DescribeCluster(clusterId)
			if err != nil {
				t.Fatalf("failed to describe cluster: %v", err)
			}

			switch got.Status {
			case status:
				return &got
			default:
				continue
			}
		}

	}
}

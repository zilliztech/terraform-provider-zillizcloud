package global_cluster

import (
	"testing"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestGlobalClusterFromAPIMapsCompleteDescribeResponse(t *testing.T) {
	got := GlobalClusterFromAPI(&zilliz.GlobalCluster{
		GlobalClusterId:   "glo-1",
		GlobalClusterName: "global-a",
		ProjectId:         "proj-1",
		RegionIds:         []string{"aws-us-west-2", "aws-eu-west-1"},
		CuType:            "Performance-optimized",
		CuSize:            4,
		ConnectAddress:    "https://glo-1.global-cluster.vectordb.zillizcloud.com",
		CreateTime:        "2026-06-04T10:00:00Z",
		Clusters: []zilliz.GlobalClusterMember{
			{ClusterId: "in01-primary", ClusterName: "primary-a", RegionId: "aws-us-west-2", Role: "PRIMARY", Status: "RUNNING"},
			{ClusterId: "in01-secondary", ClusterName: "secondary-eu", RegionId: "aws-eu-west-1", Role: "SECONDARY", Status: "RUNNING"},
		},
	})

	if got.GlobalClusterID != "glo-1" || got.GlobalClusterName != "global-a" || got.ProjectID != "proj-1" {
		t.Fatalf("unexpected identity fields: %+v", got)
	}
	if got.CUType != "Performance-optimized" || got.CUSize != 4 {
		t.Fatalf("unexpected CU fields: %+v", got)
	}
	if got.ConnectAddress == "" || got.CreateTime != "2026-06-04T10:00:00Z" {
		t.Fatalf("unexpected endpoint/time: %+v", got)
	}
	if len(got.RegionIDs) != 2 || got.RegionIDs[0] != "aws-us-west-2" || got.RegionIDs[1] != "aws-eu-west-1" {
		t.Fatalf("unexpected region ids: %+v", got.RegionIDs)
	}
	if len(got.Clusters) != 2 {
		t.Fatalf("unexpected members: %+v", got.Clusters)
	}
	if got.Clusters[0].ClusterID != "in01-primary" || got.Clusters[0].Role != GlobalClusterMemberRolePrimary || got.Clusters[0].Status != "RUNNING" {
		t.Fatalf("unexpected primary: %+v", got.Clusters[0])
	}
	if got.Clusters[1].ClusterID != "in01-secondary" || got.Clusters[1].Role != GlobalClusterMemberRoleSecondary || got.Clusters[1].RegionID != "aws-eu-west-1" {
		t.Fatalf("unexpected secondary: %+v", got.Clusters[1])
	}
}

func TestGlobalClusterFromAPINil(t *testing.T) {
	if got := GlobalClusterFromAPI(nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestGlobalClusterSecondaryClusterDeletedCondition(t *testing.T) {
	cluster := &GlobalCluster{Clusters: []GlobalClusterMember{
		{ClusterID: "in01-primary", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
		{ClusterID: "in01-secondary", Role: GlobalClusterMemberRoleSecondary, Status: "DELETING"},
	}}

	done, lastStatus, err := cluster.isInstanceNotExists("in01-secondary")
	if done || lastStatus != "DELETING" || err != nil {
		t.Fatalf("expected deleting secondary to keep waiting, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	done, _, err = cluster.isInstanceNotExists("in01-missing")
	if !done || err != nil {
		t.Fatalf("expected missing secondary to be deleted, got done=%v err=%v", done, err)
	}
}

func TestGlobalClusterSecondaryClusterRunningCondition(t *testing.T) {
	cluster := &GlobalCluster{Clusters: []GlobalClusterMember{
		{ClusterID: "in01-primary", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
		{ClusterID: "in01-secondary", Role: GlobalClusterMemberRoleSecondary, Status: "CREATING"},
		{ClusterID: "in01-secondary-running", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
	}}

	done, lastStatus, err := cluster.isInstanceRunning("in01-secondary")
	if done || lastStatus != "CREATING" || err != nil {
		t.Fatalf("expected creating secondary to keep waiting, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	done, lastStatus, err = cluster.isInstanceRunning("in01-secondary-running")
	if !done || lastStatus != "RUNNING" || err != nil {
		t.Fatalf("expected running secondary to be done, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	done, lastStatus, err = cluster.isInstanceRunning("in01-missing")
	if done || lastStatus != "missing" || err != nil {
		t.Fatalf("expected missing secondary to keep waiting, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	_, _, err = cluster.isInstanceRunning("in01-primary")
	if err == nil {
		t.Fatalf("expected primary member to be rejected")
	}
}

func TestGlobalClusterSecondaryMemberRunningCondition(t *testing.T) {
	cluster := &GlobalCluster{Clusters: []GlobalClusterMember{
		{ClusterID: "in01-primary", ClusterName: "primary-a", RegionID: "aws-us-west-2", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
		{ClusterID: "in01-secondary", ClusterName: "secondary-eu", RegionID: "aws-eu-west-1", Role: GlobalClusterMemberRoleSecondary, Status: "CREATING"},
		{ClusterID: "in01-secondary-running", ClusterName: "secondary-ap", RegionID: "aws-ap-southeast-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
	}}

	done, lastStatus, err := cluster.isSecondaryMemberRunning(GlobalClusterMemberSpec{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1"})
	if done || lastStatus != "CREATING" || err != nil {
		t.Fatalf("expected creating secondary to keep waiting, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	done, lastStatus, err = cluster.isSecondaryMemberRunning(GlobalClusterMemberSpec{ClusterName: "secondary-ap", RegionID: "aws-ap-southeast-1"})
	if !done || lastStatus != "RUNNING" || err != nil {
		t.Fatalf("expected running secondary to be done, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	done, lastStatus, err = cluster.isSecondaryMemberRunning(GlobalClusterMemberSpec{ClusterName: "secondary-missing", RegionID: "aws-ap-southeast-2"})
	if done || lastStatus != "missing" || err != nil {
		t.Fatalf("expected missing secondary to keep waiting, got done=%v lastStatus=%s err=%v", done, lastStatus, err)
	}

	_, _, err = cluster.isSecondaryMemberRunning(GlobalClusterMemberSpec{ClusterName: "primary-a", RegionID: "aws-us-west-2"})
	if err == nil {
		t.Fatalf("expected primary member to be rejected")
	}
}

package global_cluster

import (
	"strings"
	"testing"
)

func testCurrentGlobalCluster() *GlobalCluster {
	return &GlobalCluster{
		GlobalClusterID: "glo-1",
		Clusters: []GlobalClusterMember{
			{ClusterID: "in01-primary", ClusterName: "primary-a", RegionID: "aws-us-west-2", Role: GlobalClusterMemberRolePrimary, Status: "RUNNING"},
			{ClusterID: "in01-secondary", ClusterName: "secondary-eu", RegionID: "aws-eu-west-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
			{ClusterID: "in01-secondary-ap", ClusterName: "secondary-ap", RegionID: "aws-ap-southeast-1", Role: GlobalClusterMemberRoleSecondary, Status: "RUNNING"},
		},
	}
}

func TestPlanSecondaryClusterChangeAddsSecondaries(t *testing.T) {
	plan, err := testCurrentGlobalCluster().PlanSecondaryClusterChange([]GlobalClusterMemberSpec{
		{ClusterName: "primary-a", RegionID: "aws-us-west-2"},
		{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1"},
		{ClusterName: "secondary-ap", RegionID: "aws-ap-southeast-1"},
		{ClusterName: "secondary-au", RegionID: "aws-ap-southeast-2"},
	})
	if err != nil {
		t.Fatalf("PlanSecondaryClusterChange err: %v", err)
	}
	if len(plan.Delete) != 0 {
		t.Fatalf("unexpected deletes: %+v", plan.Delete)
	}
	if len(plan.Add) != 1 || plan.Add[0].ClusterName != "secondary-au" || plan.Add[0].RegionID != "aws-ap-southeast-2" {
		t.Fatalf("unexpected adds: %+v", plan.Add)
	}
}

func TestPlanSecondaryClusterChangeDeletesSecondariesByClusterID(t *testing.T) {
	plan, err := testCurrentGlobalCluster().PlanSecondaryClusterChange([]GlobalClusterMemberSpec{
		{ClusterName: "primary-a", RegionID: "aws-us-west-2"},
		{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1"},
	})
	if err != nil {
		t.Fatalf("PlanSecondaryClusterChange err: %v", err)
	}
	if len(plan.Add) != 0 {
		t.Fatalf("unexpected adds: %+v", plan.Add)
	}
	if len(plan.Delete) != 1 || plan.Delete[0].ClusterID != "in01-secondary-ap" {
		t.Fatalf("unexpected deletes: %+v", plan.Delete)
	}
}

func TestPlanSecondaryClusterChangeRejectsEmptyDesiredMembers(t *testing.T) {
	_, err := testCurrentGlobalCluster().PlanSecondaryClusterChange(nil)
	if err == nil || !strings.Contains(err.Error(), "cannot express empty global cluster members") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPlanSecondaryClusterChangeRejectsPrimaryChange(t *testing.T) {
	_, err := testCurrentGlobalCluster().PlanSecondaryClusterChange([]GlobalClusterMemberSpec{
		{ClusterName: "primary-renamed", RegionID: "aws-us-west-2"},
		{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1"},
	})
	if err == nil || !strings.Contains(err.Error(), "cannot express primary member changes") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestPlanSecondaryClusterChangeRejectsAddAndDeleteTogether(t *testing.T) {
	_, err := testCurrentGlobalCluster().PlanSecondaryClusterChange([]GlobalClusterMemberSpec{
		{ClusterName: "primary-a", RegionID: "aws-us-west-2"},
		{ClusterName: "secondary-eu", RegionID: "aws-eu-west-1"},
		{ClusterName: "secondary-au", RegionID: "aws-ap-southeast-2"},
	})
	if err == nil || !strings.Contains(err.Error(), "cannot express secondary replacement") {
		t.Fatalf("unexpected err: %v", err)
	}
}

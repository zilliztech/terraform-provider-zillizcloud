package cluster

import (
	"context"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestOnDemandClusterMetadata(t *testing.T) {
	ctx := context.Background()
	res := NewOnDemandClusterResource()
	var resp fwresource.MetadataResponse
	res.Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "zillizcloud"}, &resp)
	if resp.TypeName != "zillizcloud_on_demand_cluster" {
		t.Fatalf("TypeName=%q", resp.TypeName)
	}
}

func TestAutoSuspendValidator(t *testing.T) {
	tests := []struct {
		name    string
		value   int64
		wantErr bool
	}{
		{name: "minimum", value: 60},
		{name: "thirty minutes", value: 1800},
		{name: "below minimum", value: 59, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp validator.Int64Response
			autoSuspendValidator{}.ValidateInt64(context.Background(), validator.Int64Request{
				ConfigValue: types.Int64Value(tt.value),
			}, &resp)
			if resp.Diagnostics.HasError() != tt.wantErr {
				t.Fatalf("autoSuspendValidator(%d) diagnostics=%v wantErr=%v", tt.value, resp.Diagnostics, tt.wantErr)
			}
		})
	}
}

func TestOnDemandClusterCreateRequest(t *testing.T) {
	model := &OnDemandClusterResourceModel{
		ProjectID:            types.StringValue("proj-1"),
		RegionID:             types.StringValue("aws-us-west-2"),
		ClusterName:          types.StringValue("query-dev"),
		CUSize:               types.Int64Value(8),
		AutoSuspend:          types.Int64Value(1800),
		MaxQueryNodeCU:       types.Int64Value(8),
		MaxQueryNodeReplicas: types.Int64Value(4),
	}

	req := onDemandClusterCreateRequest(model)
	if req.ProjectID != "proj-1" || req.RegionID != "aws-us-west-2" || req.ClusterName != "query-dev" || req.CUSize != 8 {
		t.Fatalf("req=%+v", req)
	}
	if req.AutoSuspend == nil || *req.AutoSuspend != 1800 {
		t.Fatalf("AutoSuspend=%v", req.AutoSuspend)
	}
	if req.MaxQueryNodeCU == nil || *req.MaxQueryNodeCU != 8 {
		t.Fatalf("MaxQueryNodeCU=%v", req.MaxQueryNodeCU)
	}
	if req.MaxQueryNodeReplicas == nil || *req.MaxQueryNodeReplicas != 4 {
		t.Fatalf("MaxQueryNodeReplicas=%v", req.MaxQueryNodeReplicas)
	}
}

func TestOnDemandClusterFromAPI(t *testing.T) {
	name := "query-dev"
	cuSize := 8
	replicas := 3
	readyReplicas := 2
	endpoint := "https://query.example.com"
	privateLink := "https://private.example.com"
	createdBy := "user@example.com"
	createTime := int64(1777867000000)
	autoSuspend := 1800
	ttlSeconds := 1800

	model := onDemandClusterFromAPI(&zilliz.QueryCluster{
		ClusterID:     "in07-qc-1",
		ClusterName:   &name,
		RegionID:      "aws-us-west-2",
		CUSize:        &cuSize,
		Replicas:      &replicas,
		ReadyReplicas: &readyReplicas,
		Status:        "RUNNING",
		Endpoint:      &endpoint,
		PrivateLink:   &privateLink,
		CreatedBy:     &createdBy,
		CreateTime:    &createTime,
		AutoSuspend:   &autoSuspend,
		TTLSeconds:    &ttlSeconds,
	})

	if model.ID.ValueString() != "in07-qc-1" ||
		model.ClusterName.ValueString() != name ||
		model.RegionID.ValueString() != "aws-us-west-2" ||
		model.CUSize.ValueInt64() != int64(cuSize) ||
		model.Replicas.ValueInt64() != int64(replicas) ||
		model.ReadyReplicas.ValueInt64() != int64(readyReplicas) ||
		model.Status.ValueString() != "RUNNING" ||
		model.Endpoint.ValueString() != endpoint ||
		model.PrivateLink.ValueString() != privateLink ||
		model.CreatedBy.ValueString() != createdBy ||
		model.CreateTime.ValueInt64() != createTime ||
		model.AutoSuspend.ValueInt64() != int64(autoSuspend) ||
		model.TTLSeconds.ValueInt64() != int64(ttlSeconds) {
		t.Fatalf("model=%+v", model)
	}
}

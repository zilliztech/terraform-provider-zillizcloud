package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type OnDemandClusterStore interface {
	Create(ctx context.Context, cluster *OnDemandClusterResourceModel) (*OnDemandClusterResourceModel, error)
	Get(ctx context.Context, clusterID string) (*OnDemandClusterResourceModel, error)
	Delete(ctx context.Context, clusterID string) (*OnDemandClusterResourceModel, error)
}

var _ OnDemandClusterStore = (*OnDemandClusterStoreImpl)(nil)

type OnDemandClusterStoreImpl struct {
	client *zilliz.Client
}

func (s *OnDemandClusterStoreImpl) Create(ctx context.Context, cluster *OnDemandClusterResourceModel) (*OnDemandClusterResourceModel, error) {
	req := onDemandClusterCreateRequest(cluster)
	response, err := s.client.CreateOnDemandCluster(req)
	if err != nil {
		return nil, err
	}

	return &OnDemandClusterResourceModel{
		ID:     types.StringValue(response.ClusterId),
		Prompt: types.StringValue(response.Prompt),
	}, nil
}

func (s *OnDemandClusterStoreImpl) Get(ctx context.Context, clusterID string) (*OnDemandClusterResourceModel, error) {
	cluster, err := s.client.DescribeOnDemandCluster(clusterID)
	if err != nil {
		return nil, err
	}

	return onDemandClusterFromAPI(cluster), nil
}

func (s *OnDemandClusterStoreImpl) Delete(ctx context.Context, clusterID string) (*OnDemandClusterResourceModel, error) {
	response, err := s.client.DeleteOnDemandCluster(clusterID)
	if err != nil {
		return nil, err
	}

	return &OnDemandClusterResourceModel{
		ID:     types.StringValue(response.ClusterID),
		Prompt: types.StringValue(response.Prompt),
	}, nil
}

func onDemandClusterCreateRequest(cluster *OnDemandClusterResourceModel) *zilliz.CreateOnDemandClusterRequest {
	req := &zilliz.CreateOnDemandClusterRequest{
		ProjectID:   cluster.ProjectID.ValueString(),
		RegionID:    cluster.RegionID.ValueString(),
		CUSize:      int(cluster.CUSize.ValueInt64()),
		ClusterName: cluster.ClusterName.ValueString(),
	}

	if !cluster.AutoSuspend.IsNull() && !cluster.AutoSuspend.IsUnknown() {
		v := int(cluster.AutoSuspend.ValueInt64())
		req.AutoSuspend = &v
	}
	if !cluster.MaxQueryNodeCU.IsNull() && !cluster.MaxQueryNodeCU.IsUnknown() {
		v := int(cluster.MaxQueryNodeCU.ValueInt64())
		req.MaxQueryNodeCU = &v
	}
	if !cluster.MaxQueryNodeReplicas.IsNull() && !cluster.MaxQueryNodeReplicas.IsUnknown() {
		v := int(cluster.MaxQueryNodeReplicas.ValueInt64())
		req.MaxQueryNodeReplicas = &v
	}

	return req
}

func onDemandClusterFromAPI(cluster *zilliz.QueryCluster) *OnDemandClusterResourceModel {
	model := &OnDemandClusterResourceModel{
		ID:            types.StringValue(cluster.ClusterID),
		ClusterName:   types.StringNull(),
		RegionID:      types.StringValue(cluster.RegionID),
		CUSize:        types.Int64Null(),
		Replicas:      types.Int64Null(),
		ReadyReplicas: types.Int64Null(),
		Status:        types.StringValue(cluster.Status),
		Endpoint:      stringPtrToTerraform(cluster.Endpoint),
		PrivateLink:   stringPtrToTerraform(cluster.PrivateLink),
		CreatedBy:     stringPtrToTerraform(cluster.CreatedBy),
		CreateTime:    int64PtrToTerraform(cluster.CreateTime),
		AutoSuspend:   types.Int64Null(),
		TTLSeconds:    types.Int64Null(),
	}

	if cluster.ClusterName != nil {
		model.ClusterName = types.StringValue(*cluster.ClusterName)
	}
	if cluster.CUSize != nil {
		model.CUSize = types.Int64Value(int64(*cluster.CUSize))
	}
	if cluster.Replicas != nil {
		model.Replicas = types.Int64Value(int64(*cluster.Replicas))
	}
	if cluster.ReadyReplicas != nil {
		model.ReadyReplicas = types.Int64Value(int64(*cluster.ReadyReplicas))
	}
	if cluster.AutoSuspend != nil {
		model.AutoSuspend = types.Int64Value(int64(*cluster.AutoSuspend))
	}
	if cluster.TTLSeconds != nil {
		model.TTLSeconds = types.Int64Value(int64(*cluster.TTLSeconds))
	}

	return model
}

func stringPtrToTerraform(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func int64PtrToTerraform(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

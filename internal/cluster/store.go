package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type ClusterStore interface {
	Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error)
	Create(ctx context.Context, cluster *ClusterResourceModel) (*ClusterResourceModel, error)
	Delete(ctx context.Context, clusterId string) error
	UpgradeCuSize(ctx context.Context, clusterId string, cuSize int) error
	ModifyReplica(ctx context.Context, clusterId string, replica int) error
	SuspendCluster(ctx context.Context, clusterId string) error
	ResumeCluster(ctx context.Context, clusterId string) error
}

type ClusterStoreImpl struct {
	client *zilliz.Client
}

func (c *ClusterStoreImpl) Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error) {
	cluster, err := c.client.DescribeCluster(clusterId)
	if err != nil {
		return nil, err
	}

	labels := types.MapNull(types.StringType)
	if len(cluster.Labels) > 0 {
		labelValues := make(map[string]attr.Value)
		for k, v := range cluster.Labels {
			labelValues[k] = types.StringValue(v)
		}
		labels, _ = types.MapValue(types.StringType, labelValues)
	}

	return &ClusterResourceModel{
		ClusterId:   types.StringValue(cluster.ClusterId),
		Plan:        types.StringValue(string(cluster.Plan)),
		ClusterName: types.StringValue(cluster.ClusterName),
		CuSize:      types.Int64Value(cluster.CuSize),
		CuType:      types.StringValue(cluster.CuType),
		ProjectId:   types.StringValue(cluster.ProjectId),
		// Username:           types.StringValue(cluster.Username),
		// Password:           types.StringValue(cluster.Password),
		// Prompt:             types.StringValue(cluster.Prompt),
		Description: types.StringValue(cluster.Description),
		RegionId:    types.StringValue(cluster.RegionId),
		Status: types.StringValue(
			func() string {
				switch cluster.Status {
				case "STOPPING":
					return "SUSPENDING"
				case "STOPPED":
					return "SUSPENDED"
				default:
					return cluster.Status
				}
			}(),
		),
		ConnectAddress:     types.StringValue(cluster.ConnectAddress),
		PrivateLinkAddress: types.StringValue(cluster.PrivateLinkAddress),
		Labels:             labels,
		Replica:            types.Int64Value(cluster.Replica),
	}, nil
}

func (c *ClusterStoreImpl) Create(ctx context.Context, cluster *ClusterResourceModel) (ret *ClusterResourceModel, err error) {
	var response *zilliz.CreateClusterResponse

	regionId := cluster.RegionId.ValueString()
	if cluster.RegionId.IsNull() || cluster.RegionId.ValueString() == "" {
		regionId = c.client.RegionId
	}

	// Convert terraform map to Go map for labels
	labels := make(map[string]string)
	if !cluster.Labels.IsNull() && !cluster.Labels.IsUnknown() {
		elements := cluster.Labels.Elements()
		for k, v := range elements {
			if strValue, ok := v.(types.String); ok {
				labels[k] = strValue.ValueString()
			}
		}
	}

	if zilliz.Plan(cluster.Plan.ValueString()) == zilliz.FreePlan {
		response, err = c.client.CreateFreeCluster(zilliz.CreateServerlessClusterParams{
			RegionId:    regionId,
			ClusterName: cluster.ClusterName.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
		})
	} else if zilliz.Plan(cluster.Plan.ValueString()) == zilliz.ServerlessPlan {
		response, err = c.client.CreateServerlessCluster(zilliz.CreateServerlessClusterParams{
			RegionId:    regionId,
			ClusterName: cluster.ClusterName.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
		})
	} else if zilliz.Plan(cluster.Plan.ValueString()) == zilliz.StandardPlan || zilliz.Plan(cluster.Plan.ValueString()) == zilliz.EnterprisePlan {
		response, err = c.client.CreateDedicatedCluster(zilliz.CreateClusterParams{
			RegionId:    regionId,
			Plan:        zilliz.Plan(cluster.Plan.ValueString()),
			ClusterName: cluster.ClusterName.ValueString(),
			CUSize:      int(cluster.CuSize.ValueInt64()),
			CUType:      cluster.CuType.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
			Labels:      labels,
		})
	} else {
		// byoc env if plan is not set
		response, err = c.client.CreateDedicatedCluster(zilliz.CreateClusterParams{
			ClusterName: cluster.ClusterName.ValueString(),
			CUSize:      int(cluster.CuSize.ValueInt64()),
			CUType:      cluster.CuType.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
			Labels:      labels,
		})
	}

	if err != nil {
		return nil, err
	}

	ret = &ClusterResourceModel{
		ClusterId: types.StringValue(response.ClusterId),
		Username:  types.StringValue(response.Username),
		Password:  types.StringValue(response.Password),
		Prompt:    types.StringValue(response.Prompt),
	}
	return ret, nil
}

func (c *ClusterStoreImpl) Delete(ctx context.Context, clusterId string) error {
	_, err := c.client.DropCluster(clusterId)
	return err
}

func (c *ClusterStoreImpl) UpgradeCuSize(ctx context.Context, clusterId string, cuSize int) error {
	_, err := c.client.ModifyCluster(clusterId, &zilliz.ModifyClusterParams{
		CuSize: cuSize,
	})
	return err
}

func (c *ClusterStoreImpl) ModifyReplica(ctx context.Context, clusterId string, replica int) error {
	_, err := c.client.ModifyReplica(clusterId, &zilliz.ModifyReplicaParams{
		Replica: replica,
	})
	return err
}

func (c *ClusterStoreImpl) SuspendCluster(ctx context.Context, clusterId string) error {
	_, err := c.client.SuspendCluster(clusterId)
	return err
}

func (c *ClusterStoreImpl) ResumeCluster(ctx context.Context, clusterId string) error {
	_, err := c.client.ResumeCluster(clusterId)
	return err
}

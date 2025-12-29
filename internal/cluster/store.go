package cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	"github.com/zilliztech/terraform-provider-zillizcloud/internal/util/conv"
)

type ClusterStore interface {
	Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error)
	GetLabels(ctx context.Context, clusterId string) (types.Map, error)
	Create(ctx context.Context, cluster *ClusterResourceModel) (*ClusterResourceModel, error)
	Delete(ctx context.Context, clusterId string) error
	UpgradeCuSize(ctx context.Context, clusterId string, cuSize int) error
	ModifyReplica(ctx context.Context, clusterId string, replica int) error
	SuspendCluster(ctx context.Context, clusterId string) error
	ResumeCluster(ctx context.Context, clusterId string) error
	UpdateLabels(ctx context.Context, clusterId string, labels map[string]string) error
	ModifyClusterProperties(ctx context.Context, clusterId string, clusterName string) error
	UpsertSecurityGroups(ctx context.Context, clusterId string, securityGroupIds []string) error
	GetSecurityGroups(ctx context.Context, clusterId string) ([]string, error)
	ModifyAutoscaling(ctx context.Context, clusterId string, minCU int, maxCU int) error
	ModifySchedules(ctx context.Context, clusterId string, schedules []zilliz.ScheduleConfig) error
	ModifyReplicaAutoscaling(ctx context.Context, clusterId string, minCU int, maxCU int) error
	ModifyReplicaSchedules(ctx context.Context, clusterId string, schedules []zilliz.ScheduleConfig) error
}

var _ ClusterStore = (*ClusterStoreImpl)(nil)

type ClusterStoreImpl struct {
	client *zilliz.Client
}

func (c *ClusterStoreImpl) Get(ctx context.Context, clusterId string) (*ClusterResourceModel, error) {
	cluster, err := c.client.DescribeCluster(clusterId)
	if err != nil {
		return nil, err
	}

	var dynamicScaling *DynamicScaling
	if cluster.Autoscaling.CU.Min != nil && cluster.Autoscaling.CU.Max != nil {
		dynamicScaling = &DynamicScaling{
			Min: types.Int64Value(int64(*cluster.Autoscaling.CU.Min)),
			Max: types.Int64Value(int64(*cluster.Autoscaling.CU.Max)),
		}
	}

	var schedules []ScheduleScaling
	if len(cluster.Autoscaling.CU.Schedules) > 0 {
		schedules = make([]ScheduleScaling, len(cluster.Autoscaling.CU.Schedules))
		for i, s := range cluster.Autoscaling.CU.Schedules {
			schedules[i] = ScheduleScaling{
				Cron:   types.StringValue(s.Cron),
				Target: types.Int64Value(int64(s.Target)),
			}
		}
	}

	return &ClusterResourceModel{
		ClusterId:   types.StringValue(cluster.ClusterId),
		Plan:        types.StringValue(cluster.Plan),
		ClusterName: types.StringValue(cluster.ClusterName),
		CuSize:      types.Int64Value(cluster.CuSize),
		CuType:      types.StringValue(cluster.CuType),
		ProjectId:   types.StringValue(cluster.ProjectId),
		// state fixed by input
		// Username:           types.StringValue(cluster.Username),
		// Password:           types.StringValue(cluster.Password),
		// Prompt:             types.StringValue(cluster.Prompt),
		CreateTime:  types.StringValue(cluster.CreateTime),
		Labels:      types.MapUnknown(types.StringType), // read by GetLabels api
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
		Replica: types.Int64Value(func() int64 {
			if cluster.Replica == 0 {
				return 1
			}
			return cluster.Replica
		}()),
		CuSettings: &CuSettings{
			DynamicScaling:  dynamicScaling,
			ScheduleScaling: schedules,
		},
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
	zillizPlan := cluster.Plan.ValueString()
	switch zillizPlan {
	case FreePlan:
		response, err = c.client.CreateFreeCluster(zilliz.CreateServerlessClusterParams{
			RegionId:    regionId,
			ClusterName: cluster.ClusterName.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
		})
	case ServerlessPlan:
		response, err = c.client.CreateServerlessCluster(zilliz.CreateServerlessClusterParams{
			RegionId:    regionId,
			ClusterName: cluster.ClusterName.ValueString(),
			ProjectId:   cluster.ProjectId.ValueString(),
		})
	default:

		// only for the byoc case
		var bucketInfo *zilliz.BucketInfo
		if cluster.BucketInfo != nil {
			bucketInfo = &zilliz.BucketInfo{
				BucketName: cluster.BucketInfo.BucketName.ValueString(),
				Prefix:     cluster.BucketInfo.Prefix.ValueStringPointer(),
			}
		}

		// dedicated:
		response, err = c.client.CreateDedicatedCluster(zilliz.CreateClusterParams{
			RegionId: regionId,
			Plan: func() *string {
				if cluster.Plan.IsNull() || cluster.Plan.IsUnknown() {
					return nil
				}
				return conv.StringPtr(cluster.Plan.ValueString())
			}(),
			ClusterName: cluster.ClusterName.ValueString(),
			CUSize: func() int {
				if cluster.CuSize.IsNull() || cluster.CuSize.IsUnknown() {
					return 1
				}
				return int(cluster.CuSize.ValueInt64())
			}(),
			CUType:     cluster.CuType.ValueString(),
			ProjectId:  cluster.ProjectId.ValueString(),
			Labels:     labels,
			BucketInfo: bucketInfo,
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

func (c *ClusterStoreImpl) UpdateLabels(ctx context.Context, clusterId string, labels map[string]string) error {
	_, err := c.client.UpdateLabels(clusterId, &zilliz.UpdateLabelsParams{
		Labels: labels,
	})
	return err
}

func (c *ClusterStoreImpl) GetLabels(ctx context.Context, clusterId string) (types.Map, error) {
	labels, err := c.client.GetLabels(clusterId)
	if err != nil {
		return types.MapValueMust(types.StringType, map[string]attr.Value{}), err
	}
	return convertLabelsToTypesMap(labels), nil
}

func (c *ClusterStoreImpl) ModifyClusterProperties(ctx context.Context, clusterId string, clusterName string) error {
	_, err := c.client.ModifyClusterProperties(clusterId, &zilliz.ModifyPropertiesParams{
		ClusterName: clusterName,
	})
	return err
}

func (c *ClusterStoreImpl) UpsertSecurityGroups(ctx context.Context, clusterId string, securityGroupIds []string) error {
	_, err := c.client.UpsertSecurityGroups(clusterId, &zilliz.UpsertSecurityGroupsParams{
		Ids: securityGroupIds,
	})
	return err
}

func (c *ClusterStoreImpl) GetSecurityGroups(ctx context.Context, clusterId string) ([]string, error) {
	return c.client.GetSecurityGroups(clusterId)
}

func (c *ClusterStoreImpl) ModifyAutoscaling(ctx context.Context, clusterId string, minCU int, maxCU int) error {
	ptrInt := func(i int) *int {
		return &i
	}
	params := &zilliz.ModifyClusterAutoscalingParams{}
	params.Autoscaling.CU.Min = ptrInt(minCU)
	params.Autoscaling.CU.Max = ptrInt(maxCU)
	_, err := c.client.ModifyClusterAutoscaling(clusterId, params)
	return err
}

func (c *ClusterStoreImpl) ModifySchedules(ctx context.Context, clusterId string, schedules []zilliz.ScheduleConfig) error {
	params := &zilliz.ModifyClusterAutoscalingParams{}
	params.Autoscaling.CU.Schedules = &schedules
	_, err := c.client.ModifyClusterAutoscaling(clusterId, params)
	return err
}

func (c *ClusterStoreImpl) ModifyReplicaAutoscaling(ctx context.Context, clusterId string, minCU int, maxCU int) error {
	ptrInt := func(i int) *int {
		return &i
	}
	params := &zilliz.ModifyReplicaSettings{}
	params.Autoscaling.Replica.Min = ptrInt(minCU)
	params.Autoscaling.Replica.Max = ptrInt(maxCU)
	_, err := c.client.ModifyReplicaSettings(clusterId, params)
	return err
}

func (c *ClusterStoreImpl) ModifyReplicaSchedules(ctx context.Context, clusterId string, schedules []zilliz.ScheduleConfig) error {
	params := &zilliz.ModifyReplicaSettings{}
	params.Autoscaling.Replica.Schedules = &schedules
	_, err := c.client.ModifyReplicaSettings(clusterId, params)
	return err
}

// convertLabelsToTypesMap converts a map[string]string into a Terraform types.Map of strings.
// Returns an empty map value when the input is nil or empty.
func convertLabelsToTypesMap(src map[string]string) types.Map {
	if len(src) == 0 {
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}
	values := make(map[string]attr.Value, len(src))
	for k, v := range src {
		values[k] = types.StringValue(v)
	}
	ret, _ := types.MapValue(types.StringType, values)
	return ret
}

// convertStringSliceToTypesSet converts a []string into a Terraform types.Set of strings.
// Returns an empty set value when the input is nil or empty.
func convertStringSliceToTypesSet(src []string) types.Set {
	if len(src) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	values := make([]attr.Value, len(src))
	for i, v := range src {
		values[i] = types.StringValue(v)
	}
	ret, _ := types.SetValue(types.StringType, values)
	return ret
}

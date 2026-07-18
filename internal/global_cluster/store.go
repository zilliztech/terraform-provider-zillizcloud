package global_cluster

import (
	"context"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type GlobalClusterStore interface {
	Create(ctx context.Context, command CreateGlobalClusterCommand) (*CreateGlobalClusterResult, error)
	Describe(ctx context.Context, globalClusterID string) (*GlobalCluster, error)
	ModifyCU(ctx context.Context, globalClusterID string, cuSize int64) error
	AddSecondaryClusters(ctx context.Context, globalClusterID string, members []GlobalClusterMemberSpec) error
	DeleteCluster(ctx context.Context, globalClusterID string, clusterID string) error
}

type globalClusterStore struct {
	client *zilliz.Client
}

func NewGlobalClusterStore(client *zilliz.Client) GlobalClusterStore {
	return &globalClusterStore{client: client}
}

func (s *globalClusterStore) Create(ctx context.Context, command CreateGlobalClusterCommand) (*CreateGlobalClusterResult, error) {
	primary, secondaries := memberParamsForCreate(command.Members)
	created, err := s.client.CreateGlobalCluster(&zilliz.CreateGlobalClusterParams{
		GlobalClusterName: command.GlobalClusterName,
		ProjectId:         command.ProjectID,
		CuType:            command.CUType,
		CuSize:            int(command.CUSize),
		Autoscaling:       autoscalingParam(command.Autoscaling),
		PrimaryCluster:    primary,
		SecondaryClusters: secondaries,
	})
	if err != nil {
		return nil, err
	}
	return &CreateGlobalClusterResult{
		GlobalClusterID: created.GlobalClusterId,
		Username:        created.Username,
		Password:        created.Password,
		JobID:           created.JobId,
	}, nil
}

func (s *globalClusterStore) Describe(ctx context.Context, globalClusterID string) (*GlobalCluster, error) {
	globalCluster, err := s.client.DescribeGlobalCluster(globalClusterID)
	if err != nil {
		return nil, err
	}
	return GlobalClusterFromAPI(globalCluster), nil
}

func (s *globalClusterStore) ModifyCU(ctx context.Context, globalClusterID string, cuSize int64) error {
	_, err := s.client.ModifyGlobalClusterCU(globalClusterID, &zilliz.ModifyGlobalClusterCUParams{CuSize: int(cuSize)})
	return err
}

func (s *globalClusterStore) AddSecondaryClusters(ctx context.Context, globalClusterID string, members []GlobalClusterMemberSpec) error {
	_, err := s.client.AddSecondaryClusters(globalClusterID, &zilliz.AddSecondaryClustersParams{SecondaryClusters: memberParams(members)})
	return err
}

func (s *globalClusterStore) DeleteCluster(ctx context.Context, globalClusterID string, clusterID string) error {
	_, err := s.client.DeleteCluster(globalClusterID, clusterID)
	return err
}

func memberParamsForCreate(members []GlobalClusterMemberSpec) (zilliz.GlobalClusterMemberParams, []zilliz.GlobalClusterMemberParams) {
	if len(members) == 0 {
		return zilliz.GlobalClusterMemberParams{}, nil
	}
	return createMemberParam(members[0]), createMemberParams(members[1:])
}

func createMemberParams(members []GlobalClusterMemberSpec) []zilliz.GlobalClusterMemberParams {
	result := make([]zilliz.GlobalClusterMemberParams, 0, len(members))
	for _, member := range members {
		result = append(result, createMemberParam(member))
	}
	return result
}

func createMemberParam(member GlobalClusterMemberSpec) zilliz.GlobalClusterMemberParams {
	return zilliz.GlobalClusterMemberParams{
		ClusterName: member.ClusterName,
		RegionId:    member.RegionID,
		Replica:     intPointer(member.Replica),
	}
}

func memberParams(members []GlobalClusterMemberSpec) []zilliz.GlobalClusterMemberParams {
	result := make([]zilliz.GlobalClusterMemberParams, 0, len(members))
	for _, member := range members {
		result = append(result, memberParam(member))
	}
	return result
}

func memberParam(member GlobalClusterMemberSpec) zilliz.GlobalClusterMemberParams {
	return zilliz.GlobalClusterMemberParams{ClusterName: member.ClusterName, RegionId: member.RegionID}
}

func autoscalingParam(autoscaling GlobalClusterAutoscaling) *zilliz.AutoscalingConfig {
	if autoscaling.CU == nil && autoscaling.Replica == nil {
		return nil
	}
	return &zilliz.AutoscalingConfig{
		CU:      autoscalingPolicyParam(autoscaling.CU),
		Replica: autoscalingPolicyParam(autoscaling.Replica),
	}
}

func autoscalingPolicyParam(policy *GlobalClusterAutoscalingPolicy) *zilliz.AutoscalingPolicy {
	if policy == nil {
		return nil
	}
	return &zilliz.AutoscalingPolicy{
		Min: intPointer(policy.Min),
		Max: intPointer(policy.Max),
	}
}

func intPointer(value *int64) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

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
	return memberParam(members[0]), memberParams(members[1:])
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

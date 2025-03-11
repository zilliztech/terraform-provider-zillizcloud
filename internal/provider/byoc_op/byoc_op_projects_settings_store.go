package byoc_op

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type ByocOpProjectSettingsStore interface {
	Create(ctx context.Context, data *BYOCOpProjectSettingsResourceModel, updateStatusFunc func(project *BYOCOpProjectSettingsResourceModel) error) (err error)
	Delete(ctx context.Context, data *BYOCOpProjectSettingsResourceModel) (err error)
	Describe(ctx context.Context, projectID string, dataPlaneID string) (model BYOCOpProjectSettingsResourceModel, err error)
}

type byocOpProjectSettingsStore struct {
	client *zilliz.Client
}

var _ ByocOpProjectSettingsStore = &byocOpProjectSettingsStore{}

func (s *byocOpProjectSettingsStore) Create(ctx context.Context, data *BYOCOpProjectSettingsResourceModel, updateStatusFunc func(project *BYOCOpProjectSettingsResourceModel) error) (err error) {
	request := zilliz.CreateByocOpProjectSettingsRequest{
		ProjectName:   data.ProjectName.ValueString(),
		CloudId:       data.CloudProvider.ValueString(),
		RegionId:      data.Region.ValueString(),
		SearchVm:      data.Instances.SearchVM.ValueString(),
		FundamentalVm: data.Instances.FundamentalVM.ValueString(),
		CoreVm:        data.Instances.CoreVM.ValueString(),
		DeployType:    TERRAFORM_DEPLOY_TYPE,
	}

	tflog.Info(ctx, fmt.Sprintf("Create BYOC Op Project Settings request: %+v", request))

	response, err := s.client.CreateByocOpProjectSetting(&request)
	if err != nil {
		return fmt.Errorf("failed to create BYOC Op project settings: %w", err)
	}
	tflog.Info(ctx, fmt.Sprintf("Create BYOC Op Project Settings response: %+v", response))

	data.ID = types.StringValue(response.ProjectId)
	data.ProjectID = types.StringValue(response.ProjectId)
	data.DataPlaneID = types.StringValue(response.DataPlaneId)
	if err := updateStatusFunc(data); err != nil {
		return fmt.Errorf("failed to update status")
	}

	return nil
}

func (s *byocOpProjectSettingsStore) Delete(ctx context.Context, data *BYOCOpProjectSettingsResourceModel) (err error) {
	request := zilliz.DeleteByocOpProjectSettingRequest{
		ProjectId:   data.ProjectID.ValueString(),
		DataPlaneId: data.DataPlaneID.ValueString(),
	}

	response, err := s.client.DescribeByocOpProject(&zilliz.DescribeByocOpProjectRequest{
		ProjectId:   data.ProjectID.ValueString(),
		DataPlaneID: data.DataPlaneID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("failed to describe BYOC Op project settings: %w", err)
	}

	if response.Status == int(BYOCProjectStatusConnected) {
		return fmt.Errorf("BYOC Op project settings is not deleted when agent is connected")
	}

	// if the project is not connected, delete the project settings
	if response.Status == int(BYOCProjectStatusInit) {
		deleteResponse, err := s.client.DeleteByocOpProjectSetting(&request)
		if err != nil {
			return fmt.Errorf("failed to delete BYOC Op project settings: %w", err)
		}
		tflog.Info(ctx, fmt.Sprintf("Delete BYOC Op Project Settings response: %+v", deleteResponse))
	}

	return nil
}

func (s *byocOpProjectSettingsStore) Describe(ctx context.Context, projectID string, dataPlaneID string) (data BYOCOpProjectSettingsResourceModel, err error) {
	{
		response, err := s.client.DescribeByocOpProject(&zilliz.DescribeByocOpProjectRequest{
			ProjectId:   projectID,
			DataPlaneID: dataPlaneID,
		})
		if err != nil {
			return data, fmt.Errorf("failed to describe BYOC Op project settings: %w", err)
		}

		// Convert response to model
		data.ID = types.StringValue(response.ProjectID)
		data.DataPlaneID = types.StringValue(response.DataPlaneID)
		data.ProjectID = types.StringValue(response.ProjectID)
		data.CloudProvider = types.StringValue(response.CloudID)
		data.Region = types.StringValue(response.RegionID)

	}
	{
		response, err := s.client.DescribeByocOpProjectSettings(&zilliz.DescribeByocOpProjectSettingsRequest{
			ProjectId:   projectID,
			DataPlaneId: dataPlaneID,
		})
		if err != nil {
			return data, fmt.Errorf("failed to describe BYOC Op project settings: %w", err)
		}
		OpConfig, diag := types.ObjectValue(map[string]attr.Type{
			"token":           types.StringType,
			"agent_image_url": types.StringType,
		}, map[string]attr.Value{
			"token":           types.StringValue(response.OpConfig.Token),
			"agent_image_url": types.StringValue(response.OpConfig.AgentImageUrl),
		})
		if diag.HasError() {
			return data, fmt.Errorf("failed to abstract OpConfig from response")
		}
		data.OpConfig = OpConfig

		core, err := buildNodeQuotas("core", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		index, err := buildNodeQuotas("index", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		search, err := buildNodeQuotas("search", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		fundamental, err := buildNodeQuotas("fundamental", response.NodeQuotas)
		if err != nil {
			return data, err
		}

		NodeQuotas, diag := types.ObjectValue(nodeQuotasGenerateAttrTypes, map[string]attr.Value{
			"core":        core,
			"index":       index,
			"search":      search,
			"fundamental": fundamental,
		})
		if diag.HasError() {
			return data, fmt.Errorf("failed to abstract NodeQuotas from response")
		}
		data.NodeQuotas = NodeQuotas

	}
	return data, nil
}

func buildNodeQuotas(kind string, nodes []zilliz.NodeQuota) (types.Object, error) {
	var item *zilliz.NodeQuota

	filter := func(kind string, nodes []zilliz.NodeQuota) *zilliz.NodeQuota {
		for _, node := range nodes {
			if node.Name == kind {
				return &node
			}
		}
		return nil
	}
	if item = filter(kind, nodes); item == nil {
		return types.ObjectNull(map[string]attr.Type{
			"disk_size":      types.Int64Type,
			"min_size":       types.Int64Type,
			"max_size":       types.Int64Type,
			"desired_size":   types.Int64Type,
			"instance_types": types.StringType,
			"capacity_type":  types.StringType,
		}), nil
	}
	getInstanceTypes := func() types.String {
		if len(item.InstanceTypes) == 0 {
			return types.StringNull()
		}
		return types.StringValue(item.InstanceTypes[0])
	}

	ret, diag := types.ObjectValue(nodeQuotasAttrTypes, map[string]attr.Value{
		"disk_size":      types.Int64Value(int64(item.DiskSize)),
		"min_size":       types.Int64Value(int64(item.MinSize)),
		"max_size":       types.Int64Value(int64(item.MaxSize)),
		"desired_size":   types.Int64Value(int64(item.DesiredSize)),
		"instance_types": getInstanceTypes(),
		"capacity_type":  types.StringValue(item.CapacityType),
	})
	if diag.HasError() {
		var errMsg string
		for _, d := range diag {
			errMsg += fmt.Sprintf("detail: %s\n", d.Detail())
		}
		return types.ObjectNull(map[string]attr.Type{}), fmt.Errorf("failed to abstract NodeQuotas from response, detail: %s", errMsg)
	}
	return ret, nil
}

var nodeQuotasAttrTypes = map[string]attr.Type{
	"disk_size":      types.Int64Type,
	"min_size":       types.Int64Type,
	"max_size":       types.Int64Type,
	"desired_size":   types.Int64Type,
	"instance_types": types.StringType,
	"capacity_type":  types.StringType,
}

var nodeQuotasGenerateAttrTypes = map[string]attr.Type{
	"core":        types.ObjectType{AttrTypes: nodeQuotasAttrTypes},
	"index":       types.ObjectType{AttrTypes: nodeQuotasAttrTypes},
	"search":      types.ObjectType{AttrTypes: nodeQuotasAttrTypes},
	"fundamental": types.ObjectType{AttrTypes: nodeQuotasAttrTypes},
}

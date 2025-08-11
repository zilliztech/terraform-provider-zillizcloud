package byoc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	util "github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
)

type ByocProjectStore interface {
	Create(ctx context.Context, data *BYOCProjectResourceModel, updateStatusFunc func(project *BYOCProjectResourceModel) error) (err error)
	Delete(ctx context.Context, data *BYOCProjectResourceModel) (err error)
	Describe(ctx context.Context, projectID string, dataPlaneID string) (model BYOCProjectResourceModel, err error)
	Suspend(ctx context.Context, data *BYOCProjectResourceModel) (err error)
	Resume(ctx context.Context, data *BYOCProjectResourceModel) (err error)
}

type byocProjectStore struct {
	client *zilliz.Client
}

var _ ByocProjectStore = &byocProjectStore{}

func (s *byocProjectStore) Suspend(ctx context.Context, data *BYOCProjectResourceModel) (err error) {
	_, err = s.client.SuspendBYOCProject(&zilliz.SuspendBYOCProjectRequest{
		ProjectId:   data.ID.ValueString(),
		DataPlaneID: data.DataPlaneID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("failed to suspend BYOC project: %w", err)
	}
	timeout, diags := data.Timeouts.Update(ctx, defaultBYOCProjectUpdateTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get update timeout")
	}
	_, err = util.Poll[any](ctx, timeout, func() (*any, *util.Err) {

		project, err := s.Describe(ctx, data.ID.ValueString(), data.DataPlaneID.ValueString())
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC project status")}
		}

		status := project.Status.ValueString()
		switch status {
		case BYOCProjectStatusStopping.String():
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is stopping")}
		case BYOCProjectStatusStopped.String():
			// achieved the target status
			return nil, nil
		default:
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("BYOC project is in unknown state: %s", status)}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to create BYOC project: %w", err)
	}
	return
}

func (s *byocProjectStore) Resume(ctx context.Context, data *BYOCProjectResourceModel) (err error) {
	_, err = s.client.ResumeBYOCProject(&zilliz.ResumeBYOCProjectRequest{
		ProjectId:   data.ID.ValueString(),
		DataPlaneID: data.DataPlaneID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("failed to resume BYOC project: %w", err)
	}
	timeout, diags := data.Timeouts.Update(ctx, defaultBYOCProjectUpdateTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get update timeout")
	}
	_, err = util.Poll[any](ctx, timeout, func() (*any, *util.Err) {
		project, err := s.Describe(ctx, data.ID.ValueString(), data.DataPlaneID.ValueString())
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC project status")}
		}

		status := project.Status.ValueString()
		switch status {
		case BYOCProjectStatusResuming.String():
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is resuming")}
		case BYOCProjectStatusRunning.String():
			// achieved the target status
			return nil, nil
		default:
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("BYOC project is in unknown state: %s", status)}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to resume BYOC project: %w", err)
	}
	return
}

func (s *byocProjectStore) Describe(ctx context.Context, projectID string, dataPlaneID string) (data BYOCProjectResourceModel, _ error) {
	var err error

	project, err := s.client.DescribeBYOCProject(&zilliz.DescribeBYOCProjectRequest{
		ProjectId:   projectID,
		DataPlaneID: dataPlaneID,
	})
	if err != nil {
		return data, fmt.Errorf("failed to describe BYOC project: %w", err)
	}

	data.Status = types.StringValue(BYOCProjectStatus(project.Status).String())
	data.ID = types.StringValue(project.ProjectID)
	// data.Name = types.StringValue(project.ProjectName)
	data.DataPlaneID = types.StringValue(project.DataPlaneID)
	subnetIDs, diags := sliceToTerraformSet(project.AWSConfig.SubnetIDs)

	if diags.HasError() {
		return data, fmt.Errorf("failed to convert subnet IDs to Terraform set: %+v", project.AWSConfig.SubnetIDs)
	}
	securityGroupIDs, diags := sliceToTerraformSet(project.AWSConfig.SecurityGroupIDs)
	if diags.HasError() {
		return data, fmt.Errorf("failed to convert security group IDs to Terraform set: %+v", project.AWSConfig.SecurityGroupIDs)
	}

	data.AWS = &AWSConfig{
		Region: types.StringValue(project.RegionID),
		Network: NetworkConfig{
			VPCID:            types.StringValue(project.AWSConfig.VPCID),
			SubnetIDs:        subnetIDs,
			SecurityGroupIDs: securityGroupIDs,
			VPCEndpointID:    types.StringPointerValue(project.AWSConfig.EndpointID),
		},
		RoleARN: RoleARNConfig{
			Storage:      types.StringValue(project.AWSConfig.ARN.StorageRoleArn),
			EKS:          types.StringValue(project.AWSConfig.ARN.EksRoleArn),
			CrossAccount: types.StringValue(project.AWSConfig.ARN.BootstrapRoleArn),
		},
		Storage: StorageConfig{
			BucketID: types.StringValue(project.AWSConfig.BucketID),
		},
	}

	data.Instances = InstancesConfig{
		Core: CoreVMConfig{
			VM:    types.StringValue(project.VMCombine.CoreVM),
			Count: types.Int64Value(project.VMCombine.CoreMin),
		},
		Fundamental: VMConfig{
			VM:       types.StringValue(project.VMCombine.FundamentalVM),
			MinCount: types.Int64Value(project.VMCombine.FundamentalMin),
			MaxCount: types.Int64Value(project.VMCombine.FundamentalMax),
		},
		Search: VMConfig{
			VM:       types.StringValue(project.VMCombine.SearchVM),
			MinCount: types.Int64Value(project.VMCombine.SearchMin),
			MaxCount: types.Int64Value(project.VMCombine.SearchMax),
		},
		Index: VMConfig{
			VM:       types.StringValue(project.VMCombine.IndexVM),
			MinCount: types.Int64Value(project.VMCombine.IndexMin),
			MaxCount: types.Int64Value(project.VMCombine.IndexMax),
		},
		// AutoScaling: types.BoolValue(project.AWSConfig.VMCombine.AutoScaling),
		// Arch:        types.StringValue(project.AWSConfig.VMCombine.Arch),
	}

	return data, nil
}

func (s *byocProjectStore) Create(ctx context.Context, data *BYOCProjectResourceModel, updateStateFunc func(project *BYOCProjectResourceModel) error) (err error) {
	var request zilliz.CreateBYOCProjectRequest
	if data.AWS == nil {
		request = zilliz.CreateBYOCProjectRequest{
			ProjectName: data.Name.ValueString(),
		}
	} else if data.AWS != nil {
		var subnetIDs []string
		var securityGroupIDs []string
		data.AWS.Network.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)
		data.AWS.Network.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)

		request = zilliz.CreateBYOCProjectRequest{
			ProjectName: data.Name.ValueString(),
			RegionID:    data.AWS.Region.ValueString(),
			CloudID:     zilliz.CloudId("aws"),

			FundamentalVM: data.Instances.Fundamental.VM.ValueString(),
			SearchVM:      data.Instances.Search.VM.ValueString(),
			CoreVM:        data.Instances.Core.VM.ValueString(),
			IndexVM:       data.Instances.Index.VM.ValueString(),

			SearchMin:      data.Instances.Search.MinCount.ValueInt64(),
			SearchMax:      data.Instances.Search.MaxCount.ValueInt64(),
			FundamentalMin: data.Instances.Fundamental.MinCount.ValueInt64(),
			FundamentalMax: data.Instances.Fundamental.MaxCount.ValueInt64(),
			CoreMin:        data.Instances.Core.Count.ValueInt64(),
			CoreMax:        data.Instances.Core.Count.ValueInt64(),
			IndexMin:       data.Instances.Index.MinCount.ValueInt64(),
			IndexMax:       data.Instances.Index.MaxCount.ValueInt64(),

			AutoScaling: data.Instances.AutoScaling.ValueBool(),
			Arch:        data.Instances.Arch.ValueString(),

			DeployType: TERRAFORM_DEPLOY_TYPE,
			AWSParam: &zilliz.AWSParam{
				BucketID:         data.AWS.Storage.BucketID.ValueString(),
				StorageRoleArn:   data.AWS.RoleARN.Storage.ValueString(),
				EksRoleArn:       data.AWS.RoleARN.EKS.ValueString(),
				BootstrapRoleArn: data.AWS.RoleARN.CrossAccount.ValueString(),
				UserVpcID:        data.AWS.Network.VPCID.ValueString(),
				SubnetIDs:        subnetIDs,
				SecurityGroupIDs: securityGroupIDs,
				VPCEndpointID:    data.AWS.Network.VPCEndpointID.ValueStringPointer(),
			},
		}
	}
	tflog.Info(ctx, fmt.Sprintf("Create BYOC Project request: %s", func() string {
		json, _ := json.Marshal(request)
		return string(json)
	}()))

	response, err := s.client.CreateBYOCProject(&request)
	if err != nil {
		return fmt.Errorf("failed to create BYOC project: %w", err)
	}

	data.ID = types.StringValue(response.ProjectId)
	data.DataPlaneID = types.StringValue(response.DataPlaneId)

	if err = updateStateFunc(data); err != nil {
		return err
	}

	return nil

}

func (s *byocProjectStore) Delete(ctx context.Context, data *BYOCProjectResourceModel) (err error) {
	projectID := data.ID.ValueString()
	dataPlaneID := data.DataPlaneID.ValueString()
	{
		project, err := s.Describe(ctx, projectID, dataPlaneID)
		if err != nil {
			return fmt.Errorf("failed to describe BYOC project: %w", err)
		}

		if project.Status.ValueString() != BYOCProjectStatusDeleted.String() && project.Status.ValueString() != BYOCProjectStatusDeleting.String() {
			_, err = s.client.DeleteBYOCProject(&zilliz.DeleteBYOCProjectRequest{
				ProjectId:   projectID,
				DataPlaneID: dataPlaneID,
			})

			if err != nil {
				return fmt.Errorf("failed to delete BYOC project: %w", err)
			}
		}

	}
	timeout, diags := data.Timeouts.Delete(ctx, defaultBYOCProjectDeleteTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get delete timeout")
	}

	_, err = util.Poll[any](ctx, timeout, func() (*any, *util.Err) {

		project, err := s.Describe(ctx, projectID, dataPlaneID)
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC project status")}
		}

		if project.Status.ValueString() == BYOCProjectStatusDeleting.String() {
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is still deleting")}
		}

		if project.Status.ValueString() == BYOCProjectStatusDeleted.String() {
			return nil, nil
		}

		return nil, &util.Err{Halt: true, Err: fmt.Errorf("BYOC project is in unknown state: %s", project.Status.ValueString())}
	})

	if err != nil {
		return fmt.Errorf("failed to delete BYOC project: %w", err)
	}
	return nil
}

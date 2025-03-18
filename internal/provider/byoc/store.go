package byoc

import (
	"context"
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
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is stopping...")}
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
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is resuming...")}
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
		Instances: InstancesConfig{
			CoreVM:        types.StringValue(project.AWSConfig.VMCombine.CoreVM),
			FundamentalVM: types.StringValue(project.AWSConfig.VMCombine.FundamentalVM),
			SearchVM:      types.StringValue(project.AWSConfig.VMCombine.SearchVM),
		},
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
			ProjectName:   data.Name.ValueString(),
			RegionID:      data.AWS.Region.ValueString(),
			CloudID:       zilliz.CloudId("aws"),
			FundamentalVM: data.AWS.Instances.FundamentalVM.ValueString(),
			SearchVM:      data.AWS.Instances.SearchVM.ValueString(),
			CoreVM:        data.AWS.Instances.CoreVM.ValueString(),
			DeployType:    TERRAFORM_DEPLOY_TYPE,
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
	tflog.Info(ctx, fmt.Sprintf("Create BYOC Project request: %+v", request))

	response, err := s.client.CreateBYOCProject(&request)
	if err != nil {
		return fmt.Errorf("failed to create BYOC project: %w", err)
	}

	data.ID = types.StringValue(response.ProjectId)
	data.DataPlaneID = types.StringValue(response.DataPlaneId)

	if err = updateStateFunc(data); err != nil {
		return err
	}

	timeout, diags := data.Timeouts.Create(ctx, defaultBYOCProjectCreateTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get create timeout")
	}

	_, err = util.Poll[BYOCProjectResourceModel](ctx, timeout, func() (*BYOCProjectResourceModel, *util.Err) {

		project, err := s.Describe(ctx, data.ID.ValueString(), data.DataPlaneID.ValueString())
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC project status")}
		}

		switch project.Status.ValueString() {
		case BYOCProjectStatusPending.String():
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is pending...")}
		case BYOCProjectStatusFailed.String():
			return nil, &util.Err{Err: fmt.Errorf("BYOC project failed to create...")}
		case BYOCProjectStatusRunning.String():
			// achieved the target status
			return &project, nil
		default:
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("BYOC project is in unknown state: %s", project.Status.ValueString())}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to create BYOC project: %w", err)
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

		if !(project.Status.ValueString() == BYOCProjectStatusDeleted.String() || project.Status.ValueString() == BYOCProjectStatusDeleting.String()) {
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
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is still deleting...")}
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

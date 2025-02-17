package byoc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type ByocProjectStore interface {
	Create(ctx context.Context, model *BYOCProjectResourceModel) (projectID string, dataPlaneID string, err error)
	Delete(ctx context.Context, projectID string, dataPlaneID string) (err error)
	Describe(ctx context.Context, projectID string, dataPlaneID string) (model BYOCProjectResourceModel, diags diag.Diagnostics)
	waitForStatus(ctx context.Context, timeout time.Duration, projectID string, dataPlaneID string, status BYOCProjectStatus) (diags diag.Diagnostics)
}

type byocProjectStore struct {
	client *zilliz.Client
}

func (s *byocProjectStore) Describe(ctx context.Context, projectID string, dataPlaneID string) (data BYOCProjectResourceModel, diags diag.Diagnostics) {
	var err error

	project, err := s.client.DescribeBYOCProject(&zilliz.DescribeBYOCProjectRequest{
		ProjectId:   projectID,
		DataPlaneID: dataPlaneID,
	})
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to DescribeBYOCProject, got error: %s", err))
		return data, diags
	}

	data.Status = types.Int64Value(int64(project.Status))
	data.ID = types.StringValue(project.ProjectID)
	// data.Name = types.StringValue(project.ProjectName)
	data.DataPlaneID = types.StringValue(project.DataPlaneID)
	subnetIDs, diags := sliceToTerraformSet(project.AWSConfig.SubnetIDs)

	if diags.HasError() {
		return data, diags
	}
	securityGroupIDs, diags := sliceToTerraformSet(project.AWSConfig.SecurityGroupIDs)
	if diags.HasError() {
		return data, diags
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

func (s *byocProjectStore) Create(ctx context.Context, data *BYOCProjectResourceModel) (projectID string, dataPlaneID string, err error) {
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
			// BYOCID:        data.AWS.Storage.BucketID.ValueString(),
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

	// TODO: Implement create logic using client
	response, err := s.client.CreateBYOCProject(&request)
	if err != nil {
		return "", "", fmt.Errorf("failed to create BYOC project: %w", err)
	}
	return response.ProjectId, response.DataPlaneId, nil
}

func (s *byocProjectStore) Delete(ctx context.Context, projectID string, dataPlaneID string) (err error) {
	_, err = s.client.DeleteBYOCProject(&zilliz.DeleteBYOCProjectRequest{
		ProjectId:   projectID,
		DataPlaneID: dataPlaneID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete BYOC project: %w", err)
	}
	return nil
}

func (s *byocProjectStore) waitForStatus(ctx context.Context, timeout time.Duration, projectID string, dataPlaneID string, status BYOCProjectStatus) (diags diag.Diagnostics) {
	tflog.Info(ctx, fmt.Sprintf("Waiting for BYOC project to enter the %s state...", status))
	tflog.Info(ctx, fmt.Sprintf("Project ID: %s, Data Plane ID: %s", projectID, dataPlaneID))

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		project, diags := s.Describe(ctx, projectID, dataPlaneID)
		if diags.HasError() {
			return retry.NonRetryableError(fmt.Errorf("failed to describe BYOC project"))
		}

		tflog.Info(ctx, fmt.Sprintf("Describe BYOC Project response: %+v", project))

		if BYOCProjectStatus(project.Status.ValueInt64()) == BYOCProjectStatusFailed {
			return retry.NonRetryableError(fmt.Errorf("BYOC project failed to create..."))
		}

		if BYOCProjectStatus(project.Status.ValueInt64()) != status {
			return retry.RetryableError(fmt.Errorf("BYOC project not yet in the %s state. Current state: %d", status, project.Status))
		}
		return nil
	})
	if err != nil {
		diags.AddError("Failed to wait for BYOC project to enter the RUNNING state.", err.Error())
	}
	return diags
}

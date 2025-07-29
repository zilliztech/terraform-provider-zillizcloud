package byoc_op

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	util "github.com/zilliztech/terraform-provider-zillizcloud/client/retry"
)

type ByocOpProjectStore interface {
	Create(ctx context.Context, data *BYOCOpProjectResourceModel, updateStatusFunc func(project *BYOCOpProjectResourceModel) error) (err error)
	Delete(ctx context.Context, data *BYOCOpProjectResourceModel) (err error)
	Describe(ctx context.Context, projectID string, dataPlaneID string) (model BYOCOpProjectResourceModel, err error)
}

type byocOpProjectStore struct {
	client *zilliz.Client
}

var _ ByocOpProjectStore = &byocOpProjectStore{}

func (s *byocOpProjectStore) Create(ctx context.Context, data *BYOCOpProjectResourceModel, updateStateFunc func(project *BYOCOpProjectResourceModel) error) (err error) {

	var request zilliz.CreateByocOpProjectRequest
	var subnetIDs []string
	var securityGroupIDs []string
	data.AWS.Network.SubnetIDs.ElementsAs(ctx, &subnetIDs, false)
	data.AWS.Network.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIDs, false)

	if data.AWS != nil {
		request = zilliz.CreateByocOpProjectRequest{
			ProjectId:   data.ProjectID.ValueString(),
			DataPlaneId: data.DataPlaneID.ValueString(),

			RegionID: data.AWS.Region.ValueString(),

			CloudID: zilliz.CloudId("aws"),

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

		// if data.ExtConfig.ValueString() is set, set it to ExtConfig
		if !data.ExtConfig.IsNull() {
			request.ExtConfig = data.ExtConfig.ValueStringPointer()
		}

	}

	tflog.Info(ctx, fmt.Sprintf("Create BYOC-I Project request: %+v", request))

	response, err := s.client.CreateByocOpProject(&request)
	if err != nil {
		return fmt.Errorf("failed to create BYOC-I project: %w", err)
	}

	data.ID = types.StringValue(response.ProjectId)
	data.Status = types.Int64Value(0) // Pending status

	if err = updateStateFunc(data); err != nil {
		return err
	}

	_, diags := data.Timeouts.Create(ctx, defaultBYOCOpProjectCreateTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get create timeout")
	}

	if err = updateStateFunc(data); err != nil {
		return err
	}

	return nil
}

func (s *byocOpProjectStore) Delete(ctx context.Context, data *BYOCOpProjectResourceModel) (err error) {
	request := &zilliz.DeleteBYOCProjectRequest{
		ProjectId:   data.ProjectID.ValueString(),
		DataPlaneID: data.DataPlaneID.ValueString(),
	}

	tflog.Info(ctx, fmt.Sprintf("Delete BYOC-I Project request: %+v", request))

	{

		project, err := s.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
		if err != nil {
			return fmt.Errorf("failed to describe BYOC-I project: %w", err)
		}

		tflog.Info(ctx, fmt.Sprintf("Before delete BYOC-I Project, peek the status: %d", project.Status.ValueInt64()))

		//prompt user delete project from console if project has not been deleted
		if project.Status.ValueInt64() != int64(BYOCProjectStatusDeleted) {
			return fmt.Errorf("please initiate the project deletion directly from the console and wait for that process to fully complete. Once the project is confirmed as deleted from the console, you can then attempt to delete it using Terraform")
		}

	}

	// get timeout for delete
	timeout, diags := data.Timeouts.Delete(ctx, defaultBYOCOpProjectDeleteTimeout)
	if diags.HasError() {
		return fmt.Errorf("failed to get delete timeout")
	}

	_, err = util.Poll[BYOCOpProjectResourceModel](ctx, timeout, func() (*BYOCOpProjectResourceModel, *util.Err) {
		project, err := s.Describe(ctx, data.ProjectID.ValueString(), data.DataPlaneID.ValueString())
		if err != nil {
			return nil, &util.Err{Halt: true, Err: fmt.Errorf("failed to check BYOC project status")}
		}

		if project.Status.ValueInt64() == int64(BYOCProjectStatusDeleting) {
			return nil, &util.Err{Err: fmt.Errorf("BYOC project is still deleting")}
		}

		if project.Status.ValueInt64() == int64(BYOCProjectStatusDeleted) {
			return nil, nil
		}

		return nil, &util.Err{Halt: true, Err: fmt.Errorf("BYOC project is in unknown state: %d", project.Status.ValueInt64())}
	})

	return err
}

func (s *byocOpProjectStore) Describe(ctx context.Context, projectID string, dataPlaneID string) (data BYOCOpProjectResourceModel, err error) {
	request := &zilliz.DescribeByocOpProjectRequest{
		ProjectId:   projectID,
		DataPlaneID: dataPlaneID,
	}

	tflog.Info(ctx, fmt.Sprintf("Describe BYOC-I Project request: %+v", request))

	response, err := s.client.DescribeByocOpProject(request)
	if err != nil {
		return data, fmt.Errorf("failed to describe BYOC-I project: %w", err)
	}

	data.ID = types.StringValue(response.ProjectID)
	data.ProjectID = types.StringValue(response.ProjectID)
	data.DataPlaneID = types.StringValue(response.DataPlaneID)
	data.Status = types.Int64Value(int64(response.Status))
	if response.OpConfig != nil {
		data.ExtConfig = types.StringValue(response.OpConfig.Token)
	}

	// Convert response to model
	data = BYOCOpProjectResourceModel{
		ID:          types.StringValue(response.ProjectID),
		ProjectID:   types.StringValue(response.ProjectID),
		DataPlaneID: types.StringValue(response.DataPlaneID),
		Status:      types.Int64Value(int64(response.Status)),
	}

	return data, nil
}

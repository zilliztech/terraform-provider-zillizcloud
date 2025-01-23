package byoc

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BYOCProjectResourceModel describes the resource data model.
type BYOCProjectResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	DataPlaneID types.String   `tfsdk:"data_plane_id"`
	AWS         *AWSConfig     `tfsdk:"aws"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	Status      types.Int64    `tfsdk:"status"`
}

type AWSConfig struct {
	Region    types.String     `tfsdk:"region"`
	Network   *NetworkConfig   `tfsdk:"network"`
	RoleARN   *RoleARNConfig   `tfsdk:"role_arn"`
	Storage   *StorageConfig   `tfsdk:"storage"`
	Instances *InstancesConfig `tfsdk:"instances"`
}

type NetworkConfig struct {
	VPCID            types.String `tfsdk:"vpc_id"`
	SubnetIDs        types.Set    `tfsdk:"subnet_ids"`
	SecurityGroupIDs types.Set    `tfsdk:"security_group_ids"`
	VPCEndpointID    types.String `tfsdk:"vpc_endpoint_id"`
}

type RoleARNConfig struct {
	Storage      types.String `tfsdk:"storage"`
	EKS          types.String `tfsdk:"eks"`
	CrossAccount types.String `tfsdk:"cross_account"`
}

type StorageConfig struct {
	BucketID types.String `tfsdk:"bucket_id"`
}

type InstancesConfig struct {
	CoreVM        types.String `tfsdk:"core_vm"`
	FundamentalVM types.String `tfsdk:"fundamental_vm"`
	SearchVM      types.String `tfsdk:"search_vm"`
}

// refresh
func (data *BYOCProjectResourceModel) refresh(input BYOCProjectResourceModel) {

	data.AWS.Network.SubnetIDs = input.AWS.Network.SubnetIDs
	data.AWS.Network.SecurityGroupIDs = input.AWS.Network.SecurityGroupIDs
	data.AWS.Network.VPCEndpointID = input.AWS.Network.VPCEndpointID

	data.AWS.RoleARN.Storage = input.AWS.RoleARN.Storage
	data.AWS.RoleARN.EKS = input.AWS.RoleARN.EKS
	data.AWS.RoleARN.CrossAccount = input.AWS.RoleARN.CrossAccount

	data.AWS.Storage.BucketID = input.AWS.Storage.BucketID

	data.AWS.Instances.CoreVM = input.AWS.Instances.CoreVM
	data.AWS.Instances.FundamentalVM = input.AWS.Instances.FundamentalVM
	data.AWS.Instances.SearchVM = input.AWS.Instances.SearchVM

	data.Status = input.Status
}

type BYOCProjectStatus int

const (
	BYOCProjectStatusPending BYOCProjectStatus = iota
	BYOCProjectStatusRunning
	BYOCProjectStatusDeleting
	BYOCProjectStatusDeleted
	BYOCProjectStatusUpgrading
	BYOCProjectStatusFailed
	BYOCProjectStatusStopping
	BYOCProjectStatusStopped
	BYOCProjectStatusResuming
)

func (s BYOCProjectStatus) String() string {
	switch s {
	case BYOCProjectStatusPending:
		return "PENDING"
	case BYOCProjectStatusRunning:
		return "RUNNING"
	case BYOCProjectStatusDeleting:
		return "DELETING"
	case BYOCProjectStatusDeleted:
		return "DELETED"
	case BYOCProjectStatusUpgrading:
		return "UPGRADING"
	case BYOCProjectStatusFailed:
		return "FAILED"
	case BYOCProjectStatusStopping:
		return "STOPPING"
	case BYOCProjectStatusStopped:
		return "STOPPED"
	case BYOCProjectStatusResuming:
		return "RESUMING"
	default:
		return "UNKNOWN"
	}
}

package byoc

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BYOCProjectResourceModel describes the resource data model.
type BYOCProjectResourceModel struct {
	ID          types.String    `tfsdk:"id"`
	Name        types.String    `tfsdk:"name"`
	DataPlaneID types.String    `tfsdk:"data_plane_id"`
	AWS         *AWSConfig      `tfsdk:"aws"`
	Instances   InstancesConfig `tfsdk:"instances"`
	Timeouts    timeouts.Value  `tfsdk:"timeouts"`
	Status      types.String    `tfsdk:"status"`
}

type AWSConfig struct {
	Region  types.String  `tfsdk:"region"`
	Network NetworkConfig `tfsdk:"network"`
	RoleARN RoleARNConfig `tfsdk:"role_arn"`
	Storage StorageConfig `tfsdk:"storage"`
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

type VMConfig struct {
	VM       types.String `tfsdk:"vm"`
	MinCount types.Int64  `tfsdk:"min_count"`
	MaxCount types.Int64  `tfsdk:"max_count"`
}

type CoreVMConfig struct {
	VM    types.String `tfsdk:"vm"`
	Count types.Int64  `tfsdk:"count"`
}

type InstancesConfig struct {
	Core        CoreVMConfig `tfsdk:"core"`
	Fundamental VMConfig     `tfsdk:"fundamental"`
	Search      VMConfig     `tfsdk:"search"`
	Index       VMConfig     `tfsdk:"index"`
	AutoScaling types.Bool   `tfsdk:"auto_scaling"`
	Arch        types.String `tfsdk:"arch"`
}

func (data *BYOCProjectResourceModel) refresh(input BYOCProjectResourceModel) {
	data.AWS = input.AWS
	data.Instances = input.Instances
	data.Status = input.Status
	data.DataPlaneID = input.DataPlaneID
	// data.Name = input.Name
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

// Additional statuses with explicit values.
const (
	BYOCProjectStatusInit      BYOCProjectStatus = 99
	BYOCProjectStatusConnected BYOCProjectStatus = 90
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
	case BYOCProjectStatusInit:
		return "INIT"
	case BYOCProjectStatusConnected:
		return "CONNECTED"
	default:
		return "UNKNOWN"
	}
}

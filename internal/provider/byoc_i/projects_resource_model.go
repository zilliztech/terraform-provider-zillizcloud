package byoc_op

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// 7 is the deploy type for BYOC-I Project.
	TERRAFORM_DEPLOY_TYPE = 7
)

type BYOCOpProjectResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	ProjectID   types.String   `tfsdk:"project_id"`
	DataPlaneID types.String   `tfsdk:"data_plane_id"`
	ExtConfig   types.String   `tfsdk:"ext_config"`
	AWS         *AWSConfig     `tfsdk:"aws"`
	Azure       *AzureConfig   `tfsdk:"azure"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	Status      types.Int64    `tfsdk:"status"`
}

type AWSConfig struct {
	Region  types.String  `tfsdk:"region"`
	Network NetworkConfig `tfsdk:"network"`
	RoleARN RoleARNConfig `tfsdk:"role_arn"`
	Storage StorageConfig `tfsdk:"storage"`
	CSE     *CSEConfig    `tfsdk:"cse"`
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

// CSEConfig contains Client-Side Encryption configuration for AWS KMS
type CSEConfig struct {
	AwsCseRoleArn       types.String `tfsdk:"aws_cse_role_arn"`
	DefaultAwsCseKeyArn types.String `tfsdk:"default_aws_cse_key_arn"`
	ExternalID          types.String `tfsdk:"external_id"`
}

type AzureConfig struct {
	Region   types.String        `tfsdk:"region"`
	Network  AzureNetworkConfig  `tfsdk:"network"`
	Identity AzureIdentityConfig `tfsdk:"identity"`
	Storage  AzureStorageConfig  `tfsdk:"storage"`
}

type AzureNetworkConfig struct {
	VNetID            types.String `tfsdk:"vnet_id"`
	SubnetIDs         types.Set    `tfsdk:"subnet_ids"`
	NSGIDs            types.Set    `tfsdk:"nsg_ids"`
	PrivateEndpointID types.String `tfsdk:"private_endpoint_id"`
}

type AzureIdentityConfig struct {
	Storages    types.Set     `tfsdk:"storages"`
	Kubelet     AzureIdentity `tfsdk:"kubelet"`
	Maintenance AzureIdentity `tfsdk:"maintenance"`
}

type AzureIdentity struct {
	PrincipalID types.String `tfsdk:"principal_id"`
	ClientID    types.String `tfsdk:"client_id"`
	ResourceID  types.String `tfsdk:"resource_id"`
}

type AzureStorageConfig struct {
	StorageAccountName types.String `tfsdk:"storage_account_name"`
	ContainerName      types.String `tfsdk:"container_name"`
}

type Instances struct {
	CoreVM        types.String `tfsdk:"core_vm"`
	FundamentalVM types.String `tfsdk:"fundamental_vm"`
	SearchVM      types.String `tfsdk:"search_vm"`
}

type VMOpConfig struct {
	VM       types.String `tfsdk:"vm"`
	MinCount types.Int64  `tfsdk:"min_count"`
	MaxCount types.Int64  `tfsdk:"max_count"`
}

type CoreVMOpConfig struct {
	VM    types.String `tfsdk:"vm"`
	Count types.Int64  `tfsdk:"count"`
}

type InstancesOpConfig struct {
	Core        CoreVMOpConfig `tfsdk:"core"`
	Fundamental VMOpConfig     `tfsdk:"fundamental"`
	Search      VMOpConfig     `tfsdk:"search"`
	Index       VMOpConfig     `tfsdk:"index"`
	AutoScaling types.Bool     `tfsdk:"auto_scaling"`
	Arch        types.String   `tfsdk:"arch"`
}

func (data *BYOCOpProjectResourceModel) refresh(input BYOCOpProjectResourceModel) {
	data.AWS = input.AWS
	data.Azure = input.Azure
	data.Status = input.Status
	data.DataPlaneID = input.DataPlaneID
	data.ProjectID = input.ProjectID
	data.ExtConfig = input.ExtConfig
}

type BYOCOpProjectSettingsResourceModel struct {
	ID                 types.String      `tfsdk:"id"`
	ProjectID          types.String      `tfsdk:"project_id"`
	ProjectName        types.String      `tfsdk:"project_name"`
	DataPlaneID        types.String      `tfsdk:"data_plane_id"`
	Instances          InstancesOpConfig `tfsdk:"instances"`
	CloudProvider      types.String      `tfsdk:"cloud_provider"`
	Region             types.String      `tfsdk:"region"`
	OpConfig           types.Object      `tfsdk:"op_config"`
	NodeQuotas         types.Object      `tfsdk:"node_quotas"`
	PrivateLinkEnabled types.Bool        `tfsdk:"private_link_enabled"`
}

type OpConfig struct {
	Token         types.String `tfsdk:"token"`
	AgentImageUrl types.String `tfsdk:"agent_image_url"`
}

type NodeQuotas struct {
	Core        NodeQuota `tfsdk:"core"`
	Index       NodeQuota `tfsdk:"index"`
	Search      NodeQuota `tfsdk:"search"`
	Fundamental NodeQuota `tfsdk:"fundamental"`
}

type NodeQuota struct {
	DiskSize      types.Int64  `tfsdk:"disk_size"`
	MinSize       types.Int64  `tfsdk:"min_size"`
	MaxSize       types.Int64  `tfsdk:"max_size"`
	DesiredSize   types.Int64  `tfsdk:"desired_size"`
	InstanceTypes types.String `tfsdk:"instance_types"`
	CapacityType  types.String `tfsdk:"capacity_type"`
}

type BYOCOpProjectSettingsDataModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	DataPlaneID        types.String `tfsdk:"data_plane_id"`
	ProjectName        types.String `tfsdk:"project_name"`
	CloudProvider      types.String `tfsdk:"cloud_provider"`
	Region             types.String `tfsdk:"region"`
	OpConfig           types.Object `tfsdk:"op_config"`
	NodeQuotas         types.Object `tfsdk:"node_quotas"`
	PrivateLinkEnabled types.Bool   `tfsdk:"private_link_enabled"`
}

func (data *BYOCOpProjectSettingsDataModel) refresh(input BYOCOpProjectSettingsDataModel) {
	data.ID = input.ID
	data.ProjectID = input.ProjectID
	data.DataPlaneID = input.DataPlaneID
	data.CloudProvider = input.CloudProvider
	data.Region = input.Region
	data.OpConfig = input.OpConfig
	data.NodeQuotas = input.NodeQuotas
	data.PrivateLinkEnabled = input.PrivateLinkEnabled
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
	case BYOCProjectStatusInit:
		return "init"
	case BYOCProjectStatusConnected:
		return "connected"
	case BYOCProjectStatusPending:
		return "pending"
	case BYOCProjectStatusRunning:
		return "running"
	case BYOCProjectStatusDeleting:
		return "deleting"
	case BYOCProjectStatusDeleted:
		return "deleted"
	case BYOCProjectStatusUpgrading:
		return "upgrading"
	case BYOCProjectStatusFailed:
		return "failed"
	case BYOCProjectStatusStopping:
		return "stopping"
	case BYOCProjectStatusStopped:
		return "stopped"
	case BYOCProjectStatusResuming:
		return "resuming"
	default:
		return "unknown"
	}
}

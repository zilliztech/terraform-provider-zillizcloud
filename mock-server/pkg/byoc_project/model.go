package byoc_project

import (
	"time"

	"github.com/google/uuid"
)

// AWS specific parameters
type AWSParams struct {
	BucketID string `json:"bucketId"`

	StorageRoleArn   string `json:"storageRoleArn"`
	EksRoleArn       string `json:"eksRoleArn"`
	BootstrapRoleArn string `json:"bootstrapRoleArn"`

	UserVpcID        string   `json:"userVpcId"`
	SubnetIds        []string `json:"subnetIds"`
	SecurityGroupIds []string `json:"securityGroupIds"`

	EndpointId *string `json:"endpointId"`
}

// Main request structure
type CreateDataplaneRequest struct {
	AWSParam    AWSParams `json:"awsParam"`
	ProjectName string    `json:"projectName"`
	RegionID    string    `json:"regionId"`
	CloudID     string    `json:"cloudId"`
	ByocID      string    `json:"byocId"`

	FundamentalVm string `json:"fundamentalVm"`
	SearchVm      string `json:"searchVm"`
	CoreVm        string `json:"coreVm"`

	SearchMin      int64 `json:"searchMin"`
	FundamentalMin int64 `json:"fundamentalMin"`
	CoreMin        int64 `json:"coreMin"`

	DeployType int `json:"deployType"`
}
type CreateOpDataplaneRequest struct {
	AWSParam      AWSParams `json:"awsParam"`
	ProjectName   string    `json:"projectName"`
	ProjectID     string    `json:"projectId"`
	DataPlaneID   string    `json:"dataPlaneId"`
	RegionID      string    `json:"regionId"`
	CloudID       string    `json:"cloudId"`
	ByocID        string    `json:"byocId"`
	FundamentalVm string    `json:"fundamentalVm"`
	SearchVm      string    `json:"searchVm"`
	CoreVm        string    `json:"coreVm"`
	DeployType    int       `json:"deployType"`
	ExtConfig     string    `json:"extConfig"`
}

type Response[T any] struct {
	Code int `json:"code"`
	Data T   `json:"data"`
}

type DataplaneResponse struct {
	CloudID     string     `json:"cloudId"`
	RegionID    string     `json:"regionId"`
	ProjectID   string     `json:"projectId"`
	ProjectName string     `json:"projectName"`
	DataPlaneID string     `json:"dataPlaneId"`
	Status      int        `json:"status"`
	DeployType  DeployType `json:"deployType"`
	Message     string     `json:"message"`
	AWSConfig   struct {
		BucketID string `json:"bucketId"`
		ARN      struct {
			BootstrapRoleArn string `json:"bootstrapRoleArn"`
			EksRoleArn       string `json:"eksRoleArn"`
			StorageRoleArn   string `json:"storageRoleArn"`
		} `json:"arn"`
		VMCombine struct {
			SearchVM       string `json:"searchVm"`
			FundamentalVM  string `json:"fundamentalVm"`
			CoreVM         string `json:"coreVm"`
			SearchMin      int64  `json:"searchMin"`
			FundamentalMin int64  `json:"fundamentalMin"`
			CoreMin        int64  `json:"coreMin"`
		} `json:"vmCombine"`
		VpcID            string   `json:"vpcId"`
		SubnetIds        []string `json:"subnetIds"`
		SecurityGroupIds []string `json:"securityGroupIds"`
		EndpointID       *string  `json:"endpointId,omitempty"`
	} `json:"awsConfig"`
	GCPConfig       interface{} `json:"gcpConfig"`
	CreateTimeMilli int64       `json:"createTimeMilli"`
	LastUpdateMilli int64       `json:"lastUpdateMilli"`
}

func (r *CreateDataplaneRequest) ToDataplane() *DataplaneResponse {
	response := &DataplaneResponse{}
	response.CloudID = r.CloudID
	response.RegionID = r.RegionID
	response.ProjectName = r.ProjectName
	response.ProjectID = "proj-" + uuid.New().String()
	response.DataPlaneID = "zilliz-" + uuid.New().String()
	response.Status = 0
	response.Message = ""
	response.AWSConfig.BucketID = r.AWSParam.BucketID

	response.AWSConfig.ARN.BootstrapRoleArn = r.AWSParam.BootstrapRoleArn
	response.AWSConfig.ARN.EksRoleArn = r.AWSParam.EksRoleArn
	response.AWSConfig.ARN.StorageRoleArn = r.AWSParam.StorageRoleArn

	response.AWSConfig.VMCombine.SearchVM = r.SearchVm
	response.AWSConfig.VMCombine.FundamentalVM = r.FundamentalVm
	response.AWSConfig.VMCombine.CoreVM = r.CoreVm
	response.AWSConfig.VMCombine.CoreMin = r.CoreMin
	response.AWSConfig.VMCombine.FundamentalMin = r.FundamentalMin
	response.AWSConfig.VMCombine.SearchMin = r.SearchMin

	response.AWSConfig.VpcID = r.AWSParam.UserVpcID
	response.AWSConfig.SubnetIds = r.AWSParam.SubnetIds
	response.AWSConfig.SecurityGroupIds = r.AWSParam.SecurityGroupIds
	response.AWSConfig.EndpointID = r.AWSParam.EndpointId

	response.CreateTimeMilli = time.Now().UnixMilli()
	response.LastUpdateMilli = time.Now().UnixMilli()
	return response
}

type DescribeDataplaneRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type DeleteDataplaneRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type SuspendDataplaneRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type ResumeDataplaneRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type DeployType int

const (
	DeployTypeManual         DeployType = 3
	DeployTypeCloudFormation DeployType = 4
	DeployTypeTerraform      DeployType = 5
	DeployTypeOpConsole      DeployType = 6
	DeployTypeOpTerraform    DeployType = 7
)

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

// Additional statuses with explicit values
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

//	{
//		"projectName": "project03",
//		"regionId": "aws-us-west-2",
//		"cloudId": "aws",
//		"byocId": "byoc-d8f2316a76e213dc5d5410",
//		"fundamentalVm": "m6i.2xlarge",
//		"searchVm": "m6id.2xlarge",
//		"coreVm": "m6i.2xlarge",
//		"deployType": 5
//	  }
type SettingsRequest struct {
	ProjectName    string  `json:"projectName"`
	ProjectId      *string `json:"projectId"`
	DataPlaneId    *string `json:"dataPlaneId"`
	Id             *string `json:"id"`
	RegionId       string  `json:"regionId"`
	CloudId        string  `json:"cloudId"`
	ByocId         string  `json:"byocId"`
	FundamentalVm  string  `json:"fundamentalVm"`
	SearchVm       string  `json:"searchVm"`
	CoreVm         string  `json:"coreVm"`
	SearchMin      int64   `json:"searchMin"`
	FundamentalMin int64   `json:"fundamentalMin"`
	CoreMin        int64   `json:"coreMin"`
	DeployType     int     `json:"deployType"`
	OpenPl         int     `json:"openPl"`
}
type NodeQuota struct {
	DesiredSize   int      `json:"desired_size"`
	DiskSize      int      `json:"disk_size"`
	InstanceTypes []string `json:"instance_types"`
	MaxSize       int      `json:"max_size"`
	MinSize       int      `json:"min_size"`
	Name          string   `json:"name"`
	CapacityType  string   `json:"capacity_type"`
}

type SettingsResponse struct {
	ProjectId   string      `json:"projectId"`
	DataPlaneId string      `json:"dataPlaneId"`
	ProjectName string      `json:"projectName"`
	RegionId    string      `json:"regionId"`
	CloudId     string      `json:"cloudId"`
	ByocId      string      `json:"byocId"`
	NodeQuotas  []NodeQuota `json:"nodeQuotas"`
	OpenPl      int         `json:"openPl"`
	OpConfig    struct {
		Token         string `json:"tunnelToken"`
		AgentImageUrl string `json:"agentImageUrl"`
	} `json:"opConfig"`
}

func (s *SettingsResponse) IntoDataplane() *DataplaneResponse {
	return &DataplaneResponse{
		ProjectID:   s.ProjectId,
		DataPlaneID: s.DataPlaneId,
		ProjectName: s.ProjectName,
		RegionID:    s.RegionId,
		CloudID:     s.CloudId,
	}
}

type Option func(*SettingsResponse)

func WithProjectId(projectId *string) Option {
	return func(s *SettingsResponse) {
		if projectId != nil {
			s.ProjectId = *projectId
		} else {
			s.ProjectId = "proj-" + uuid.New().String()
		}
	}
}

func WithDataPlaneId(dataPlaneId *string) Option {
	return func(s *SettingsResponse) {
		if dataPlaneId != nil {
			s.DataPlaneId = *dataPlaneId
		} else {
			s.DataPlaneId = "zilliz-" + uuid.New().String()
		}
	}
}

func WithProjectName(projectName string) Option {
	return func(s *SettingsResponse) {
		s.ProjectName = projectName
	}
}

func WithRegionId(regionId string) Option {
	return func(s *SettingsResponse) {
		s.RegionId = regionId
	}
}

func WithCloudId(cloudId string) Option {
	return func(s *SettingsResponse) {
		s.CloudId = cloudId
	}
}

func WithByocId(byocId string) Option {
	return func(s *SettingsResponse) {
		s.ByocId = byocId
	}
}

func WithNodeQuotas(nodeQuotas []NodeQuota) Option {
	return func(s *SettingsResponse) {
		s.NodeQuotas = nodeQuotas
	}
}

func WithOpenPl(openPl int) Option {
	return func(s *SettingsResponse) {
		s.OpenPl = openPl
	}
}

func WithNodeQuota(Name string, min int) NodeQuota {
	return NodeQuota{
		DesiredSize:   1,
		DiskSize:      200,
		InstanceTypes: []string{"m6i.2xlarge"},
		MaxSize:       100,
		MinSize:       min,
		Name:          Name,
		CapacityType:  "SPOT",
	}
}

func WithOpConfig(token string, agentImageUrl string) Option {
	return func(s *SettingsResponse) {
		s.OpConfig.Token = token
		s.OpConfig.AgentImageUrl = agentImageUrl
	}
}

func NewSettingsResponse(opts ...Option) *SettingsResponse {
	s := &SettingsResponse{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type CreateDedicatedClusterRequest struct {
	ClusterId   string            `json:"clusterId"`
	ClusterName string            `json:"clusterName"`
	ProjectId   string            `json:"projectId"`
	Description string            `json:"description"`
	RegionId    string            `json:"regionId"`
	CuType      string            `json:"cuType"`
	Plan        string            `json:"plan"`
	Replica     int               `json:"replica"`
	CuSize      int               `json:"cuSize"`
	Labels      map[string]string `json:"labels"`
}

type DedicatedClusterResponse struct {
	ClusterId          string            `json:"clusterId"`
	ClusterName        string            `json:"clusterName"`
	ProjectId          string            `json:"projectId"`
	Description        string            `json:"description"`
	RegionId           string            `json:"regionId"`
	CuType             string            `json:"cuType"`
	Plan               string            `json:"plan"`
	Status             string            `json:"status"`
	ConnectAddress     string            `json:"connectAddress"`
	PrivateLinkAddress string            `json:"privateLinkAddress"`
	CreateTime         string            `json:"createTime"`
	Replica            int               `json:"replica"`
	CuSize             int               `json:"cuSize"`
	StorageSize        int               `json:"storageSize"`
	SnapshotNumber     int               `json:"snapshotNumber"`
	CreateProgress     int               `json:"createProgress"`
	Labels             map[string]string `json:"labels"`
	SecurityGroups     []string          `json:"securityGroups"`
	Username           string            `json:"username"`
	Password           string            `json:"password"`
	Prompt             string            `json:"prompt"`
	Autoscaling        Autoscaling       `json:"autoscaling"`
}

type Autoscaling struct {
	CU CU `json:"cu"`
}

type CU struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

type Project struct {
	ProjectName     string `json:"projectName"`
	ProjectId       string `json:"projectId"`
	InstanceCount   int64  `json:"instanceCount"`
	CreateTimeMilli int64  `json:"createTimeMilli"`
	Plan            string `json:"plan"`
}

type CreateProjectRequest struct {
	ProjectName string `json:"projectName"`
	Plan        string `json:"plan"`
}

type UpgradeProjectPlanRequest struct {
	Plan string `json:"plan"`
}

type ModifyReplicaRequest struct {
	Replica int `json:"replica"`
}

type ModifyClusterRequest struct {
	CuSize      *int         `json:"cuSize,omitempty"`
	Autoscaling *Autoscaling `json:"autoscaling,omitempty"`
}

type UpdateLabelsRequest struct {
	Labels map[string]string `json:"labels"`
}

type ModifyPropertiesRequest struct {
	ClusterName string `json:"clusterName"`
}

// Serverless cluster types
type CreateServerlessClusterRequest struct {
	ClusterId   string            `json:"clusterId"`
	ClusterName string            `json:"clusterName"`
	ProjectId   string            `json:"projectId"`
	Description string            `json:"description"`
	RegionId    string            `json:"regionId"`
	Plan        string            `json:"plan"`
	Labels      map[string]string `json:"labels"`
}

type ServerlessClusterResponse struct {
	ClusterId          string            `json:"clusterId"`
	ClusterName        string            `json:"clusterName"`
	ProjectId          string            `json:"projectId"`
	Description        string            `json:"description"`
	RegionId           string            `json:"regionId"`
	Plan               string            `json:"plan"`
	Status             string            `json:"status"`
	ConnectAddress     string            `json:"connectAddress"`
	PrivateLinkAddress string            `json:"privateLinkAddress"`
	CreateTime         string            `json:"createTime"`
	StorageSize        int               `json:"storageSize"`
	SnapshotNumber     int               `json:"snapshotNumber"`
	CreateProgress     int               `json:"createProgress"`
	Labels             map[string]string `json:"labels"`
	Username           string            `json:"username"`
	Password           string            `json:"password"`
	Prompt             string            `json:"prompt"`
}

// Free cluster types
type CreateFreeClusterRequest struct {
	ClusterId   string            `json:"clusterId"`
	ClusterName string            `json:"clusterName"`
	ProjectId   string            `json:"projectId"`
	Description string            `json:"description"`
	RegionId    string            `json:"regionId"`
	Plan        string            `json:"plan"`
	Labels      map[string]string `json:"labels"`
}

type FreeClusterResponse struct {
	ClusterId          string            `json:"clusterId"`
	ClusterName        string            `json:"clusterName"`
	ProjectId          string            `json:"projectId"`
	Description        string            `json:"description"`
	RegionId           string            `json:"regionId"`
	Plan               string            `json:"plan"`
	Status             string            `json:"status"`
	ConnectAddress     string            `json:"connectAddress"`
	PrivateLinkAddress string            `json:"privateLinkAddress"`
	CreateTime         string            `json:"createTime"`
	StorageSize        int               `json:"storageSize"`
	SnapshotNumber     int               `json:"snapshotNumber"`
	CreateProgress     int               `json:"createProgress"`
	Labels             map[string]string `json:"labels"`
	Username           string            `json:"username"`
	Password           string            `json:"password"`
	Prompt             string            `json:"prompt"`
}

// Security Groups request/response types
type UpsertSecurityGroupsRequest struct {
	Ids []string `json:"ids"`
}

type GetSecurityGroupsResponse struct {
	Ids []string `json:"ids,omitempty"`
}

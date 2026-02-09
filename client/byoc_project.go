package client

type CreateBYOCProjectRequest struct {
	AWSParam      *AWSParam `json:"awsParam"`
	ProjectName   string    `json:"projectName"`
	RegionID      string    `json:"regionId"`
	CloudID       CloudId   `json:"cloudId"`
	BYOCID        *string   `json:"byocId"`
	FundamentalVM string    `json:"fundamentalVm"`
	SearchVM      string    `json:"searchVm"`
	CoreVM        string    `json:"coreVm"`
	IndexVM       string    `json:"indexVm"`

	SearchMin      int64 `json:"searchMin"`
	SearchMax      int64 `json:"searchMax"`
	FundamentalMin int64 `json:"fundamentalMin"`
	FundamentalMax int64 `json:"fundamentalMax"`
	CoreMin        int64 `json:"coreMin"`
	CoreMax        int64 `json:"coreMax"`
	IndexMin       int64 `json:"indexMin"`
	IndexMax       int64 `json:"indexMax"`

	AutoScaling bool   `json:"autoScaling"`
	Arch        string `json:"arch"` //X86, ARM

	DeployType int `json:"deployType"`
}

type AWSParam struct {
	BucketID         string `json:"bucketId"`
	StorageRoleArn   string `json:"storageRoleArn"`
	EksRoleArn       string `json:"eksRoleArn"`
	BootstrapRoleArn string `json:"bootstrapRoleArn"`

	// CSE(Client Side Encryption) parameters
	AwsCseRoleArn       string `json:"awsCseRoleArn"`
	DefaultAwsCseKeyArn string `json:"defaultAwsCseKeyArn"`
	ExternalID          string `json:"externalId"`

	UserVpcID        string   `json:"userVpcId"`
	SubnetIDs        []string `json:"subnetIds"`
	SecurityGroupIDs []string `json:"securityGroupIds"`
	VPCEndpointID    *string  `json:"endpointId"`
}

type AzureIdentityParam struct {
	ClientID    string `json:"clientId"`
	ResourceID  string `json:"resourceId"`
	PrincipalID string `json:"principalId"`
}

type AzureParam struct {
	// network parameters
	VNetID            string   `json:"vnetId"`
	SubnetIDs         []string `json:"subnetIds"`
	NSGIDs            []string `json:"nsgIds"`
	PrivateEndpointID *string  `json:"privateEndpointId"`
	// storage parameters
	StorageAccountName string `json:"storageAccountName"`
	ContainerName      string `json:"containerName"`
	// identity parameters
	StorageIdentities   []AzureIdentityParam `json:"storageIdentities"`
	KubeletIdentity     AzureIdentityParam   `json:"kubeletIdentity"`
	MaintenanceIdentity AzureIdentityParam   `json:"maintenanceIdentity"`
}

func (c *Client) CreateBYOCProject(params *CreateBYOCProjectRequest) (*CreateBYOCProjectResponse, error) {
	var response zillizResponse[CreateBYOCProjectResponse]
	err := c.do("POST", "byoc/dataplane/create", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

func (c *Client) SuspendBYOCProject(params *SuspendBYOCProjectRequest) (*SuspendBYOCProjectResponse, error) {
	var response zillizResponse[SuspendBYOCProjectResponse]
	err := c.do("POST", "byoc/dataplane/stop", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type SuspendBYOCProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type SuspendBYOCProjectResponse struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

func (c *Client) ResumeBYOCProject(params *ResumeBYOCProjectRequest) (*ResumeBYOCProjectResponse, error) {
	var response zillizResponse[ResumeBYOCProjectResponse]
	err := c.do("POST", "byoc/dataplane/resume", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type ResumeBYOCProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type ResumeBYOCProjectResponse struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

func (c *Client) DeleteBYOCProject(params *DeleteBYOCProjectRequest) (*DeleteBYOCProjectResponse, error) {
	var response zillizResponse[DeleteBYOCProjectResponse]
	err := c.do("DELETE", "byoc/dataplane/delete", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type DeleteBYOCProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}
type DeleteBYOCProjectResponse struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

func (c *Client) DescribeBYOCProject(params *DescribeBYOCProjectRequest) (*GetBYOCProjectResponse, error) {
	var response zillizResponse[GetBYOCProjectResponse]
	err := c.do("GET", "byoc/dataplane/describe?projectId="+params.ProjectId+"&dataPlaneId="+params.DataPlaneID, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

func (c *Client) GetExternalId() (string, error) {
	var response zillizResponse[GetExternalIdResponse]
	err := c.do("GET", "byoc/describe", nil, &response)
	if err != nil {
		return "", err
	}
	return response.Data.ExternalId, err
}

type GetExternalIdResponse struct {
	OrgId          string   `json:"orgId"`
	ExternalId     string   `json:"externalId"`
	ServiceAccount string   `json:"serviceAccount"`
	Clouds         []string `json:"clouds"`
}

type DescribeBYOCProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type CreateBYOCProjectResponse struct {
	JobId       string `json:"jobId"`
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}
type GetBYOCProjectResponse struct {
	AWSConfig struct {
		ARN struct {
			BootstrapRoleArn string `json:"bootstrapRoleArn"`
			EksRoleArn       string `json:"eksRoleArn"`
			StorageRoleArn   string `json:"storageRoleArn"`
		} `json:"arn"`
		BucketID         string   `json:"bucketId"`
		EndpointID       *string  `json:"endpointId"`
		SecurityGroupIDs []string `json:"securityGroupIds"`
		SubnetIDs        []string `json:"subnetIds"`

		// CSE(Client Side Encryption) parameters fields
		AwsCseRoleArn       string `json:"awsCseRoleArn"`
		DefaultAwsCseKeyArn string `json:"defaultAwsCseKeyArn"`
		ExternalID          string `json:"externalId"`

		VPCID string `json:"vpcId"`
	} `json:"awsConfig"`
	VMCombine struct {
		CoreVM  string `json:"coreVm"`
		CoreMax int64  `json:"coreMax"`
		CoreMin int64  `json:"coreMin"`

		FundamentalVM  string `json:"fundamentalVm"`
		FundamentalMin int64  `json:"fundamentalMin"`
		FundamentalMax int64  `json:"fundamentalMax"`

		IndexVM  string `json:"indexVm"`
		IndexMin int64  `json:"indexMin"`
		IndexMax int64  `json:"indexMax"`

		SearchVM  string `json:"searchVm"`
		SearchMax int64  `json:"searchMax"`
		SearchMin int64  `json:"searchMin"`

		AutoScaling bool   `json:"autoScaling"`
		Arch        string `json:"arch"`
	} `json:"vmCombine"`
	CloudID         string      `json:"cloudId"`
	CreateTimeMilli int64       `json:"createTimeMilli"`
	DataPlaneID     string      `json:"dataPlaneId"`
	GCPConfig       interface{} `json:"gcpConfig"`
	LastUpdateMilli int64       `json:"lastUpdateMilli"`
	Message         string      `json:"message"`
	ProjectID       string      `json:"projectId"`
	ProjectName     string      `json:"projectName"`
	RegionID        string      `json:"regionId"`
	Status          int         `json:"status"`
}

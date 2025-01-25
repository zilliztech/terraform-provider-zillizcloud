package client

type CreateBYOCProjectRequest struct {
	AWSParam      AWSParam `json:"awsParam"`
	ProjectName   string   `json:"projectName"`
	RegionID      string   `json:"regionId"`
	CloudID       CloudId  `json:"cloudId"`
	BYOCID        *string  `json:"byocId"`
	FundamentalVM string   `json:"fundamentalVm"`
	SearchVM      string   `json:"searchVm"`
	CoreVM        string   `json:"coreVm"`
	DeployType    int      `json:"deployType"`
}

type AWSParam struct {
	BucketID         string   `json:"bucketId"`
	StorageRoleArn   string   `json:"storageRoleArn"`
	EksRoleArn       string   `json:"eksRoleArn"`
	BootstrapRoleArn string   `json:"bootstrapRoleArn"`
	UserVpcID        string   `json:"userVpcId"`
	SubnetIDs        []string `json:"subnetIds"`
	SecurityGroupIDs []string `json:"securityGroupIds"`
}

func (c *Client) CreateBYOCProject(params *CreateBYOCProjectRequest) (*CreateBYOCProjectResponse, error) {
	var response zillizResponse[CreateBYOCProjectResponse]
	err := c.do("POST", "byoc/dataplane/create", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

func (c *Client) DescribeBYOCProject(params *DescribeBYOCProjectRequest) (*GetBYOCProjectResponse, error) {
	var response zillizResponse[GetBYOCProjectResponse]
	err := c.do("POST", "byoc/dataplane/describe", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
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
		VMCombine        struct {
			CoreVM        string `json:"coreVm"`
			FundamentalVM string `json:"fundamentalVm"`
			SearchVM      string `json:"searchVm"`
		} `json:"vmCombine"`
		VPCID string `json:"vpcId"`
	} `json:"awsConfig"`
	CloudID         string      `json:"cloudId"`
	CreateTimeMilli int64       `json:"createTimeMilli"`
	DataPlaneID     string      `json:"dataPlaneId"`
	GCPConfig       interface{} `json:"gcpConfig"`
	LastUpdateMilli int64       `json:"lastUpdateMilli"`
	Message         string      `json:"message"`
	ProjectID       string      `json:"projectId"`
	RegionID        string      `json:"regionId"`
	Status          int         `json:"status"`
}

package client

type CreateByocOpProjectRequest struct {
	AWSParam   *AWSParam   `json:"awsParam,omitempty"`
	AzureParam *AzureParam `json:"azureParam,omitempty"`

	RegionID string  `json:"regionId"`
	CloudID  CloudId `json:"cloudId"`

	DataPlaneId string `json:"dataPlaneId"`
	ProjectId   string `json:"projectId"`

	DeployType int `json:"deployType"`

	// optional
	ExtConfig *string `json:"extConfig"`
	// no need to set
	FundamentalVM *string `json:"fundamentalVm"`
	SearchVM      *string `json:"searchVm"`
	CoreVM        *string `json:"coreVm"`
}

func (c *Client) CreateByocOpProject(params *CreateByocOpProjectRequest) (*CreateByocOpProjectResponse, error) {
	var response zillizResponse[CreateByocOpProjectResponse]
	err := c.do("POST", "byoc/op/dataplane/create", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type CreateByocOpProjectResponse struct {
	JobId       string `json:"jobId"`
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type DescribeByocOpProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type DescribeByocOpProjectResponse struct {
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
	AzureConfig     AzureParam  `json:"azureConfig"`
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
	OpConfig        *struct {
		Token         string `json:"token"`
		AgentImageUrl string `json:"agentImageUrl"`
	} `json:"vpcOpConfig"`
	Mode int `json:"mode"`
}

func (c *Client) DescribeByocOpProject(params *DescribeByocOpProjectRequest) (*DescribeByocOpProjectResponse, error) {
	var response zillizResponse[DescribeByocOpProjectResponse]
	err := c.do("GET", "byoc/dataplane/describe?projectId="+params.ProjectId+"&dataPlaneId="+params.DataPlaneID, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

func (c *Client) DeleteByocOpProject(params *DeleteByocOpProjectRequest) (*DeleteByocOpProjectResponse, error) {
	var response zillizResponse[DeleteByocOpProjectResponse]
	err := c.do("DELETE", "byoc/dataplane/delete", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type DeleteByocOpProjectRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type DeleteByocOpProjectResponse struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

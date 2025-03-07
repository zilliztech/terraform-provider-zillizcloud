package client

type CreateByocOpProjectSettingsRequest struct {
	ProjectName   string `json:"projectName"`
	CloudId       string `json:"cloudId"`
	RegionId      string `json:"regionId"`
	SearchVm      string `json:"searchVm"`
	FundamentalVm string `json:"fundamentalVm"`
	CoreVm        string `json:"coreVm"`
	DeployType    int    `json:"deployType"`
}

func (c *Client) CreateByocOpProjectSetting(params *CreateByocOpProjectSettingsRequest) (*CreateByocOpProjectSettingResponse, error) {
	var response zillizResponse[CreateByocOpProjectSettingResponse]
	err := c.do("POST", "byoc/op/dataplane/setting", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type CreateByocOpProjectSettingResponse struct {
	JobId       string `json:"jobId"`
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

type DescribeByocOpProjectSettingsRequest struct {
	DataPlaneId string `json:"dataPlaneId"`
	ProjectId   string `json:"projectId"`
}

type GetByocOpProjectSettingsResponse struct {
	ProjectId   string      `json:"projectId"`
	DataPlaneId string      `json:"dataPlaneId"`
	ProjectName string      `json:"projectName"`
	CloudId     string      `json:"cloudId"`
	RegionId    string      `json:"regionId"`
	ByocId      string      `json:"byocId"`
	OpConfig    OpConfig    `json:"opConfig"`
	NodeQuotas  []NodeQuota `json:"nodeQuotas"`
}

type OpConfig struct {
	Token         string `json:"tunnelToken"`
	AgentImageUrl string `json:"agentImageUrl"`
	ExtConfig     string `json:"extConfig"`
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

func (c *Client) DescribeByocOpProjectSettings(params *DescribeByocOpProjectSettingsRequest) (*GetByocOpProjectSettingsResponse, error) {
	var response zillizResponse[GetByocOpProjectSettingsResponse]
	err := c.do("GET", "byoc/op/dataplane/setting?dataPlaneId="+params.DataPlaneId+"&projectId="+params.ProjectId, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

func (c *Client) DeleteByocOpProjectSetting(params *DeleteByocOpProjectSettingRequest) (*DeleteByocOpProjectSettingResponse, error) {
	var response zillizResponse[DeleteByocOpProjectSettingResponse]
	err := c.do("DELETE", "byoc/dataplane/delete", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

type DeleteByocOpProjectSettingRequest struct {
	DataPlaneId string `json:"dataPlaneId"`
	ProjectId   string `json:"projectId"`
}

type DeleteByocOpProjectSettingResponse struct {
	ProjectId   string `json:"projectId"`
	DataPlaneId string `json:"dataPlaneId"`
}

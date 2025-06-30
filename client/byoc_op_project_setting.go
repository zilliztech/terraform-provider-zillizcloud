package client

type CreateByocOpProjectSettingsRequest struct {
	ProjectName string `json:"projectName"`
	CloudId     string `json:"cloudId"`
	RegionId    string `json:"regionId"`

	SearchVm      string `json:"searchVm"`
	FundamentalVm string `json:"fundamentalVm"`
	CoreVm        string `json:"coreVm"`
	IndexVm       string `json:"indexVm"`

	SearchMin      int64 `json:"searchMin"`
	SearchMax      int64 `json:"searchMax"`
	FundamentalMin int64 `json:"fundamentalMin"`
	FundamentalMax int64 `json:"fundamentalMax"`
	CoreMin        int64 `json:"coreMin"`
	CoreMax        int64 `json:"coreMax"`
	IndexMin       int64 `json:"indexMin"`
	IndexMax       int64 `json:"indexMax"`

	AutoScaling bool   `json:"autoScaling"`
	Arch        string `json:"arch"`

	DeployType         int `json:"deployType"`
	PrivateLinkEnabled int `json:"openPl"`
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
	ProjectId          string      `json:"projectId"`
	DataPlaneId        string      `json:"dataPlaneId"`
	ProjectName        string      `json:"projectName"`
	CloudId            string      `json:"cloudId"`
	RegionId           string      `json:"regionId"`
	ByocId             string      `json:"byocId"`
	OpConfig           OpConfig    `json:"opConfig"`
	NodeQuotas         []NodeQuota `json:"nodeQuotas"`
	PrivateLinkEnabled int         `json:"openPl"`
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

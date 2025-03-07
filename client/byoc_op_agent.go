package client

type DescribeByocAgentRequest struct {
	ProjectId   string `json:"projectId"`
	DataPlaneID string `json:"dataPlaneId"`
}

type DescribeByocAgentResponse struct {
	ProjectID string `json:"projectId"`
	RegionID  string `json:"regionId"`
	Status    int    `json:"status"`
}

func (c *Client) DescribeByocAgent(params *DescribeByocAgentRequest) (*DescribeByocAgentResponse, error) {
	var response zillizResponse[DescribeByocAgentResponse]
	err := c.do("GET", "byoc/dataplane/describe?projectId="+params.ProjectId+"&dataPlaneId="+params.DataPlaneID, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, err
}

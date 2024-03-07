package client

type Project struct {
	ProjectId       string `json:"projectId"`
	ProjectName     string `json:"projectName"`
	CreateTimeMilli int64  `json:"createTimeMilli"`
	InstanceCount   int64  `json:"instanceCount"`
}

func (c *Client) ListProjects() ([]Project, error) {
	var response zillizResponse[[]Project]
	err := c.do("GET", "projects", nil, &response)
	return response.Data, err
}

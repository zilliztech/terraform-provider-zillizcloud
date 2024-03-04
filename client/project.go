package client

type Project struct {
	ProjectId   string `json:"projectId"`
	ProjectName string `json:"projectName"`
}

func (c *Client) ListProjects() ([]Project, error) {
	var response zillizResponse[[]Project]
	err := c.do("GET", "projects", nil, &response)
	return response.Data, err
}

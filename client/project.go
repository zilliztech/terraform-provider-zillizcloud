package client

type Project struct {
	ProjectId       string `json:"projectId"`
	ProjectName     string `json:"projectName"`
	CreateTimeMilli int64  `json:"createTimeMilli"`
	InstanceCount   int64  `json:"instanceCount"`
	Plan            string `json:"plan"`
}

type CreateProjectRequest struct {
	ProjectName string `json:"projectName"`
	Plan        string `json:"plan"`
}

type UpgradeProjectPlanRequest struct {
	Plan string `json:"plan"`
}

func (c *Client) CreateProject(params *CreateProjectRequest) (*string, error) {
	var response zillizResponse[string]
	err := c.do("POST", "projects", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ListProjects() ([]Project, error) {
	var response zillizResponse[[]Project]
	err := c.do("GET", "projects", nil, &response)
	return response.Data, err
}

// get project by id
func (c *Client) GetProjectById(projectId string) (*Project, error) {
	var response zillizResponse[Project]
	err := c.do("GET", "projects/"+projectId, nil, &response)
	return &response.Data, err
}

// uprade project plan
func (c *Client) UpgradeProjectPlan(projectId string, plan string) (*string, error) {
	var response zillizResponse[string]
	err := c.do("PATCH", "projects/"+projectId+"/plan", &UpgradeProjectPlanRequest{Plan: plan}, &response)
	return &response.Data, err
}

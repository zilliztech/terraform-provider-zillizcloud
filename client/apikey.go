package client

type ApiKeyProjectAccess struct {
	ProjectId  string   `json:"projectId"`
	Role       string   `json:"role,omitempty"`
	AllCluster *bool    `json:"allCluster,omitempty"`
	ClusterIds []string `json:"clusterIds,omitempty"`
	AllStage   *bool    `json:"allStage,omitempty"`
	StageIds   []string `json:"stageIds,omitempty"`
}

type CreateApiKeyRequest struct {
	Name     string                `json:"name"`
	Role     string                `json:"role"`
	Projects []ApiKeyProjectAccess `json:"projects,omitempty"`
}

type UpdateApiKeyRequest struct {
	Name     string                `json:"name,omitempty"`
	Role     string                `json:"role,omitempty"`
	Projects []ApiKeyProjectAccess `json:"projects,omitempty"`
}

type CreateApiKeyResponse struct {
	ApiKey  string `json:"apiKey"`
	ShortId string `json:"shortId"`
}

type ApiKeyClusterResponse struct {
	ClusterId   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
}

type ApiKeyProjectResponse struct {
	ProjectId   string                  `json:"projectId"`
	ProjectName string                  `json:"projectName"`
	Role        string                  `json:"role"`
	AllCluster  bool                    `json:"allCluster"`
	Clusters    []ApiKeyClusterResponse `json:"clusters"`
}

type ApiKeyResponse struct {
	ApiKeyId     string                  `json:"apiKeyId"`
	ShortId      string                  `json:"shortId"`
	Name         string                  `json:"name"`
	CreatorName  string                  `json:"creatorName"`
	CreatorEmail string                  `json:"creatorEmail"`
	Role         string                  `json:"role"`
	Projects     []ApiKeyProjectResponse `json:"projects"`
	CreateTime   int64                   `json:"createTime"`
}

type ApiKeyListResponse struct {
	ApiKeys []ApiKeyResponse `json:"apiKeys"`
}

func (c *Client) CreateApiKey(req *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	var response zillizResponse[CreateApiKeyResponse]
	err := c.do("POST", "apiKeys", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ListApiKeys() ([]ApiKeyResponse, error) {
	var response zillizResponse[ApiKeyListResponse]
	err := c.do("GET", "apiKeys", nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Data.ApiKeys, nil
}

func (c *Client) GetApiKey(apiKeyId string) (*ApiKeyResponse, error) {
	var response zillizResponse[ApiKeyResponse]
	err := c.do("GET", "apiKeys/"+apiKeyId, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) UpdateApiKey(apiKeyId string, req *UpdateApiKeyRequest) (*ApiKeyResponse, error) {
	var response zillizResponse[ApiKeyResponse]
	err := c.do("PUT", "apiKeys/"+apiKeyId, req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteApiKey(apiKeyId string) error {
	var response zillizResponse[any]
	return c.do("DELETE", "apiKeys/"+apiKeyId, nil, &response)
}

package client

type ApiKeyProjectAccess struct {
	ProjectId  string   `json:"projectId"`
	Role       string   `json:"role,omitempty"`
	AllCluster *bool    `json:"allCluster,omitempty"`
	ClusterIds []string `json:"clusterIds,omitempty"`
	AllVolume  *bool    `json:"allVolume,omitempty"`
	VolumeIds  []string `json:"volumeIds,omitempty"`
}

type CreateApiKeyRequest struct {
	Name     string                `json:"name"`
	OrgRole  string                `json:"orgRole"`
	Projects []ApiKeyProjectAccess `json:"projects,omitempty"`
}

type UpdateApiKeyRequest struct {
	Name     string                `json:"name,omitempty"`
	OrgRole  string                `json:"orgRole,omitempty"`
	Projects []ApiKeyProjectAccess `json:"projects,omitempty"`
}

type CreateApiKeyResponse struct {
	ApiKeyId string `json:"apiKeyId"`
	ApiKey   string `json:"apiKey"`
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
	Name         string                  `json:"name"`
	CreatorName  string                  `json:"creatorName"`
	CreatorEmail string                  `json:"creatorEmail"`
	OrgRole      string                  `json:"orgRole"`
	Projects     []ApiKeyProjectResponse `json:"projects"`
	CreateTime   string                  `json:"createTime"`
}

type ApiKeyListResponse struct {
	ApiKeys     []ApiKeyResponse `json:"apiKeys"`
	Count       int              `json:"count"`
	CurrentPage int              `json:"currentPage"`
	PageSize    int              `json:"pageSize"`
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

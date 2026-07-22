package client

import (
	"errors"
	"fmt"
)

// maxListPages is a safety cap to prevent infinite pagination loops.
const maxListPages = 100

type ApiKeyProjectAccess struct {
	ProjectId  string   `json:"projectId"`
	Role       string   `json:"role,omitempty"`
	AllCluster *bool    `json:"allCluster,omitempty"`
	ClusterIds []string `json:"clusterIds,omitempty"`
	AllVolume  *bool    `json:"allVolume,omitempty"`
	VolumeIds  []string `json:"volumeIds,omitempty"`
}

type CreateApiKeyRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"` // omitempty OK on create
	OrgRole     string                `json:"orgRole"`
	Projects    []ApiKeyProjectAccess `json:"projects,omitempty"`
}

// UpdateApiKeyRequest sends the full desired state (FC5).
// NO omitempty on scalar fields so clearing description is expressible.
type UpdateApiKeyRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	OrgRole     string                `json:"orgRole"`
	Projects    []ApiKeyProjectAccess `json:"projects"`
}

type CreateApiKeyResponse struct {
	ApiKeyId string `json:"apiKeyId"`
	ApiKey   string `json:"apiKey"`
}

type ApiKeyClusterResponse struct {
	ClusterId   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
}

type ApiKeyVolumeResponse struct {
	VolumeId   string `json:"volumeId"`
	VolumeName string `json:"volumeName"`
}

type ApiKeyProjectResponse struct {
	ProjectId   string                  `json:"projectId"`
	ProjectName string                  `json:"projectName"`
	Role        string                  `json:"role"`
	AllCluster  bool                    `json:"allCluster"`
	Clusters    []ApiKeyClusterResponse `json:"clusters"`
	AllVolume   bool                    `json:"allVolume"`
	Volumes     []ApiKeyVolumeResponse  `json:"volumes"`
}

type ApiKeyResponse struct {
	ApiKeyId    string                  `json:"apiKeyId"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"` // "" when unset
	CreatorName string                  `json:"creatorName"`
	CreatedBy   string                  `json:"createdBy"`
	OrgRole     string                  `json:"orgRole"`
	Projects    []ApiKeyProjectResponse `json:"projects"`
	CreateTime  string                  `json:"createTime"`
}

type ApiKeyListResponse struct {
	ApiKeys     []ApiKeyResponse `json:"apiKeys"`
	Count       int              `json:"count"`
	CurrentPage int              `json:"currentPage"`
	PageSize    int              `json:"pageSize"`
}

// IsApiKeyNotFound returns true if the error represents a not-found condition.
// FC4: accepts both 404 (current gateway) and 80001 (legacy internal code).
func IsApiKeyNotFound(err error) bool {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == 404 || apiErr.Code == 80001
	}
	return false
}

// wrapApiKeyPermissionError wraps the generic 96041 error with a helpful message (FC7).
func wrapApiKeyPermissionError(err error) error {
	var apiErr *Error
	if errors.As(err, &apiErr) && apiErr.Code == 96041 {
		return fmt.Errorf(
			"API key management requires an Org Owner API key. "+
				"The configured key appears to be project-scoped and cannot manage API keys. "+
				"See the provider documentation for the bootstrap pattern. Original error: %w", err)
	}
	return err
}

func (c *Client) CreateApiKey(req *CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	var response zillizResponse[CreateApiKeyResponse]
	err := c.do("POST", "apiKeys", req, &response)
	if err != nil {
		return nil, wrapApiKeyPermissionError(err)
	}
	return &response.Data, nil
}

// ListApiKeys returns all Customized API keys with full pagination (FC6).
func (c *Client) ListApiKeys() ([]ApiKeyResponse, error) {
	var all []ApiKeyResponse
	for page := 1; page <= maxListPages; page++ {
		var resp zillizResponse[ApiKeyListResponse]
		err := c.do("GET", fmt.Sprintf("apiKeys?currentPage=%d&pageSize=100", page), nil, &resp)
		if err != nil {
			return nil, wrapApiKeyPermissionError(err)
		}
		all = append(all, resp.Data.ApiKeys...)
		if len(resp.Data.ApiKeys) == 0 || len(resp.Data.ApiKeys) < 100 || len(all) >= resp.Data.Count {
			break
		}
	}
	return all, nil
}

func (c *Client) GetApiKey(apiKeyId string) (*ApiKeyResponse, error) {
	var response zillizResponse[ApiKeyResponse]
	err := c.do("GET", "apiKeys/"+apiKeyId, nil, &response)
	if err != nil {
		return nil, wrapApiKeyPermissionError(err)
	}
	return &response.Data, nil
}

func (c *Client) UpdateApiKey(apiKeyId string, req *UpdateApiKeyRequest) (*ApiKeyResponse, error) {
	var response zillizResponse[ApiKeyResponse]
	err := c.do("PUT", "apiKeys/"+apiKeyId, req, &response)
	if err != nil {
		return nil, wrapApiKeyPermissionError(err)
	}
	return &response.Data, nil
}

func (c *Client) DeleteApiKey(apiKeyId string) error {
	var response zillizResponse[any]
	err := c.do("DELETE", "apiKeys/"+apiKeyId, nil, &response)
	if err != nil {
		return wrapApiKeyPermissionError(err)
	}
	return nil
}

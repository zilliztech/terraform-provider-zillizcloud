package client

import (
	"fmt"
	"net/url"
)

// EndpointService represents an available private link endpoint service.
type EndpointService struct {
	RegionId          string `json:"regionId"`
	CloudId           string `json:"cloudId"`
	EndpointService   string `json:"endpointService"`
	WhitelistRequired bool   `json:"whitelistRequired"`
}

// Endpoint represents a VPC private link endpoint under a project.
type Endpoint struct {
	RegionId              string  `json:"regionId"`
	CloudId               string  `json:"cloudId"`
	EndpointService       string  `json:"endpointService"`
	EndpointServiceStatus string  `json:"endpointServiceStatus"`
	EndpointId            string  `json:"endpointId"`
	EndpointStatus        string  `json:"endpointStatus"`
	GcpProjectId          *string `json:"gcpProjectId"`
}

// listEndpointServicesData is the inner payload for GET /v2/endpointServices.
type listEndpointServicesData struct {
	EndpointServices []EndpointService `json:"endpointServices"`
	zillizPage
}

// listEndpointsData is the inner payload for GET /v2/projects/{projectId}/endpoints.
type listEndpointsData struct {
	Endpoints []Endpoint `json:"endpoints"`
	zillizPage
}

// CreateEndpointRequest is the body for POST /v2/projects/{projectId}/endpoints.
type CreateEndpointRequest struct {
	RegionId     string `json:"regionId"`
	EndpointId   string `json:"endpointId"`
	GcpProjectId string `json:"gcpProjectId,omitempty"`
}

// CreateEndpointResponse is the response payload for POST /v2/projects/{projectId}/endpoints.
type CreateEndpointResponse struct {
	EndpointId string `json:"endpointId"`
	RegionId   string `json:"regionId"`
}

// AddEndpointWhitelistRequest is the body for POST /v2/projects/{projectId}/endpointWhitelist.
type AddEndpointWhitelistRequest struct {
	RegionId    string `json:"regionId"`
	OuterUserId string `json:"outerUserId"`
}

// ListEndpointServices lists available private link endpoint services for a region.
func (c *Client) ListEndpointServices(regionId string, currentPage, pageSize int) ([]EndpointService, zillizPage, error) {
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := url.Values{}
	q.Set("regionId", regionId)
	q.Set("currentPage", fmt.Sprintf("%d", currentPage))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))

	var response zillizResponse[listEndpointServicesData]
	err := c.do("GET", "endpointServices?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.EndpointServices, response.Data.zillizPage, nil
}

// ListEndpoints lists private link endpoints under a project.
func (c *Client) ListEndpoints(projectId string, currentPage, pageSize int) ([]Endpoint, zillizPage, error) {
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	q := url.Values{}
	q.Set("currentPage", fmt.Sprintf("%d", currentPage))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))

	var response zillizResponse[listEndpointsData]
	err := c.do("GET", "projects/"+projectId+"/endpoints?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}
	return response.Data.Endpoints, response.Data.zillizPage, nil
}

// CreateEndpoint creates a private link endpoint under a project.
func (c *Client) CreateEndpoint(projectId string, req *CreateEndpointRequest) (*CreateEndpointResponse, error) {
	var response zillizResponse[CreateEndpointResponse]
	err := c.do("POST", "projects/"+projectId+"/endpoints", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

// DeleteEndpoint deletes a private link endpoint. regionId is required; gcpProjectId is required only for GCP regions.
func (c *Client) DeleteEndpoint(projectId, endpointId, regionId string, gcpProjectId *string) error {
	q := url.Values{}
	q.Set("regionId", regionId)
	if gcpProjectId != nil && *gcpProjectId != "" {
		q.Set("gcpProjectId", *gcpProjectId)
	}
	var response zillizResponse[map[string]any]
	return c.do("DELETE", "projects/"+projectId+"/endpoints/"+endpointId+"?"+q.Encode(), nil, &response)
}

// AddEndpointWhitelist adds an external cloud account to the endpoint whitelist.
func (c *Client) AddEndpointWhitelist(projectId string, req *AddEndpointWhitelistRequest) error {
	var response zillizResponse[string]
	return c.do("POST", "projects/"+projectId+"/endpointWhitelist", req, &response)
}

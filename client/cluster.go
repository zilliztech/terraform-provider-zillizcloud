package client

import "strings"

type Plan string

var (
	FreePlan       Plan = "Free"
	ServerlessPlan Plan = "Serverless"
	StandardPlan   Plan = "Standard"
	EnterprisePlan Plan = "Enterprise"
)

type ModifyClusterParams struct {
	CuSize int `json:"cuSize"`
}

type ModifyClusterResponse struct {
	ClusterId string `json:"clusterId"`
}

func (c *Client) ModifyCluster(clusterId string, params *ModifyClusterParams) (*string, error) {
	var response zillizResponse[ModifyClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/modify", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

type DropClusterResponse struct {
	ClusterId string `json:"clusterId"`
}

func (c *Client) DropCluster(clusterId string) (*string, error) {
	var response zillizResponse[DropClusterResponse]
	err := c.do("DELETE", "clusters/"+clusterId+"/drop", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

type Clusters struct {
	zillizPage
	Clusters []Cluster `json:"clusters"`
}

type Cluster struct {
	ClusterId          string `json:"clusterId"`
	ClusterName        string `json:"clusterName"`
	Description        string `json:"description"`
	RegionId           string `json:"regionId"`
	ClusterType        string `json:"clusterType"`
	CuType             string `json:"cuType"`
	Plan               Plan   `json:"plan"`
	CuSize             int64  `json:"cuSize"`
	Status             string `json:"status"`
	ConnectAddress     string `json:"connectAddress"`
	PrivateLinkAddress string `json:"privateLinkAddress"`
	CreateTime         string `json:"createTime"`
	ProjectId          string `json:"projectId"`
}

func (c *Client) ListClusters() (Clusters, error) {
	var clusters zillizResponse[Clusters]
	err := c.do("GET", "clusters", nil, &clusters)
	return clusters.Data, err
}

func (c *Client) DescribeCluster(clusterId string) (Cluster, error) {
	var response zillizResponse[Cluster]
	err := c.do("GET", "clusters/"+clusterId, nil, &response)
	if err != nil {
		return Cluster{}, err
	}
	cluster := response.Data

	// TODO: remove this once we have a better way to determine the plan
	// in03- is a free cluster
	if strings.HasPrefix(cluster.ClusterId, "in03-") {
		cluster.Plan = FreePlan
	}

	return cluster, err
}

type CreateClusterParams struct {
	Plan        Plan   `json:"plan"`
	ClusterName string `json:"clusterName"`
	CUSize      int    `json:"cuSize"`
	CUType      string `json:"cuType"`
	ProjectId   string `json:"projectId"`
	RegionId    string `json:"regionId"`
}

type CreateServerlessClusterParams struct {
	ClusterName string `json:"clusterName"`
	ProjectId   string `json:"projectId"`
	RegionId    string `json:"regionId"`
}

type CreateClusterResponse struct {
	ClusterId string `json:"clusterId"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Prompt    string `json:"prompt"`
}

func (c *Client) CreateCluster(params CreateClusterParams) (*CreateClusterResponse, error) {
	if params.RegionId == "" && c.RegionId == "" {
		return nil, errRegionIdRequired
	}
	var clusterResponse zillizResponse[CreateClusterResponse]
	err := c.do("POST", "clusters/create", params, &clusterResponse)
	return &clusterResponse.Data, err
}

func (c *Client) CreateDedicatedCluster(params CreateClusterParams) (*CreateClusterResponse, error) {
	if params.RegionId == "" && c.RegionId == "" {
		return nil, errRegionIdRequired
	}
	var clusterResponse zillizResponse[CreateClusterResponse]
	err := c.do("POST", "clusters/createDedicated", params, &clusterResponse)
	return &clusterResponse.Data, err
}

func (c *Client) CreateFreeCluster(params CreateServerlessClusterParams) (*CreateClusterResponse, error) {
	if params.RegionId == "" && c.RegionId == "" {
		return nil, errRegionIdRequired
	}
	var clusterResponse zillizResponse[CreateClusterResponse]
	err := c.do("POST", "clusters/createFree", params, &clusterResponse)
	return &clusterResponse.Data, err
}

func (c *Client) CreateServerlessCluster(params CreateServerlessClusterParams) (*CreateClusterResponse, error) {
	var clusterResponse zillizResponse[CreateClusterResponse]
	err := c.do("POST", "clusters/createServerless", params, &clusterResponse)
	return &clusterResponse.Data, err
}

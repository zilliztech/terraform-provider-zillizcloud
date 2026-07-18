package client

import "fmt"

type GlobalClusterMemberParams struct {
	ClusterName string `json:"clusterName"`
	RegionId    string `json:"regionId"`
	Replica     *int   `json:"replica,omitempty"`
}

type CreateGlobalClusterParams struct {
	GlobalClusterName string                      `json:"globalClusterName"`
	ProjectId         string                      `json:"projectId"`
	CuType            string                      `json:"cuType"`
	CuSize            int                         `json:"cuSize,omitempty"`
	Autoscaling       *AutoscalingConfig          `json:"autoscaling,omitempty"`
	PrimaryCluster    GlobalClusterMemberParams   `json:"primaryCluster"`
	SecondaryClusters []GlobalClusterMemberParams `json:"secondaryClusters"`
}

type CreateGlobalClusterResponse struct {
	GlobalClusterId string `json:"globalClusterId"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	JobId           string `json:"jobId"`
}

type GlobalCluster struct {
	GlobalClusterId   string                `json:"globalClusterId"`
	GlobalClusterName string                `json:"globalClusterName"`
	ProjectId         string                `json:"projectId"`
	RegionIds         []string              `json:"regionIds"`
	CuType            string                `json:"cuType"`
	CuSize            int64                 `json:"cuSize"`
	Autoscaling       AutoscalingConfig     `json:"autoscaling"`
	ConnectAddress    string                `json:"connectAddress"`
	CreateTime        string                `json:"createTime"`
	Clusters          []GlobalClusterMember `json:"clusters"`
}

type GlobalClusterMember struct {
	ClusterId   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	RegionId    string `json:"regionId"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Replica     int64  `json:"replica,omitempty"`
}

type ModifyGlobalClusterCUParams struct {
	CuSize int `json:"cuSize"`
}

type AddSecondaryClustersParams struct {
	SecondaryClusters []GlobalClusterMemberParams `json:"secondaryClusters"`
}

type GlobalClusterJobResponse struct {
	JobId string `json:"jobId"`
}

type RemoveGlobalEndpointResponse struct {
	OldGlobalClusterId string `json:"oldGlobalClusterId"`
}

type DeleteClusterResponse struct {
	GlobalClusterId string `json:"globalClusterId"`
	ClusterId       string `json:"clusterId"`
	Prompt          string `json:"prompt"`
}

func (c *Client) CreateGlobalCluster(params *CreateGlobalClusterParams) (*CreateGlobalClusterResponse, error) {
	var response zillizResponse[CreateGlobalClusterResponse]
	err := c.do("POST", "globalClusters/create", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DescribeGlobalCluster(globalClusterId string) (*GlobalCluster, error) {
	if globalClusterId == "" {
		return nil, fmt.Errorf("globalClusterId is required")
	}

	var response zillizResponse[GlobalCluster]
	err := c.do("GET", "globalClusters/"+globalClusterId, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ModifyGlobalClusterCU(globalClusterId string, params *ModifyGlobalClusterCUParams) (*GlobalClusterJobResponse, error) {
	if globalClusterId == "" {
		return nil, fmt.Errorf("globalClusterId is required")
	}

	var response zillizResponse[GlobalClusterJobResponse]
	err := c.do("POST", "globalClusters/"+globalClusterId+"/modifyCU", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) RemoveGlobalEndpoint(globalClusterId string) (*RemoveGlobalEndpointResponse, error) {
	if globalClusterId == "" {
		return nil, fmt.Errorf("globalClusterId is required")
	}

	var response zillizResponse[RemoveGlobalEndpointResponse]
	err := c.do("POST", "globalClusters/"+globalClusterId+"/removeGlobalEndpoint", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) AddSecondaryClusters(globalClusterId string, params *AddSecondaryClustersParams) (*GlobalClusterJobResponse, error) {
	if globalClusterId == "" {
		return nil, fmt.Errorf("globalClusterId is required")
	}

	var response zillizResponse[GlobalClusterJobResponse]
	err := c.do("POST", "globalClusters/"+globalClusterId+"/secondaryClusters", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteCluster(globalClusterId string, clusterId string) (*DeleteClusterResponse, error) {
	if globalClusterId == "" {
		return nil, fmt.Errorf("globalClusterId is required")
	}
	if clusterId == "" {
		return nil, fmt.Errorf("clusterId is required")
	}

	var response zillizResponse[DeleteClusterResponse]
	err := c.do("DELETE", "globalClusters/"+globalClusterId+"/clusters/"+clusterId, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

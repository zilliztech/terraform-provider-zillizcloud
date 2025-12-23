package client

import (
	"fmt"
	"strings"
)

type Plan string

var (
	FreePlan             Plan = "Free"
	ServerlessPlan       Plan = "Serverless"
	StandardPlan         Plan = "Standard"
	EnterprisePlan       Plan = "Enterprise"
	BusinessCriticalPlan Plan = "BusinessCritical"
	BuiltInPlan          Plan = "" // one can leave plan empty for BYOC env
)

type ModifyClusterParams struct {
	CuSize int `json:"cuSize"`
}

type ClusterResponse struct {
	ClusterId string `json:"clusterId"`
}

type UpsertSecurityGroupsParams struct {
	Ids []string `json:"ids"`
}

// upsert security groups
func (c *Client) UpsertSecurityGroups(clusterId string, params *UpsertSecurityGroupsParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("PUT", "clusters/"+clusterId+"/securityGroups", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, nil
}

type GetSecurityGroupsResponse struct {
	Ids []string `json:"ids"`
}

// get security groups
func (c *Client) GetSecurityGroups(clusterId string) ([]string, error) {
	var response zillizResponse[GetSecurityGroupsResponse]
	err := c.do("GET", "clusters/"+clusterId+"/securityGroups", nil, &response)
	if err != nil {
		return nil, err
	}
	return response.Data.Ids, nil
}

// suspend cluster

func (c *Client) SuspendCluster(clusterId string) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/suspend", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

func (c *Client) ResumeCluster(clusterId string) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/resume", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

func (c *Client) ModifyCluster(clusterId string, params *ModifyClusterParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/modify", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

// modify cluster cu size
type ModifyClusterAutoscalingParams struct {
	Autoscaling struct {
		CU struct {
			Min *int `json:"min,omitempty"`
			Max *int `json:"max,omitempty"`
		} `json:"cu"`
	} `json:"autoscaling"`
}

func (c *Client) ModifyClusterAutoscaling(clusterId string, params *ModifyClusterAutoscalingParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/modify", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

// add or remove cu
type ModifyPropertiesParams struct {
	ClusterName string `json:"clusterName"`
}

func (c *Client) ModifyClusterProperties(clusterId string, params *ModifyPropertiesParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/modifyProperties", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

// modifyReplica.
type ModifyReplicaParams struct {
	Replica int `json:"replica"`
}

func (c *Client) ModifyReplica(clusterId string, params *ModifyReplicaParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("POST", "clusters/"+clusterId+"/modifyReplica", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

// update labels
type UpdateLabelsParams struct {
	Labels map[string]string `json:"labels"`
}

func (c *Client) UpdateLabels(clusterId string, params *UpdateLabelsParams) (*string, error) {
	var response zillizResponse[ClusterResponse]
	err := c.do("PUT", "clusters/"+clusterId+"/labels", params, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data.ClusterId, err
}

func (c *Client) GetLabels(clusterId string) (map[string]string, error) {
	var response zillizResponse[struct {
		Labels map[string]string `json:"labels"`
	}]
	err := c.do("GET", "clusters/"+clusterId+"/labels", nil, &response)
	return response.Data.Labels, err
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
	ClusterId          string            `json:"clusterId"`
	ClusterName        string            `json:"clusterName"`
	Description        string            `json:"description"`
	RegionId           string            `json:"regionId"`
	ClusterType        string            `json:"clusterType"`
	CuType             string            `json:"cuType"`
	Plan               Plan              `json:"plan"`
	CuSize             int64             `json:"cuSize"`
	Status             string            `json:"status"`
	ConnectAddress     string            `json:"connectAddress"`
	PrivateLinkAddress string            `json:"privateLinkAddress"`
	CreateTime         string            `json:"createTime"`
	ProjectId          string            `json:"projectId"`
	Labels             map[string]string `json:"labels,omitempty"`
	Replica            int64             `json:"replica,omitempty"`
	Autoscaling        struct {
		CU struct {
			Min *int `json:"min,omitempty"`
			Max *int `json:"max,omitempty"`
		} `json:"cu"`
	} `json:"autoscaling"`
}

type Autoscaling struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func (c *Client) ListClusters() (Clusters, error) {
	var clusters zillizResponse[Clusters]
	err := c.do("GET", "clusters", nil, &clusters)
	return clusters.Data, err
}

func (c *Client) DescribeCluster(clusterId string) (Cluster, error) {
	if clusterId == "" {
		return Cluster{}, fmt.Errorf("clusterId is required")
	}
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

	switch cluster.Status {
	case "STOPPING":
		cluster.Status = "SUSPENDING"
	case "STOPPED":
		cluster.Status = "SUSPENDED"
	}

	return cluster, err
}

type CreateClusterParams struct {
	Plan        *string           `json:"plan,omitempty"`
	ClusterName string            `json:"clusterName"`
	CUSize      int               `json:"cuSize"`
	CUType      string            `json:"cuType"`
	ProjectId   string            `json:"projectId"`
	RegionId    string            `json:"regionId"`
	Labels      map[string]string `json:"labels,omitempty"`
	BucketInfo  *BucketInfo       `json:"bucketInfo,omitempty"`
}
type BucketInfo struct {
	BucketName string  `json:"bucketName"`
	Prefix     *string `json:"prefix,omitempty"`
}

type CreateServerlessClusterParams struct {
	ClusterName string            `json:"clusterName"`
	ProjectId   string            `json:"projectId"`
	RegionId    string            `json:"regionId"`
	Labels      map[string]string `json:"labels,omitempty"`
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

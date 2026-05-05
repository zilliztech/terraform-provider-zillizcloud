package client

import (
	"fmt"
	"net/url"
)

type CreateOnDemandClusterRequest struct {
	ProjectID            string `json:"projectId"`
	RegionID             string `json:"regionId"`
	CUSize               int    `json:"cuSize"`
	AutoSuspend          *int   `json:"autoSuspend,omitempty"`
	MaxQueryNodeCU       *int   `json:"maxQueryNodeCU,omitempty"`
	MaxQueryNodeReplicas *int   `json:"maxQueryNodeReplicas,omitempty"`
	ClusterName          string `json:"clusterName"`
}

type ActionClusterResponse struct {
	ClusterID string `json:"clusterId"`
	Prompt    string `json:"prompt"`
}

type CreateOnDemandClusterResponse = CreateClusterResponse

type QueryCluster struct {
	ClusterID     string  `json:"clusterId"`
	ClusterName   *string `json:"clusterName,omitempty"`
	RegionID      string  `json:"regionId"`
	CUSize        *int    `json:"cuSize,omitempty"`
	Replicas      *int    `json:"replicas,omitempty"`
	ReadyReplicas *int    `json:"readyReplicas,omitempty"`
	Status        string  `json:"status"`
	Endpoint      *string `json:"endpoint,omitempty"`
	PrivateLink   *string `json:"privateLink,omitempty"`
	CreatedBy     *string `json:"createdBy,omitempty"`
	CreateTime    *int64  `json:"createTime,omitempty"`
	AutoSuspend   *int    `json:"autoSuspend,omitempty"`
	TTLSeconds    *int    `json:"ttlSeconds,omitempty"`
}

type ListOnDemandClustersResponse struct {
	OnDemandClusters []QueryCluster `json:"onDemandClusters"`
	Count            int            `json:"count"`
}

func (c *Client) CreateOnDemandCluster(req *CreateOnDemandClusterRequest) (*CreateOnDemandClusterResponse, error) {
	var response zillizResponse[CreateOnDemandClusterResponse]
	err := c.do("POST", "clusters/createOnDemandCluster", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DescribeOnDemandCluster(clusterID string) (*QueryCluster, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterId is required")
	}

	var response zillizResponse[QueryCluster]
	err := c.do("GET", "clusters/onDemandClusters/"+clusterID, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) ListOnDemandClusters(projectID, regionID string) (*ListOnDemandClustersResponse, error) {
	q := url.Values{}
	q.Set("projectId", projectID)
	q.Set("regionId", regionID)

	var response zillizResponse[ListOnDemandClustersResponse]
	err := c.do("GET", "clusters/onDemandClusters?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

func (c *Client) DeleteOnDemandCluster(clusterID string) (*ActionClusterResponse, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterId is required")
	}

	var response zillizResponse[ActionClusterResponse]
	err := c.do("DELETE", "clusters/onDemandClusters/"+clusterID, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

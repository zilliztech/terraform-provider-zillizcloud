package client

import (
	"context"
	"net/http"
	"net/url"
)

type ClusterVersionSelection struct {
	MilvusVersion string `json:"milvusVersion"`
	Channel       string `json:"channel"`
}

type RegionalClusterVersions struct {
	RegionID string                    `json:"regionId"`
	Versions []ClusterVersionSelection `json:"versions"`
}

func (c *Client) ListClusterVersions(ctx context.Context, regionID, backupID string) ([]RegionalClusterVersions, error) {
	query := url.Values{}
	if regionID != "" {
		query.Set("regionId", regionID)
	}
	if backupID != "" {
		query.Set("backupId", backupID)
	}
	var response zillizResponse[[]RegionalClusterVersions]
	err := c.doContext(ctx, http.MethodGet, "clusters/versions?"+query.Encode(), nil, &response)
	return response.Data, err
}

type RestoreClusterParams struct {
	ClusterName          string                   `json:"clusterName"`
	ProjectID            string                   `json:"projectId"`
	CUSize               int64                    `json:"cuSize"`
	Description          string                   `json:"description,omitempty"`
	CollectionStatus     string                   `json:"collectionStatus,omitempty"`
	RestoreVersionPolicy string                   `json:"restoreVersionPolicy,omitempty"`
	Version              *ClusterVersionSelection `json:"version,omitempty"`
}

type RestoreClusterResponse struct {
	ClusterID string `json:"clusterId"`
	JobID     string `json:"jobId"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (c *Client) RestoreCluster(ctx context.Context, sourceClusterID, backupID string, params RestoreClusterParams) (RestoreClusterResponse, error) {
	path := "clusters/" + url.PathEscape(sourceClusterID) + "/backups/" + url.PathEscape(backupID) + "/restoreCluster"
	var response zillizResponse[RestoreClusterResponse]
	err := c.doContext(ctx, http.MethodPost, path, params, &response)
	return response.Data, err
}

type RestoreJob struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (c *Client) DescribeRestoreJob(ctx context.Context, jobID string) (RestoreJob, error) {
	var response zillizResponse[RestoreJob]
	err := c.doContext(ctx, http.MethodGet, "jobs/"+url.PathEscape(jobID), nil, &response)
	return response.Data, err
}

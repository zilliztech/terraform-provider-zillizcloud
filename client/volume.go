package client

import (
	"fmt"
	"net/url"
)

// CreateVolumeRequest is the body for POST /v2/volumes/create.
type CreateVolumeRequest struct {
	ProjectID            string `json:"projectId"`
	RegionID             string `json:"regionId"`
	VolumeName           string `json:"volumeName"`
	Type                 string `json:"type,omitempty"`
	StorageIntegrationID string `json:"storageIntegrationId,omitempty"`
	Path                 string `json:"path,omitempty"`
}

// CreateVolumeData is the response payload for POST /v2/volumes/create.
type CreateVolumeData struct {
	VolumeName string `json:"volumeName"`
}

// VolumeSummary represents a volume returned by list/search APIs.
type VolumeSummary struct {
	VolumeName string `json:"volumeName"`
	Type       string `json:"type"`
}

// ListVolumesData is the inner payload for GET /v2/volumes.
type ListVolumesData struct {
	Volumes []VolumeSummary `json:"volumes"`
	zillizPage
}

// DescribeVolumeData is the response payload for GET /v2/volumes/{volumeName}.
type DescribeVolumeData struct {
	VolumeName           string `json:"volumeName"`
	Type                 string `json:"type"`
	RegionID             string `json:"regionId"`
	StorageIntegrationID string `json:"storageIntegrationId,omitempty"`
	Path                 string `json:"path,omitempty"`
	Status               string `json:"status"`
	CreateTime           string `json:"createTime"`
}

// DeleteVolumeData is the response payload for DELETE /v2/volumes/{volumeName}.
type DeleteVolumeData struct {
	VolumeName string `json:"volumeName"`
}

// ListVolumes lists volumes under a project.
func (c *Client) ListVolumes(projectId string, currentPage, pageSize int, volumeType string) ([]VolumeSummary, zillizPage, error) {
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	q := url.Values{}
	q.Set("projectId", projectId)
	q.Set("currentPage", fmt.Sprintf("%d", currentPage))
	q.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if volumeType != "" {
		q.Set("type", volumeType)
	}

	var response zillizResponse[ListVolumesData]
	err := c.do("GET", "volumes?"+q.Encode(), nil, &response)
	if err != nil {
		return nil, zillizPage{}, err
	}

	return response.Data.Volumes, response.Data.zillizPage, nil
}

// DescribeVolume describes a volume by name.
func (c *Client) DescribeVolume(volumeName string) (*DescribeVolumeData, error) {
	var response zillizResponse[DescribeVolumeData]
	err := c.do("GET", "volumes/"+url.PathEscape(volumeName), nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

// CreateVolume creates a volume.
func (c *Client) CreateVolume(req *CreateVolumeRequest) (*CreateVolumeData, error) {
	var response zillizResponse[CreateVolumeData]
	err := c.do("POST", "volumes/create", req, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

// DeleteVolume deletes a volume by name.
func (c *Client) DeleteVolume(volumeName string) (*DeleteVolumeData, error) {
	var response zillizResponse[DeleteVolumeData]
	err := c.do("DELETE", "volumes/"+url.PathEscape(volumeName), nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.Data, nil
}

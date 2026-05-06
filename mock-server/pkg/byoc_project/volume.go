package byoc_project

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateVolumeRequest struct {
	ProjectID            string `json:"projectId"`
	RegionID             string `json:"regionId"`
	VolumeName           string `json:"volumeName"`
	Type                 string `json:"type,omitempty"`
	StorageIntegrationID string `json:"storageIntegrationId,omitempty"`
	Path                 string `json:"path,omitempty"`
}

type Volume struct {
	ProjectID            string `json:"projectId"`
	RegionID             string `json:"regionId"`
	VolumeName           string `json:"volumeName"`
	Type                 string `json:"type"`
	StorageIntegrationID string `json:"storageIntegrationId,omitempty"`
	Path                 string `json:"path,omitempty"`
	Status               string `json:"status"`
	CreateTime           string `json:"createTime"`
}

type VolumeSummary struct {
	ProjectID            string `json:"projectId,omitempty"`
	RegionID             string `json:"regionId,omitempty"`
	VolumeName           string `json:"volumeName"`
	Type                 string `json:"type"`
	StorageIntegrationID string `json:"storageIntegrationId,omitempty"`
	Path                 string `json:"path,omitempty"`
	Status               string `json:"status,omitempty"`
	CreateTime           string `json:"createTime,omitempty"`
}

type ListVolumesResponse struct {
	Volumes     []VolumeSummary `json:"volumes"`
	Count       int             `json:"count"`
	CurrentPage int             `json:"currentPage"`
	PageSize    int             `json:"pageSize"`
}

type VolumeNameResponse struct {
	VolumeName string `json:"volumeName"`
}

func CreateVolume(c *gin.Context) {
	var request CreateVolumeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	if request.ProjectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "projectId is required"})
		return
	}
	if request.RegionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "regionId is required"})
		return
	}
	if request.VolumeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "volumeName is required"})
		return
	}

	volumeType := request.Type
	if volumeType == "" {
		volumeType = "MANAGED"
	}
	if volumeType == "EXTERNAL" && request.StorageIntegrationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "storageIntegrationId is required for external volumes"})
		return
	}

	volume := &Volume{
		ProjectID:            request.ProjectID,
		RegionID:             request.RegionID,
		VolumeName:           request.VolumeName,
		Type:                 volumeType,
		StorageIntegrationID: request.StorageIntegrationID,
		Path:                 request.Path,
		Status:               "Available",
		CreateTime:           time.Now().UTC().Format(time.RFC3339),
	}
	volumeStore.Set(volume.VolumeName, volume)

	c.JSON(http.StatusOK, Response[VolumeNameResponse]{
		Code: 0,
		Data: VolumeNameResponse{VolumeName: volume.VolumeName},
	})
}

func ListVolumes(c *gin.Context) {
	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "projectId is required"})
		return
	}

	currentPage := parsePositiveInt(c.Query("currentPage"), 1)
	pageSize := parsePositiveInt(c.Query("pageSize"), 10)
	volumeType := c.Query("type")

	var filtered []VolumeSummary
	for _, volume := range volumeStore.GetAll() {
		if volume.ProjectID != projectID {
			continue
		}
		if volumeType != "" && volume.Type != volumeType {
			continue
		}
		filtered = append(filtered, volume.ToSummary())
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].VolumeName < filtered[j].VolumeName
	})

	count := len(filtered)
	start := (currentPage - 1) * pageSize
	if start > count {
		start = count
	}
	end := start + pageSize
	if end > count {
		end = count
	}

	c.JSON(http.StatusOK, Response[ListVolumesResponse]{
		Code: 0,
		Data: ListVolumesResponse{
			Volumes:     filtered[start:end],
			Count:       count,
			CurrentPage: currentPage,
			PageSize:    pageSize,
		},
	})
}

func DescribeVolume(c *gin.Context) {
	volumeName := c.Param("volumeName")
	if volumeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "volumeName is required"})
		return
	}

	volume := volumeStore.Get(volumeName)
	if volume == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": fmt.Sprintf("Volume with name %s not found", volumeName),
		})
		return
	}

	c.JSON(http.StatusOK, Response[*Volume]{
		Code: 0,
		Data: volume,
	})
}

func DeleteVolume(c *gin.Context) {
	volumeName := c.Param("volumeName")
	if volumeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "volumeName is required"})
		return
	}

	volume := volumeStore.Get(volumeName)
	if volume == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": fmt.Sprintf("Volume with name %s not found", volumeName),
		})
		return
	}

	volumeStore.Delete(volumeName)

	c.JSON(http.StatusOK, Response[VolumeNameResponse]{
		Code: 0,
		Data: VolumeNameResponse{VolumeName: volumeName},
	})
}

func (v *Volume) ToSummary() VolumeSummary {
	return VolumeSummary{
		ProjectID:            v.ProjectID,
		RegionID:             v.RegionID,
		VolumeName:           v.VolumeName,
		Type:                 v.Type,
		StorageIntegrationID: v.StorageIntegrationID,
		Path:                 v.Path,
		Status:               v.Status,
		CreateTime:           v.CreateTime,
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

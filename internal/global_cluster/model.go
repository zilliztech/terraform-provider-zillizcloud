package global_cluster

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type GlobalClusterMemberRole string

const (
	GlobalClusterMemberRolePrimary   GlobalClusterMemberRole = "PRIMARY"
	GlobalClusterMemberRoleSecondary GlobalClusterMemberRole = "SECONDARY"
)

type GlobalCluster struct {
	GlobalClusterID   string
	GlobalClusterName string
	ProjectID         string
	RegionIDs         []string
	CUType            string
	CUSize            int64
	ConnectAddress    string
	CreateTime        string
	Clusters          []GlobalClusterMember
}

type GlobalClusterMember struct {
	ClusterID   string
	ClusterName string
	RegionID    string
	Role        GlobalClusterMemberRole
	Status      string
}

func (m GlobalClusterMember) isRunning() bool {
	return m.Status == "RUNNING"
}

type GlobalClusterMemberSpec struct {
	ClusterName string
	RegionID    string
}

type CreateGlobalClusterCommand struct {
	GlobalClusterName string
	ProjectID         string
	CUType            string
	CUSize            int64
	Members           []GlobalClusterMemberSpec
}

type CreateGlobalClusterResult struct {
	GlobalClusterID string
	Username        string
	Password        string
	JobID           string
}

func (c *GlobalCluster) memberByClusterID(clusterID string) (GlobalClusterMember, bool) {
	if c == nil {
		return GlobalClusterMember{}, false
	}
	for _, member := range c.Clusters {
		if member.ClusterID == clusterID {
			return member, true
		}
	}
	return GlobalClusterMember{}, false
}

func (c *GlobalCluster) primaryMember() (GlobalClusterMember, bool) {
	if c == nil {
		return GlobalClusterMember{}, false
	}
	for _, member := range c.Clusters {
		if member.Role == GlobalClusterMemberRolePrimary {
			return member, true
		}
	}
	return GlobalClusterMember{}, false
}

func (c *GlobalCluster) isInstanceNotExists(clusterID string) (bool, string, error) {
	member, exists := c.memberByClusterID(clusterID)
	if !exists {
		return true, "", nil
	}
	return false, member.Status, nil
}

func (c *GlobalCluster) isCUUpdated(targetCUSize int64) (bool, string, error) {
	if c == nil {
		return false, "missing", nil
	}
	if c.CUSize != targetCUSize {
		return false, fmt.Sprintf("global_cu_size=%d", c.CUSize), nil
	}
	for _, member := range c.Clusters {
		if !member.isRunning() {
			return false, member.Status, nil
		}
	}
	return true, "RUNNING", nil
}

func (c *GlobalCluster) isPrimaryMemberRunning() (bool, string, error) {
	if c == nil {
		return false, "missing", nil
	}

	primary, hasPrimary := c.primaryMember()
	primaryStatus := "missing"
	if hasPrimary {
		primaryStatus = primary.Status
	}

	primaryCount := 0
	for _, member := range c.Clusters {
		if member.Role == GlobalClusterMemberRolePrimary {
			primaryCount++
		}
	}

	if primaryCount != 1 {
		return false, primaryStatus, fmt.Errorf("global cluster %s has %d primary members", c.GlobalClusterID, primaryCount)
	}

	return primary.isRunning(), primaryStatus, nil
}

func (c *GlobalCluster) isSecondaryMemberRunning(spec GlobalClusterMemberSpec) (bool, string, error) {
	if c == nil {
		return false, "missing", nil
	}
	for _, member := range c.Clusters {
		if member.ClusterName != spec.ClusterName || member.RegionID != spec.RegionID {
			continue
		}
		if member.Role != GlobalClusterMemberRoleSecondary {
			return false, "", fmt.Errorf("cluster %s in region %s appears in global cluster members with role %s, not %s", spec.ClusterName, spec.RegionID, member.Role, GlobalClusterMemberRoleSecondary)
		}
		return member.isRunning(), member.Status, nil
	}
	return false, "missing", nil
}

func GlobalClusterFromAPI(api *zilliz.GlobalCluster) *GlobalCluster {
	if api == nil {
		return nil
	}

	clusters := make([]GlobalClusterMember, 0, len(api.Clusters))
	for _, member := range api.Clusters {
		clusters = append(clusters, GlobalClusterMember{
			ClusterID:   member.ClusterId,
			ClusterName: member.ClusterName,
			RegionID:    member.RegionId,
			Role:        GlobalClusterMemberRole(member.Role),
			Status:      member.Status,
		})
	}

	return &GlobalCluster{
		GlobalClusterID:   api.GlobalClusterId,
		GlobalClusterName: api.GlobalClusterName,
		ProjectID:         api.ProjectId,
		RegionIDs:         append([]string(nil), api.RegionIds...),
		CUType:            api.CuType,
		CUSize:            api.CuSize,
		ConnectAddress:    api.ConnectAddress,
		CreateTime:        api.CreateTime,
		Clusters:          clusters,
	}
}

func IsNotFoundError(err error) bool {
	var apiErr zilliz.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
		return true
	}

	var apiErrPtr *zilliz.Error
	if errors.As(err, &apiErrPtr) && apiErrPtr != nil && apiErrPtr.Code == http.StatusNotFound {
		return true
	}

	return strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not exist")
}

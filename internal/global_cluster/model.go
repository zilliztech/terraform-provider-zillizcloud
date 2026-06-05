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

func (c *GlobalCluster) isInstanceNotExists(clusterID string) (bool, string, error) {
	member, exists := c.memberByClusterID(clusterID)
	if !exists {
		return true, "", nil
	}
	return false, member.Status, nil
}

func (c *GlobalCluster) isInstanceRunning(clusterID string) (bool, string, error) {
	member, exists := c.memberByClusterID(clusterID)
	if !exists {
		return false, "missing", nil
	}
	if member.Role != GlobalClusterMemberRoleSecondary {
		return false, "", fmt.Errorf("cluster %s appears in global cluster members with role %s, not %s", clusterID, member.Role, GlobalClusterMemberRoleSecondary)
	}
	return member.Status == "RUNNING", member.Status, nil
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
		return member.Status == "RUNNING", member.Status, nil
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

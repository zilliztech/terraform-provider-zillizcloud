package byoc_project

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	globalClusterRolePrimary   = "PRIMARY"
	globalClusterRoleSecondary = "SECONDARY"
	globalClusterMemberStatus  = "RUNNING"
)

func ListGlobalClusters(c *gin.Context) {
	currentPage := parsePositiveQuery(c.Query("currentPage"), 1)
	pageSize := parsePositiveQuery(c.Query("pageSize"), 10)
	if pageSize > 100 {
		pageSize = 100
	}

	clusters := globalClusterStore.GetAll()
	items := make([]GlobalClusterResponse, 0, len(clusters))
	for _, cluster := range clusters {
		items = append(items, clusterListItem(*cluster))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreateTime < items[j].CreateTime
	})

	start := (currentPage - 1) * pageSize
	if start > len(items) {
		start = len(items)
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"count":          len(items),
			"currentPage":    currentPage,
			"pageSize":       pageSize,
			"globalClusters": items[start:end],
		},
	})
}

func CreateGlobalCluster(c *gin.Context) {
	var request CreateGlobalClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := validateCreateGlobalClusterRequest(request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	globalClusterId := "glo-" + uuid.New().String()[:16]
	createTime := time.Now().Format(time.RFC3339)
	connectAddress := fmt.Sprintf("https://%s.global-cluster.vectordb.zillizcloud.com", globalClusterId)

	members := []GlobalClusterMember{
		newGlobalClusterMember(request.PrimaryCluster, globalClusterRolePrimary),
	}
	for _, secondary := range request.SecondaryClusters {
		members = append(members, newGlobalClusterMember(secondary, globalClusterRoleSecondary))
	}

	globalCluster := &GlobalCluster{
		GlobalClusterId:   globalClusterId,
		GlobalClusterName: request.GlobalClusterName,
		ProjectId:         request.ProjectId,
		RegionIds:         regionIdsFromMembers(members),
		CuType:            request.CuType,
		CuSize:            request.CuSize,
		Statuses:          statusesFromMembers(members),
		ConnectAddress:    connectAddress,
		CreatedBy:         "mock-server@example.com",
		CreateTime:        createTime,
		Clusters:          members,
	}

	globalClusterStore.Set(globalClusterId, globalCluster)
	for _, member := range members {
		clusterStore.Set(member.ClusterId, member.ToDedicatedCluster(request, globalClusterId, createTime))
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"globalClusterId": globalClusterId,
			"username":        "db_admin",
			"password":        "password",
			"jobId":           "job-" + uuid.New().String()[:16],
		},
	})
}

func DescribeGlobalCluster(c *gin.Context) {
	globalCluster := getGlobalClusterOrRespond(c)
	if globalCluster == nil {
		return
	}

	refreshGlobalClusterFromMembers(globalCluster)
	c.JSON(http.StatusOK, Response[GlobalClusterResponse]{Code: 0, Data: globalClusterResponse(*globalCluster, true)})
}

func ModifyGlobalClusterCU(c *gin.Context) {
	globalCluster := getGlobalClusterOrRespond(c)
	if globalCluster == nil {
		return
	}

	var request ModifyClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if request.CuSize == nil && request.Autoscaling == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "cuSize or autoscaling is required"})
		return
	}
	if request.CuSize != nil && request.Autoscaling != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "cuSize and autoscaling cannot be set at the same time"})
		return
	}

	if request.CuSize != nil {
		globalCluster.CuSize = *request.CuSize
	}
	globalCluster.Autoscaling = request.Autoscaling

	for _, member := range globalCluster.Clusters {
		cluster := clusterStore.Get(member.ClusterId)
		if cluster == nil {
			continue
		}
		if request.CuSize != nil {
			cluster.CuSize = *request.CuSize
		}
		if request.Autoscaling != nil {
			cluster.Autoscaling = *request.Autoscaling
		}
		clusterStore.Set(member.ClusterId, cluster)
	}
	globalClusterStore.Set(globalCluster.GlobalClusterId, globalCluster)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"jobId": "job-" + uuid.New().String()[:16]},
	})
}

func RemoveGlobalEndpoint(c *gin.Context) {
	globalCluster := getGlobalClusterOrRespond(c)
	if globalCluster == nil {
		return
	}

	if countSecondaryMembers(globalCluster.Clusters) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "secondary clusters must be deleted before removing global endpoint"})
		return
	}

	primary := primaryMember(globalCluster.Clusters)
	if primary == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "primary cluster is required"})
		return
	}

	cluster := clusterStore.Get(primary.ClusterId)
	if cluster != nil {
		cluster.GlobalClusterMeta = nil
		cluster.ConnectAddress = fmt.Sprintf("https://%s.%s.vectordb.zillizcloud.com:19530", cluster.ClusterId, cluster.RegionId)
		clusterStore.Set(primary.ClusterId, cluster)
	}
	globalClusterStore.Delete(globalCluster.GlobalClusterId)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"oldGlobalClusterId": globalCluster.GlobalClusterId},
	})
}

func AddSecondaryClusters(c *gin.Context) {
	globalCluster := getGlobalClusterOrRespond(c)
	if globalCluster == nil {
		return
	}

	var request AddSecondaryClustersRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if len(request.SecondaryClusters) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "secondaryClusters is required"})
		return
	}

	createTime := time.Now().Format(time.RFC3339)
	for _, secondary := range request.SecondaryClusters {
		if secondary.ClusterName == "" || secondary.RegionId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "secondary clusterName and regionId are required"})
			return
		}
		member := newGlobalClusterMember(secondary, globalClusterRoleSecondary)
		globalCluster.Clusters = append(globalCluster.Clusters, member)
		clusterStore.Set(member.ClusterId, member.ToDedicatedClusterFromGlobalCluster(*globalCluster, createTime))
	}
	refreshGlobalClusterFromMembers(globalCluster)
	globalClusterStore.Set(globalCluster.GlobalClusterId, globalCluster)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"jobId": "job-" + uuid.New().String()[:16]},
	})
}

func DeleteSecondaryCluster(c *gin.Context) {
	deleteGlobalClusterMember(c, false)
}

func DeleteCluster(c *gin.Context) {
	deleteGlobalClusterMember(c, true)
}

func deleteGlobalClusterMember(c *gin.Context, allowPrimary bool) {
	globalCluster := getGlobalClusterOrRespond(c)
	if globalCluster == nil {
		return
	}

	clusterId := c.Param("clusterId")
	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "clusterId is required"})
		return
	}

	memberIndex := -1
	for i, member := range globalCluster.Clusters {
		if member.ClusterId == clusterId {
			memberIndex = i
			break
		}
	}
	if memberIndex < 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "cluster not found"})
		return
	}
	member := globalCluster.Clusters[memberIndex]
	if member.Role == globalClusterRolePrimary && !allowPrimary {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "primary cluster cannot be deleted by secondary cluster API"})
		return
	}
	if member.Role == globalClusterRolePrimary && countSecondaryMembers(globalCluster.Clusters) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "secondary clusters must be deleted before deleting primary cluster"})
		return
	}

	globalCluster.Clusters = append(globalCluster.Clusters[:memberIndex], globalCluster.Clusters[memberIndex+1:]...)
	clusterStore.Delete(clusterId)
	if member.Role == globalClusterRolePrimary {
		globalClusterStore.Delete(globalCluster.GlobalClusterId)
	} else {
		refreshGlobalClusterFromMembers(globalCluster)
		globalClusterStore.Set(globalCluster.GlobalClusterId, globalCluster)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"globalClusterId": globalCluster.GlobalClusterId,
			"clusterId":       clusterId,
			"prompt":          "The cluster has been deleted. If you consider this action to be an error, you have the option to restore the deleted cluster from the recycle bin within a 30-day period. Kindly note, this recovery feature does not apply to free clusters.",
		},
	})
}

func getGlobalClusterOrRespond(c *gin.Context) *GlobalCluster {
	globalClusterId := c.Param("globalClusterId")
	if globalClusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "globalClusterId is required"})
		return nil
	}

	globalCluster := globalClusterStore.Get(globalClusterId)
	if globalCluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "global cluster not found"})
		return nil
	}
	return globalCluster
}

func validateCreateGlobalClusterRequest(request CreateGlobalClusterRequest) error {
	if request.GlobalClusterName == "" {
		return fmt.Errorf("globalClusterName is required")
	}
	if request.ProjectId == "" {
		return fmt.Errorf("projectId is required")
	}
	if request.CuType == "" {
		return fmt.Errorf("cuType is required")
	}
	if request.CuSize <= 0 {
		return fmt.Errorf("cuSize is required")
	}
	if request.PrimaryCluster.ClusterName == "" || request.PrimaryCluster.RegionId == "" {
		return fmt.Errorf("primaryCluster clusterName and regionId are required")
	}
	if len(request.SecondaryClusters) == 0 {
		return fmt.Errorf("secondaryClusters is required")
	}
	for _, secondary := range request.SecondaryClusters {
		if secondary.ClusterName == "" || secondary.RegionId == "" {
			return fmt.Errorf("secondary clusterName and regionId are required")
		}
	}
	return nil
}

func newGlobalClusterMember(memberRequest GlobalClusterMemberRequest, role string) GlobalClusterMember {
	return GlobalClusterMember{
		ClusterId:   "in01-" + uuid.New().String()[:15],
		ClusterName: memberRequest.ClusterName,
		RegionId:    memberRequest.RegionId,
		Role:        role,
		Status:      globalClusterMemberStatus,
	}
}

func (m GlobalClusterMember) ToDedicatedCluster(request CreateGlobalClusterRequest, globalClusterId string, createTime string) *DedicatedClusterResponse {
	return &DedicatedClusterResponse{
		ClusterId:          m.ClusterId,
		ClusterName:        m.ClusterName,
		ProjectId:          request.ProjectId,
		Description:        "",
		RegionId:           m.RegionId,
		CuType:             request.CuType,
		Plan:               "Enterprise",
		Status:             m.Status,
		ConnectAddress:     fmt.Sprintf("https://%s.%s.vectordb.zillizcloud.com:19530", m.ClusterId, m.RegionId),
		PrivateLinkAddress: "",
		CreateTime:         createTime,
		Replica:            1,
		CuSize:             request.CuSize,
		StorageSize:        0,
		SnapshotNumber:     0,
		CreateProgress:     100,
		Username:           "db_admin",
		Password:           "password",
		Prompt:             "Successfully Submitted",
		GlobalClusterMeta: &GlobalClusterMeta{
			GlobalClusterId: globalClusterId,
			Role:            m.Role,
		},
	}
}

func (m GlobalClusterMember) ToDedicatedClusterFromGlobalCluster(globalCluster GlobalCluster, createTime string) *DedicatedClusterResponse {
	return &DedicatedClusterResponse{
		ClusterId:          m.ClusterId,
		ClusterName:        m.ClusterName,
		ProjectId:          globalCluster.ProjectId,
		Description:        "",
		RegionId:           m.RegionId,
		CuType:             globalCluster.CuType,
		Plan:               "Enterprise",
		Status:             m.Status,
		ConnectAddress:     fmt.Sprintf("https://%s.%s.vectordb.zillizcloud.com:19530", m.ClusterId, m.RegionId),
		PrivateLinkAddress: "",
		CreateTime:         createTime,
		Replica:            1,
		CuSize:             globalCluster.CuSize,
		StorageSize:        0,
		SnapshotNumber:     0,
		CreateProgress:     100,
		Username:           "db_admin",
		Password:           "password",
		Prompt:             "Successfully Submitted",
		GlobalClusterMeta: &GlobalClusterMeta{
			GlobalClusterId: globalCluster.GlobalClusterId,
			Role:            m.Role,
		},
	}
}

func refreshGlobalClusterFromMembers(globalCluster *GlobalCluster) {
	for i := range globalCluster.Clusters {
		cluster := clusterStore.Get(globalCluster.Clusters[i].ClusterId)
		if cluster == nil {
			continue
		}
		globalCluster.Clusters[i].Status = cluster.Status
		globalCluster.Clusters[i].ClusterName = cluster.ClusterName
	}
	globalCluster.RegionIds = regionIdsFromMembers(globalCluster.Clusters)
	globalCluster.Statuses = statusesFromMembers(globalCluster.Clusters)
}

func clusterListItem(cluster GlobalCluster) GlobalClusterResponse {
	refreshGlobalClusterFromMembers(&cluster)
	return globalClusterResponse(cluster, false)
}

func globalClusterResponse(cluster GlobalCluster, includeMembers bool) GlobalClusterResponse {
	response := GlobalClusterResponse{
		GlobalClusterId:   cluster.GlobalClusterId,
		GlobalClusterName: cluster.GlobalClusterName,
		ProjectId:         cluster.ProjectId,
		RegionIds:         cluster.RegionIds,
		CuType:            cluster.CuType,
		CuSize:            cluster.CuSize,
		ConnectAddress:    cluster.ConnectAddress,
		CreateTime:        cluster.CreateTime,
	}
	if includeMembers {
		response.Clusters = cluster.Clusters
	}
	return response
}

func regionIdsFromMembers(members []GlobalClusterMember) []string {
	regions := make([]string, 0, len(members))
	for _, member := range members {
		regions = append(regions, member.RegionId)
	}
	return regions
}

func countSecondaryMembers(members []GlobalClusterMember) int {
	count := 0
	for _, member := range members {
		if member.Role == globalClusterRoleSecondary {
			count++
		}
	}
	return count
}

func primaryMember(members []GlobalClusterMember) *GlobalClusterMember {
	for i := range members {
		if members[i].Role == globalClusterRolePrimary {
			return &members[i]
		}
	}
	return nil
}

func statusesFromMembers(members []GlobalClusterMember) []string {
	seen := map[string]bool{}
	statuses := make([]string, 0, len(members))
	for _, member := range members {
		if seen[member.Status] {
			continue
		}
		seen[member.Status] = true
		statuses = append(statuses, member.Status)
	}
	return statuses
}

func parsePositiveQuery(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

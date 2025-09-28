package byoc_project

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateDedicatedCluster(c *gin.Context) {
	var request CreateDedicatedClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[CreateDedicatedCluster request]: %+v", request)

	var clusterId string
	if request.ClusterId == "" {
		clusterId = "in01-" + uuid.New().String()[:15]
	} else {
		clusterId = request.ClusterId
	}
	connectAddress := fmt.Sprintf("https://%s.%s.vectordb-uat3.zillizcloud.com:19540", clusterId, request.RegionId)

	cluster := &DedicatedClusterResponse{
		ClusterId:          clusterId,
		ClusterName:        request.ClusterName,
		ProjectId:          request.ProjectId,
		Description:        request.Description,
		RegionId:           request.RegionId,
		CuType:             request.CuType,
		Plan:               request.Plan,
		ConnectAddress:     connectAddress,
		PrivateLinkAddress: "",
		CreateTime:         time.Now().Format(time.RFC3339),
		CuSize:             request.CuSize,
		StorageSize:        0,
		SnapshotNumber:     0,
		CreateProgress:     100,
		Labels:             request.Labels,
	}
	if cluster.Plan == "Standard" || cluster.Plan == "Enterprise" {
		cluster.Replica = 1
	}
	cluster.Status = "CREATING"
	clusterStore.Set(clusterId, cluster)

	go func() {
		time.Sleep(5 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[ResumeCluster] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, Response[*DedicatedClusterResponse]{
		Code: 0,
		Data: cluster,
	})
}

func CreateServerlessCluster(c *gin.Context) {
	var request CreateServerlessClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[CreateServerlessCluster request]: %+v", request)

	clusterId := "in05-" + uuid.New().String()[:15]
	connectAddress := fmt.Sprintf("https://%s.%s.vectordb-uat3.zillizcloud.com:19540", clusterId, request.RegionId)

	cluster := &ServerlessClusterResponse{
		ClusterId:          clusterId,
		ClusterName:        request.ClusterName,
		ProjectId:          request.ProjectId,
		Description:        request.Description,
		RegionId:           request.RegionId,
		Plan:               request.Plan,
		ConnectAddress:     connectAddress,
		PrivateLinkAddress: "",
		CreateTime:         time.Now().Format(time.RFC3339),
		StorageSize:        0,
		SnapshotNumber:     0,
		CreateProgress:     100,
		Labels:             request.Labels,
	}
	cluster.Status = "CREATING"

	// Store the serverless cluster (we'll need to update the store to handle different cluster types)
	// For now, we'll convert it to the DedicatedClusterResponse format for storage
	dedicatedCluster := &DedicatedClusterResponse{
		ClusterId:          cluster.ClusterId,
		ClusterName:        cluster.ClusterName,
		ProjectId:          cluster.ProjectId,
		Description:        cluster.Description,
		RegionId:           cluster.RegionId,
		Plan:               "Serverless",
		Status:             cluster.Status,
		ConnectAddress:     cluster.ConnectAddress,
		PrivateLinkAddress: cluster.PrivateLinkAddress,
		CreateTime:         cluster.CreateTime,
		StorageSize:        cluster.StorageSize,
		SnapshotNumber:     cluster.SnapshotNumber,
		CreateProgress:     cluster.CreateProgress,
		Labels:             cluster.Labels,
		CuType:             "",
		CuSize:             0,
		Replica:            1,
	}
	clusterStore.Set(clusterId, dedicatedCluster)

	go func() {
		time.Sleep(5 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[CreateServerlessCluster] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, Response[*ServerlessClusterResponse]{
		Code: 0,
		Data: cluster,
	})
}

func CreateFreeCluster(c *gin.Context) {
	var request CreateFreeClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[CreateFreeCluster request]: %+v", request)

	clusterId := "in03-" + uuid.New().String()[:15]
	connectAddress := fmt.Sprintf("https://%s.%s.vectordb-uat3.zillizcloud.com:19540", clusterId, request.RegionId)

	cluster := &FreeClusterResponse{
		ClusterId:          clusterId,
		ClusterName:        request.ClusterName,
		ProjectId:          request.ProjectId,
		Description:        request.Description,
		RegionId:           request.RegionId,
		Plan:               "Serverless",
		ConnectAddress:     connectAddress,
		PrivateLinkAddress: "",
		CreateTime:         time.Now().Format(time.RFC3339),
		StorageSize:        0,
		SnapshotNumber:     0,
		CreateProgress:     100,
		Labels:             request.Labels,
	}
	cluster.Status = "CREATING"

	// Store the free cluster (we'll need to update the store to handle different cluster types)
	// For now, we'll convert it to the DedicatedClusterResponse format for storage
	dedicatedCluster := &DedicatedClusterResponse{
		ClusterId:          cluster.ClusterId,
		ClusterName:        cluster.ClusterName,
		ProjectId:          cluster.ProjectId,
		Description:        cluster.Description,
		RegionId:           cluster.RegionId,
		Plan:               cluster.Plan,
		Status:             cluster.Status,
		ConnectAddress:     cluster.ConnectAddress,
		PrivateLinkAddress: cluster.PrivateLinkAddress,
		CreateTime:         cluster.CreateTime,
		StorageSize:        cluster.StorageSize,
		SnapshotNumber:     cluster.SnapshotNumber,
		CreateProgress:     cluster.CreateProgress,
		Labels:             cluster.Labels,
		CuType:             "",
		CuSize:             0,
		Replica:            1,
	}
	clusterStore.Set(clusterId, dedicatedCluster)

	go func() {
		time.Sleep(5 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[CreateFreeCluster] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, Response[*FreeClusterResponse]{
		Code: 0,
		Data: cluster,
	})
}

func GetCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[GetCluster] clusterId: %s", clusterId)

	c.JSON(http.StatusOK, Response[*DedicatedClusterResponse]{
		Code: 0,
		Data: cluster,
	})
}

func ResumeCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	if cluster.Status != "STOPPED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster status is not STOPPED"})
		return
	}

	log.Printf("[ResumeCluster] clusterId: %s, changing status from STOPPED -> RESUMING -> RUNNING", clusterId)

	cluster.Status = "RESUMING"
	clusterStore.Set(clusterId, cluster)

	go func() {
		time.Sleep(5 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[ResumeCluster] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

func SuspendCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	if cluster.Status != "RUNNING" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster status is not RUNNING"})
		return
	}

	log.Printf("[SuspendCluster] clusterId: %s, changing status from RUNNING -> STOPPING -> STOPPED", clusterId)

	cluster.Status = "STOPPING"
	clusterStore.Set(clusterId, cluster)

	go func() {
		time.Sleep(5 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "STOPPED"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[SuspendCluster] clusterId: %s status changed to STOPPED", clusterId)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

func GetProjects(c *gin.Context) {
	projects := projectStore.GetAll()

	projectList := make([]Project, 0, len(projects))
	for _, p := range projects {
		projectList = append(projectList, *p)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": projectList,
	})
}

func ModifyClusterReplica(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	var request ModifyReplicaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[ModifyClusterReplica] clusterId: %s, changing replica from %d to %d, status: RUNNING -> MODIFYING -> RUNNING", clusterId, cluster.Replica, request.Replica)

	cluster.Status = "MODIFYING"
	cluster.Replica = request.Replica
	clusterStore.Set(clusterId, cluster)

	go func() {
		time.Sleep(10 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[ModifyClusterReplica] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}



func GetLabels(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[GetLabels] clusterId: %s", clusterId)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"labels": cluster.Labels,
		},
	})
}

func UpdateLabels(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	var request UpdateLabelsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[UpdateLabels] clusterId: %s, updating labels from %v to %v", clusterId, cluster.Labels, request.Labels)

	cluster.Labels = request.Labels
	clusterStore.Set(clusterId, cluster)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

func ModifyClusterProperties(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	var request ModifyPropertiesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[ModifyClusterProperties] clusterId: %s, changing cluster name from %s to %s", clusterId, cluster.ClusterName, request.ClusterName)

	cluster.ClusterName = request.ClusterName
	clusterStore.Set(clusterId, cluster)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

func GetSecurityGroups(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[GetSecurityGroups] clusterId: %s", clusterId)

	response := GetSecurityGroupsResponse{
		Ids: cluster.SecurityGroups,
	}

	c.JSON(http.StatusOK, Response[GetSecurityGroupsResponse]{
		Code: 0,
		Data: response,
	})
}

func UpsertSecurityGroups(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	var request UpsertSecurityGroupsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[UpsertSecurityGroups] clusterId: %s, updating security groups from %v to %v", clusterId, cluster.SecurityGroups, request.Ids)

	cluster.SecurityGroups = request.Ids
	clusterStore.Set(clusterId, cluster)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}


package byoc_project

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	store        = NewStore()
	clusterStore = NewSafeStore[DedicatedClusterResponse]()
)

func DeleteDataplane(c *gin.Context) {
	var request DeleteDataplaneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dataplane := dataplaneStore.Get(request.ProjectId)
	if dataplane == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataplane not found"})
		return
	}
	if dataplane.Status == int(BYOCProjectStatusDeleting) || dataplane.Status == int(BYOCProjectStatusDeleted) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataplane is already being deleted"})
		return
	}

	updateDataplaneStatus(dataplane, int(BYOCProjectStatusDeleting))
	go func() {
		time.Sleep(10 * time.Second)
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusDeleted))
	}()
	log.Printf("deleteDataplane request: %+v", request)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"projectId":   request.ProjectId,
			"dataPlaneId": request.DataPlaneId,
		},
	})
}

// suspend dataplane
func SuspendDataplane(c *gin.Context) {
	var request SuspendDataplaneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dataplane := dataplaneStore.Get(request.ProjectId)
	if dataplane == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataplane not found"})
		return
	}
	if dataplane.Status != int(BYOCProjectStatusRunning) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataplane status is not running"})
		return
	}
	updateDataplaneStatus(dataplane, int(BYOCProjectStatusStopping))
	go func() {
		time.Sleep(10 * time.Second)
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusStopped))
	}()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"projectId":   request.ProjectId,
			"dataPlaneId": request.DataPlaneId,
		},
	})
}

// resume dataplane
func ResumeDataplane(c *gin.Context) {
	var request ResumeDataplaneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dataplane := dataplaneStore.Get(request.ProjectId)
	if dataplane == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataplane not found"})
		return
	}
	if dataplane.Status != int(BYOCProjectStatusStopped) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataplane status is not stopped"})
		return
	}
	updateDataplaneStatus(dataplane, int(BYOCProjectStatusResuming))
	go func() {
		time.Sleep(10 * time.Second)
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusRunning))
	}()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"projectId":   request.ProjectId,
			"dataPlaneId": request.DataPlaneId,
		},
	})
}

func updateDataplaneStatus(dataplane *DataplaneResponse, status int) {
	// fmt.Printf("updateDataplaneStatus: %+v, %d\n", dataplane, status)
	fmt.Printf("status is changed !!! [projectId: %+v, dataPlaneId: %+v, oldStatus: %d, newStatus: %d]\n", dataplane.ProjectID, dataplane.DataPlaneID, dataplane.Status, status)
	if dataplane == nil {
		fmt.Println("dataplane not found")
		return
	}
	dataplane.Status = status
	dataplaneStore.Set(dataplane.ProjectID, dataplane)
}

func CreateDataplane(c *gin.Context) {
	var request CreateDataplaneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("request: %+v", request)

	response := request.ToDataplane()
	dataplaneStore.Set(response.ProjectID, response)

	go func() {

		dataplane := dataplaneStore.Get(response.ProjectID)
		if dataplane == nil {
			fmt.Println("dataplane not found")
			return
		}

		time.Sleep(10 * time.Second)
		// store.UpdateStatus(response.ProjectID, int(BYOCProjectStatusRunning))
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusRunning))

		fmt.Println("update status to 1")
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"jobId":       uuid.New().String(),
			"projectId":   response.ProjectID,
			"dataPlaneId": response.DataPlaneID,
		},
	})
}

// op dataplane is kind of like a dataplane
// before create op dataplane, we init a dataplane(op project) with settings interface
// then we create op dataplane with the dataplane id
// then we update the dataplane status to running
func CreateOpDataplane(c *gin.Context) {
	var request CreateOpDataplaneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[CreateOpDataplane request]: %+v", request)

	// assert dataplane exists
	dataplane := dataplaneStore.Get(request.ProjectID)
	if dataplane == nil {
		fmt.Println("dataplane not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "dataplane not found"})
		return
	}

	if dataplane.Status != int(BYOCProjectStatusConnected) {
		fmt.Println("dataplane status is not connected")
		c.JSON(http.StatusBadRequest, gin.H{"error": "dataplane status is not connected"})
		return
	}

	// update dataplane status to pending immediately
	updateDataplaneStatus(dataplane, int(BYOCProjectStatusPending))

	go func() {

		dataplane := dataplaneStore.Get(request.ProjectID)
		if dataplane == nil {
			fmt.Println("dataplane not found")
			return
		}

		time.Sleep(10 * time.Second)
		// store.UpdateStatus(response.ProjectID, int(BYOCProjectStatusRunning))
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusRunning))
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"jobId":       uuid.New().String(),
			"projectId":   request.ProjectID,
			"dataPlaneId": request.DataPlaneID,
		},
	})
}

func DescribeDataplane(c *gin.Context) {
	projectId := c.Query("projectId")

	// dataplane := store.Get(request.ProjectId)
	dataplane := dataplaneStore.Get(projectId)
	if dataplane == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataplane not found"})
		return
	}
	log.Printf("dataplane: %+v", dataplane)
	c.JSON(http.StatusOK, Response[*DataplaneResponse]{
		Code: 0,
		Data: dataplane,
	})
}

func GetSettings(c *gin.Context) {
	projectId := c.Query("projectId")

	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "projectId is required"})
		return
	}

	settings := settingsStore.Get(projectId)
	if settings == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "settings not found"})
		return
	}
	c.JSON(http.StatusOK, Response[*SettingsResponse]{
		Code: 0,
		Data: settings,
	})
}

func CreateSettings(c *gin.Context) {
	var request SettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[CreateSettings request]: %+v", request)
	settings := NewSettingsResponse(
		WithProjectId(request.ProjectId),
		WithDataPlaneId(request.DataPlaneId),
		WithProjectName(request.ProjectName),
		WithRegionId(request.RegionId),
		WithCloudId(request.CloudId),
		WithByocId(request.ByocId),
		WithOpenPl(request.OpenPl),
		WithNodeQuotas([]NodeQuota{
			WithNodeQuota("index", 1),
			WithNodeQuota("search", int(request.SearchMin)),
			WithNodeQuota("fundamental", int(request.FundamentalMin)),
			WithNodeQuota("core", int(request.CoreMin)),
		}),
		WithOpConfig("sk-op-token", "sk-op-agent-image-url"),
	)

	// could query by dataPlaneId or projectId later on
	// settingsStore.Set(settings.DataPlaneId, settings)
	settingsStore.Set(settings.ProjectId, settings)

	dataplane := settings.IntoDataplane()
	dataplaneStore.Set(dataplane.ProjectID, dataplane)
	go func() {
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusInit))
		time.Sleep(10 * time.Second)
		updateDataplaneStatus(dataplane, int(BYOCProjectStatusConnected))
	}()
	c.JSON(http.StatusOK, Response[*SettingsResponse]{
		Code: 0,
		Data: settings,
	})
}

func DropCluster(c *gin.Context) {
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

	log.Printf("[DropCluster] clusterId: %s, deleting cluster", clusterId)

	clusterStore.Set(clusterId, nil)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

func ModifyCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")

	if clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clusterId is required"})
		return
	}

	var request ModifyClusterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cluster := clusterStore.Get(clusterId)
	if cluster == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	log.Printf("[ModifyCluster] clusterId: %s, changing cuSize from %d to %d, status: RUNNING -> MODIFYING -> RUNNING", clusterId, cluster.CuSize, request.CuSize)

	cluster.Status = "MODIFYING"
	cluster.CuSize = request.CuSize
	clusterStore.Set(clusterId, cluster)

	go func() {
		time.Sleep(10 * time.Second)
		cluster := clusterStore.Get(clusterId)
		if cluster != nil {
			cluster.Status = "RUNNING"
			clusterStore.Set(clusterId, cluster)
			log.Printf("[ModifyCluster] clusterId: %s status changed to RUNNING", clusterId)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"clusterId": clusterId,
		},
	})
}

package byoc_project

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGlobalClusterAndSecondaryClusterLifecycle(t *testing.T) {
	globalClusterStore = NewSafeStore[GlobalCluster]()
	clusterStore = NewSafeStore[DedicatedClusterResponse]()

	router := newGlobalClusterTestRouter()
	createBody := map[string]any{
		"globalClusterName": "my-global-cluster",
		"projectId":         "proj-1",
		"cuType":            "Performance-optimized",
		"cuSize":            4,
		"primaryCluster": map[string]any{
			"clusterName": "primary-cluster",
			"regionId":    "aws-us-west-2",
		},
		"secondaryClusters": []map[string]any{
			{
				"clusterName": "secondary-cluster-eu",
				"regionId":    "aws-eu-west-1",
			},
		},
	}

	createResp := performGlobalClusterRequest(t, router, http.MethodPost, "/v2/globalClusters/create", createBody)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	globalClusterId := requireJSONString(t, createResp.Body.Bytes(), "data.globalClusterId")
	requireJSONValue(t, createResp.Body.Bytes(), "data.username", "db_admin")
	requireJSONValue(t, createResp.Body.Bytes(), "data.password", "password")
	requireJSONPresent(t, createResp.Body.Bytes(), "data.jobId")
	requireJSONMissing(t, createResp.Body.Bytes(), "data.prompt")

	describeResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/globalClusters/"+globalClusterId, nil)
	if describeResp.Code != http.StatusOK {
		t.Fatalf("describe status=%d body=%s", describeResp.Code, describeResp.Body.String())
	}
	requireJSONValue(t, describeResp.Body.Bytes(), "data.globalClusterName", "my-global-cluster")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.projectId", "proj-1")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.regionIds.0", "aws-us-west-2")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.regionIds.1", "aws-eu-west-1")
	requireJSONMissing(t, describeResp.Body.Bytes(), "data.statuses")
	requireJSONMissing(t, describeResp.Body.Bytes(), "data.createdBy")
	requireJSONMissing(t, describeResp.Body.Bytes(), "data.autoscaling")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.clusters.0.role", "PRIMARY")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.clusters.1.role", "SECONDARY")
	primaryClusterId := requireJSONString(t, describeResp.Body.Bytes(), "data.clusters.0.clusterId")
	secondaryClusterId := requireJSONString(t, describeResp.Body.Bytes(), "data.clusters.1.clusterId")

	primaryResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/clusters/"+primaryClusterId, nil)
	if primaryResp.Code != http.StatusOK {
		t.Fatalf("primary describe status=%d body=%s", primaryResp.Code, primaryResp.Body.String())
	}
	requireJSONValue(t, primaryResp.Body.Bytes(), "data.globalClusterMeta.globalClusterId", globalClusterId)
	requireJSONValue(t, primaryResp.Body.Bytes(), "data.globalClusterMeta.role", "PRIMARY")

	listResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/globalClusters?projectId=proj-1&currentPage=1&pageSize=10", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	requireJSONValue(t, listResp.Body.Bytes(), "data.count", float64(1))
	requireJSONValue(t, listResp.Body.Bytes(), "data.globalClusters.0.globalClusterId", globalClusterId)
	requireJSONValue(t, listResp.Body.Bytes(), "data.globalClusters.0.cuSize", float64(4))
	requireJSONMissing(t, listResp.Body.Bytes(), "data.globalClusters.0.statuses")
	requireJSONMissing(t, listResp.Body.Bytes(), "data.globalClusters.0.clusters")
	requireJSONMissing(t, listResp.Body.Bytes(), "data.globalClusters.0.autoscaling")

	invalidModifyResp := performGlobalClusterRequest(t, router, http.MethodPost, "/v2/globalClusters/"+globalClusterId+"/modifyCU", map[string]any{
		"cuSize": 8,
		"autoscaling": map[string]any{
			"cu": map[string]any{"min": 4, "max": 16},
		},
	})
	if invalidModifyResp.Code != http.StatusBadRequest {
		t.Fatalf("modifyCU with cuSize and autoscaling status=%d body=%s", invalidModifyResp.Code, invalidModifyResp.Body.String())
	}

	modifyResp := performGlobalClusterRequest(t, router, http.MethodPost, "/v2/globalClusters/"+globalClusterId+"/modifyCU", map[string]any{"cuSize": 8})
	if modifyResp.Code != http.StatusOK {
		t.Fatalf("modifyCU status=%d body=%s", modifyResp.Code, modifyResp.Body.String())
	}
	requireJSONPresent(t, modifyResp.Body.Bytes(), "data.jobId")

	primaryAfterModifyResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/clusters/"+primaryClusterId, nil)
	if primaryAfterModifyResp.Code != http.StatusOK {
		t.Fatalf("primary after modify status=%d body=%s", primaryAfterModifyResp.Code, primaryAfterModifyResp.Body.String())
	}
	requireJSONValue(t, primaryAfterModifyResp.Body.Bytes(), "data.cuSize", float64(8))

	addSecondaryResp := performGlobalClusterRequest(t, router, http.MethodPost, "/v2/globalClusters/"+globalClusterId+"/secondaryClusters", map[string]any{
		"secondaryClusters": []map[string]any{
			{
				"clusterName": "secondary-cluster-ap",
				"regionId":    "aws-eu-west-1",
			},
		},
	})
	if addSecondaryResp.Code != http.StatusOK {
		t.Fatalf("add secondary status=%d body=%s", addSecondaryResp.Code, addSecondaryResp.Body.String())
	}
	requireJSONPresent(t, addSecondaryResp.Body.Bytes(), "data.jobId")

	describeAfterAddResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/globalClusters/"+globalClusterId, nil)
	if describeAfterAddResp.Code != http.StatusOK {
		t.Fatalf("describe after add status=%d body=%s", describeAfterAddResp.Code, describeAfterAddResp.Body.String())
	}
	requireJSONValue(t, describeAfterAddResp.Body.Bytes(), "data.regionIds.2", "aws-eu-west-1")
	requireJSONValue(t, describeAfterAddResp.Body.Bytes(), "data.clusters.2.role", "SECONDARY")

	deletePrimaryViaSecondaryResp := performGlobalClusterRequest(t, router, http.MethodDelete, "/v2/globalClusters/"+globalClusterId+"/secondaryClusters/"+primaryClusterId, nil)
	if deletePrimaryViaSecondaryResp.Code != http.StatusBadRequest {
		t.Fatalf("delete primary via secondary API status=%d body=%s", deletePrimaryViaSecondaryResp.Code, deletePrimaryViaSecondaryResp.Body.String())
	}

	deletePrimaryBeforeSecondariesResp := performGlobalClusterRequest(t, router, http.MethodDelete, "/v2/globalClusters/"+globalClusterId+"/clusters/"+primaryClusterId, nil)
	if deletePrimaryBeforeSecondariesResp.Code != http.StatusBadRequest {
		t.Fatalf("delete primary before secondaries status=%d body=%s", deletePrimaryBeforeSecondariesResp.Code, deletePrimaryBeforeSecondariesResp.Body.String())
	}

	deleteSecondaryResp := performGlobalClusterRequest(t, router, http.MethodDelete, "/v2/globalClusters/"+globalClusterId+"/clusters/"+secondaryClusterId, nil)
	if deleteSecondaryResp.Code != http.StatusOK {
		t.Fatalf("delete secondary status=%d body=%s", deleteSecondaryResp.Code, deleteSecondaryResp.Body.String())
	}
	requireJSONValue(t, deleteSecondaryResp.Body.Bytes(), "data.globalClusterId", globalClusterId)
	requireJSONValue(t, deleteSecondaryResp.Body.Bytes(), "data.clusterId", secondaryClusterId)

	deletedSecondaryResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/clusters/"+secondaryClusterId, nil)
	if deletedSecondaryResp.Code != http.StatusNotFound {
		t.Fatalf("deleted secondary status=%d body=%s", deletedSecondaryResp.Code, deletedSecondaryResp.Body.String())
	}

	removeWithSecondaryResp := performGlobalClusterRequest(t, router, http.MethodPost, "/v2/globalClusters/"+globalClusterId+"/removeGlobalEndpoint", nil)
	if removeWithSecondaryResp.Code != http.StatusBadRequest {
		t.Fatalf("remove global endpoint with secondary status=%d body=%s", removeWithSecondaryResp.Code, removeWithSecondaryResp.Body.String())
	}

	describeBeforeFinalDeleteResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/globalClusters/"+globalClusterId, nil)
	if describeBeforeFinalDeleteResp.Code != http.StatusOK {
		t.Fatalf("describe before final delete status=%d body=%s", describeBeforeFinalDeleteResp.Code, describeBeforeFinalDeleteResp.Body.String())
	}
	newSecondaryClusterId := requireJSONString(t, describeBeforeFinalDeleteResp.Body.Bytes(), "data.clusters.1.clusterId")
	deleteNewSecondaryResp := performGlobalClusterRequest(t, router, http.MethodDelete, "/v2/globalClusters/"+globalClusterId+"/clusters/"+newSecondaryClusterId, nil)
	if deleteNewSecondaryResp.Code != http.StatusOK {
		t.Fatalf("delete new secondary status=%d body=%s", deleteNewSecondaryResp.Code, deleteNewSecondaryResp.Body.String())
	}

	deletePrimaryResp := performGlobalClusterRequest(t, router, http.MethodDelete, "/v2/globalClusters/"+globalClusterId+"/clusters/"+primaryClusterId, nil)
	if deletePrimaryResp.Code != http.StatusOK {
		t.Fatalf("delete primary status=%d body=%s", deletePrimaryResp.Code, deletePrimaryResp.Body.String())
	}
	requireJSONValue(t, deletePrimaryResp.Body.Bytes(), "data.globalClusterId", globalClusterId)
	requireJSONValue(t, deletePrimaryResp.Body.Bytes(), "data.clusterId", primaryClusterId)

	missingGlobalResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/globalClusters/"+globalClusterId, nil)
	if missingGlobalResp.Code != http.StatusNotFound {
		t.Fatalf("missing global status=%d body=%s", missingGlobalResp.Code, missingGlobalResp.Body.String())
	}

	primaryAfterDeleteResp := performGlobalClusterRequest(t, router, http.MethodGet, "/v2/clusters/"+primaryClusterId, nil)
	if primaryAfterDeleteResp.Code != http.StatusNotFound {
		t.Fatalf("primary after delete status=%d body=%s", primaryAfterDeleteResp.Code, primaryAfterDeleteResp.Body.String())
	}
}

func newGlobalClusterTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v2 := router.Group("/v2")
	clusters := v2.Group("/clusters")
	{
		clusters.GET("/:clusterId", GetCluster)
	}
	globalClusters := v2.Group("/globalClusters")
	{
		globalClusters.GET("", ListGlobalClusters)
		globalClusters.POST("/create", CreateGlobalCluster)
		globalClusters.GET("/:globalClusterId", DescribeGlobalCluster)
		globalClusters.POST("/:globalClusterId/modifyCU", ModifyGlobalClusterCU)
		globalClusters.POST("/:globalClusterId/removeGlobalEndpoint", RemoveGlobalEndpoint)
		globalClusters.POST("/:globalClusterId/secondaryClusters", AddSecondaryClusters)
		globalClusters.DELETE("/:globalClusterId/clusters/:clusterId", DeleteCluster)
		globalClusters.DELETE("/:globalClusterId/secondaryClusters/:clusterId", DeleteSecondaryCluster)
	}
	return router
}

func performGlobalClusterRequest(t *testing.T, router *gin.Engine, method string, target string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var requestBody bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&requestBody).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &requestBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func requireJSONString(t *testing.T, raw []byte, path string) string {
	t.Helper()

	value := valueAtJSONPath(t, raw, path)
	text, ok := value.(string)
	if !ok || text == "" {
		t.Fatalf("%s=%v (%T), want non-empty string", path, value, value)
	}
	return text
}

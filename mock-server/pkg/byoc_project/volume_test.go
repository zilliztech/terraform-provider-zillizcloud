package byoc_project

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVolumeLifecycle(t *testing.T) {
	volumeStore = NewSafeStore[Volume]()

	router := newVolumeTestRouter()

	createBody := map[string]any{
		"projectId":            "proj-1",
		"regionId":             "aws-us-west-2",
		"volumeName":           "external-volume",
		"type":                 "EXTERNAL",
		"storageIntegrationId": "si-1",
		"path":                 "datasets/",
	}
	createResp := performVolumeRequest(t, router, http.MethodPost, "/v2/volumes/create", createBody)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	requireJSONValue(t, createResp.Body.Bytes(), "data.volumeName", "external-volume")

	describeResp := performVolumeRequest(t, router, http.MethodGet, "/v2/volumes/external-volume", nil)
	if describeResp.Code != http.StatusOK {
		t.Fatalf("describe status=%d body=%s", describeResp.Code, describeResp.Body.String())
	}
	requireJSONValue(t, describeResp.Body.Bytes(), "data.projectId", "proj-1")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.regionId", "aws-us-west-2")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.volumeName", "external-volume")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.type", "EXTERNAL")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.storageIntegrationId", "si-1")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.path", "datasets/")
	requireJSONValue(t, describeResp.Body.Bytes(), "data.status", "Available")
	requireJSONPresent(t, describeResp.Body.Bytes(), "data.createTime")

	listResp := performVolumeRequest(t, router, http.MethodGet, "/v2/volumes?projectId=proj-1&currentPage=1&pageSize=10&type=EXTERNAL", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	requireJSONValue(t, listResp.Body.Bytes(), "data.count", float64(1))
	requireJSONValue(t, listResp.Body.Bytes(), "data.currentPage", float64(1))
	requireJSONValue(t, listResp.Body.Bytes(), "data.pageSize", float64(10))
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.projectId", "proj-1")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.regionId", "aws-us-west-2")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.volumeName", "external-volume")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.type", "EXTERNAL")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.storageIntegrationId", "si-1")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.path", "datasets/")
	requireJSONValue(t, listResp.Body.Bytes(), "data.volumes.0.status", "Available")
	requireJSONPresent(t, listResp.Body.Bytes(), "data.volumes.0.createTime")

	deleteResp := performVolumeRequest(t, router, http.MethodDelete, "/v2/volumes/external-volume", nil)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	requireJSONValue(t, deleteResp.Body.Bytes(), "data.volumeName", "external-volume")

	missingResp := performVolumeRequest(t, router, http.MethodGet, "/v2/volumes/external-volume", nil)
	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("describe after delete status=%d body=%s", missingResp.Code, missingResp.Body.String())
	}
}

func TestCreateVolumeDefaultsTypeToManaged(t *testing.T) {
	volumeStore = NewSafeStore[Volume]()

	router := newVolumeTestRouter()
	createBody := map[string]any{
		"projectId":  "proj-1",
		"regionId":   "aws-us-west-2",
		"volumeName": "managed-volume",
	}

	resp := performVolumeRequest(t, router, http.MethodPost, "/v2/volumes/create", createBody)
	if resp.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", resp.Code, resp.Body.String())
	}

	describeResp := performVolumeRequest(t, router, http.MethodGet, "/v2/volumes/managed-volume", nil)
	if describeResp.Code != http.StatusOK {
		t.Fatalf("describe status=%d body=%s", describeResp.Code, describeResp.Body.String())
	}
	requireJSONValue(t, describeResp.Body.Bytes(), "data.type", "MANAGED")
}

func newVolumeTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v2 := router.Group("/v2")
	volumes := v2.Group("/volumes")
	{
		volumes.GET("", ListVolumes)
		volumes.POST("/create", CreateVolume)
		volumes.GET("/:volumeName", DescribeVolume)
		volumes.DELETE("/:volumeName", DeleteVolume)
	}
	return router
}

func performVolumeRequest(t *testing.T, router *gin.Engine, method string, target string, body any) *httptest.ResponseRecorder {
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

func requireJSONValue(t *testing.T, raw []byte, path string, want any) {
	t.Helper()

	got := valueAtJSONPath(t, raw, path)
	if got != want {
		t.Fatalf("%s=%v (%T), want %v (%T)", path, got, got, want, want)
	}
}

func requireJSONPresent(t *testing.T, raw []byte, path string) {
	t.Helper()

	got := valueAtJSONPath(t, raw, path)
	if got == nil || got == "" {
		t.Fatalf("%s is empty", path)
	}
}

func requireJSONMissing(t *testing.T, raw []byte, path string) {
	t.Helper()

	got := valueAtJSONPath(t, raw, path)
	if got != nil {
		t.Fatalf("%s=%v (%T), want missing", path, got, got)
	}
}

func valueAtJSONPath(t *testing.T, raw []byte, path string) any {
	t.Helper()

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	current := payload
	for _, segment := range bytes.Split([]byte(path), []byte(".")) {
		key := string(segment)
		if index := parsePathIndex(key); index >= 0 {
			items, ok := current.([]any)
			if !ok {
				t.Fatalf("%s is not an array", key)
			}
			if index >= len(items) {
				t.Fatalf("%s index %d out of range", key, index)
			}
			current = items[index]
			continue
		}

		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("%s is not an object", key)
		}
		current = object[key]
	}

	return current
}

func parsePathIndex(segment string) int {
	if len(segment) != 1 || segment[0] < '0' || segment[0] > '9' {
		return -1
	}
	return int(segment[0] - '0')
}

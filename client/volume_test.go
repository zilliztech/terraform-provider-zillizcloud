package client

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestUnitListVolumes(t *testing.T) {
	t.Run("defaults pagination and sends type filter", func(t *testing.T) {
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.Method != "GET" {
				t.Errorf("method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes" {
				t.Errorf("path=%s", req.URL.Path)
			}
			expectedQuery := url.Values{
				"projectId":   []string{"proj-1"},
				"currentPage": []string{"1"},
				"pageSize":    []string{"10"},
				"type":        []string{"MANAGED"},
			}
			if !reflect.DeepEqual(req.URL.Query(), expectedQuery) {
				t.Errorf("query=%+v", req.URL.Query())
			}
			return jsonResponse(t, map[string]any{
				"code": 0,
				"data": map[string]any{
					"volumes": []map[string]any{
						{"volumeName": "managed-volume", "type": "MANAGED"},
					},
					"count": 1, "currentPage": 1, "pageSize": 10,
				},
			}), nil
		})

		volumes, page, err := c.ListVolumes("proj-1", 0, 0, "MANAGED")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !reflect.DeepEqual(volumes, []VolumeSummary{{VolumeName: "managed-volume", Type: "MANAGED"}}) {
			t.Errorf("volumes=%+v", volumes)
		}
		if page.Count != 1 || page.CurrentPage != 1 || page.PageSize != 10 {
			t.Errorf("page=%+v", page)
		}
	})

	t.Run("uses explicit pagination and omits empty type filter", func(t *testing.T) {
		c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
			if req.Method != "GET" {
				t.Errorf("method=%s", req.Method)
			}
			if req.URL.Path != "/v2/volumes" {
				t.Errorf("path=%s", req.URL.Path)
			}
			expectedQuery := url.Values{
				"projectId":   []string{"proj-1"},
				"currentPage": []string{"3"},
				"pageSize":    []string{"25"},
			}
			if !reflect.DeepEqual(req.URL.Query(), expectedQuery) {
				t.Errorf("query=%+v", req.URL.Query())
			}
			return jsonResponse(t, map[string]any{
				"code": 0,
				"data": map[string]any{
					"volumes": []map[string]any{},
					"count":   0, "currentPage": 3, "pageSize": 25,
				},
			}), nil
		})

		volumes, page, err := c.ListVolumes("proj-1", 3, 25, "")
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if len(volumes) != 0 {
			t.Errorf("volumes=%+v", volumes)
		}
		if page.Count != 0 || page.CurrentPage != 3 || page.PageSize != 25 {
			t.Errorf("page=%+v", page)
		}
	})
}

func TestUnitDescribeVolume(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.EscapedPath() != "/v2/volumes/external%2Fvolume" {
			t.Errorf("path=%s", req.URL.EscapedPath())
		}
		if req.URL.RawQuery != "" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{
				"volumeName":           "external/volume",
				"type":                 "EXTERNAL",
				"regionId":             "aws-us-west-2",
				"storageIntegrationId": "si-1",
				"path":                 "datasets/",
				"status":               "Available",
				"createTime":           "2026-04-30T10:00:00Z",
			},
		}), nil
	})

	volume, err := c.DescribeVolume("external/volume")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected := &DescribeVolumeData{
		VolumeName:           "external/volume",
		Type:                 "EXTERNAL",
		RegionID:             "aws-us-west-2",
		StorageIntegrationID: "si-1",
		Path:                 "datasets/",
		Status:               "Available",
		CreateTime:           "2026-04-30T10:00:00Z",
	}
	if !reflect.DeepEqual(volume, expected) {
		t.Errorf("volume=%+v", volume)
	}
}

func TestUnitCreateManagedVolume(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/volumes/create" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.RawQuery != "" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		expected := map[string]any{
			"projectId":  "proj-1",
			"regionId":   "aws-us-west-2",
			"volumeName": "managed-volume",
			"type":       "MANAGED",
		}
		if !reflect.DeepEqual(body, expected) {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"volumeName": "managed-volume"},
		}), nil
	})

	volume, err := c.CreateVolume(&CreateVolumeRequest{
		ProjectID:  "proj-1",
		RegionID:   "aws-us-west-2",
		VolumeName: "managed-volume",
		Type:       "MANAGED",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if volume.VolumeName != "managed-volume" {
		t.Errorf("volume=%+v", volume)
	}
}

func TestUnitCreateExternalVolume(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "POST" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/volumes/create" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.RawQuery != "" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		expected := map[string]any{
			"projectId":            "proj-1",
			"regionId":             "aws-us-west-2",
			"volumeName":           "external-volume",
			"type":                 "EXTERNAL",
			"storageIntegrationId": "si-1",
			"path":                 "datasets/",
		}
		if !reflect.DeepEqual(body, expected) {
			t.Errorf("body=%+v", body)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"volumeName": "external-volume"},
		}), nil
	})

	volume, err := c.CreateVolume(&CreateVolumeRequest{
		ProjectID:            "proj-1",
		RegionID:             "aws-us-west-2",
		VolumeName:           "external-volume",
		Type:                 "EXTERNAL",
		StorageIntegrationID: "si-1",
		Path:                 "datasets/",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if volume.VolumeName != "external-volume" {
		t.Errorf("volume=%+v", volume)
	}
}

func TestUnitDeleteVolume(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "DELETE" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.EscapedPath() != "/v2/volumes/managed%2Fvolume" {
			t.Errorf("path=%s", req.URL.EscapedPath())
		}
		if req.URL.RawQuery != "" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{"volumeName": "managed/volume"},
		}), nil
	})

	volume, err := c.DeleteVolume("managed/volume")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if volume.VolumeName != "managed/volume" {
		t.Errorf("volume=%+v", volume)
	}
}

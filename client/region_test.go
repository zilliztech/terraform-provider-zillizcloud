package client

import (
	"net/http"
	"testing"
)

func TestClient_BaseUrlFromCloudRegion(t *testing.T) {
	testCases := []struct {
		cloudRegionId string
		want          string
	}{
		{"aws-us-east-2", "https://api.cloud.zilliz.com/v2"},
		{"aws-us-east-1", "https://api.cloud.zilliz.com/v2"},
		{"az-eastus", "https://api.cloud.zilliz.com/v2"},
		{"ali-cn-beijing", "https://api.cloud.zilliz.com.cn/v2"},
		{"tc-ap-shanghai", "https://api.cloud.zilliz.com.cn/v2"},
		{"unknown-eastus", "https://api.cloud.zilliz.com/v2"},
		{"unknown", "https://api.cloud.zilliz.com/v2"},
	}

	for _, tc := range testCases {
		t.Run(tc.cloudRegionId, func(t *testing.T) {
			got := BaseUrlFrom(tc.cloudRegionId)
			if got != tc.want {
				t.Errorf("want = %s, got = %s", tc.want, got)
			}
		})
	}

}

func TestUnitListCloudRegionsWithoutCloudId(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/regions" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.RawQuery != "" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": []map[string]any{
				{
					"cloudId":               "aws",
					"regionId":              "aws-us-west-2",
					"domain":                "api.cloud.zilliz.com",
					"supportedClusterTypes": []string{"free", "serverless", "dedicated"},
				},
			},
		}), nil
	})

	got, err := c.ListCloudRegions("")
	if err != nil {
		t.Fatalf("failed to ListCloudRegions: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Domain != "api.cloud.zilliz.com" {
		t.Errorf("domain=%s", got[0].Domain)
	}
	if len(got[0].SupportedClusterTypes) != 3 || got[0].SupportedClusterTypes[0] != "free" {
		t.Errorf("supportedClusterTypes=%v", got[0].SupportedClusterTypes)
	}
}

func TestUnitListCloudRegionsWithCloudId(t *testing.T) {
	c := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/regions" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.Query().Get("cloudId") != "aws" {
			t.Errorf("cloudId=%s", req.URL.Query().Get("cloudId"))
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": []map[string]any{
				{
					"cloudId":               "aws",
					"regionId":              "aws-us-east-1",
					"domain":                "api.cloud.zilliz.com",
					"supportedClusterTypes": []string{"serverless", "dedicated"},
				},
			},
		}), nil
	})

	got, err := c.ListCloudRegions("aws")
	if err != nil {
		t.Fatalf("failed to ListCloudRegions: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].CloudId != "aws" || got[0].RegionId != "aws-us-east-1" {
		t.Errorf("region=%+v", got[0])
	}
}

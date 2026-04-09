package client

import (
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

func TestClient_ListCloudRegions(t *testing.T) {
	testCases := []struct {
		cloudId CloudId
		notWant int
	}{
		{GCP, 0},
		{AWS, 0},
		{Azure, 0},
	}

	for _, tc := range testCases {
		t.Run("ListCloudRegions via "+string(tc.cloudId), func(t *testing.T) {
			c, teardown := zillizClient[[]CloudRegion](t)
			defer teardown()

			got, err := c.ListCloudRegions(string(tc.cloudId))
			if err != nil {
				t.Fatalf("failed to ListCloudRegions: %v", err)
			}

			if len(got) == tc.notWant {
				t.Errorf("want > %d, got = %v", tc.notWant, got)
			}
		})
	}
}

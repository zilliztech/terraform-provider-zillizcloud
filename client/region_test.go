package client

import (
	"testing"
)

func TestClient_BaseUrlFromCloudRegion(t *testing.T) {
	testCases := []struct {
		cloudRegionId string
		want          string
	}{
		{"aws-us-east-2", "https://controller.api.aws-us-east-2.zillizcloud.com/v1/"},
		{"aws-us-east-1", "https://controller.api.aws-us-east-1.zillizcloud.com/v1/"},
		{"az-eastus", "https://controller.api.az-eastus.zillizcloud.com/v1/"},
		{"ali-cn-beijing", "https://controller.api.ali-cn-beijing.cloud.zilliz.com.cn/v1/"},
		{"tc-ap-shanghai", "https://controller.api.tc-ap-shanghai.cloud.zilliz.com.cn/v1/"},
		{"unknown-eastus", "https://controller.api.unknown-eastus.zillizcloud.com/v1/"},
		{"unknown", "https://controller.api.unknown.zillizcloud.com/v1/"},
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

package provider

import (
	"testing"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestApiBaseUrlFromCloudRegion(t *testing.T) {
	testCases := []struct {
		name string
		in   zilliz.CloudRegion
		want string
	}{
		{
			name: "uses old api base url when present",
			in:   zilliz.CloudRegion{ApiBaseUrl: "https://api.aws-us-west-2.zillizcloud.com/v2", Domain: "api.cloud.zilliz.com"},
			want: "https://api.aws-us-west-2.zillizcloud.com/v2",
		},
		{
			name: "builds from bare domain",
			in:   zilliz.CloudRegion{Domain: "api.cloud.zilliz.com"},
			want: "https://api.cloud.zilliz.com/v2",
		},
		{
			name: "builds from https domain",
			in:   zilliz.CloudRegion{Domain: "https://api.cloud.zilliz.com"},
			want: "https://api.cloud.zilliz.com/v2",
		},
		{
			name: "keeps versioned domain",
			in:   zilliz.CloudRegion{Domain: "https://api.cloud.zilliz.com/v2"},
			want: "https://api.cloud.zilliz.com/v2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := apiBaseUrlFromCloudRegion(tc.in)
			if got != tc.want {
				t.Errorf("want=%s, got=%s", tc.want, got)
			}
		})
	}
}

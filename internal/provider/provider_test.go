package provider

import "testing"

func TestProviderUserAgent(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "release version",
			version: "0.6.36",
			want:    "terraform-provider-zillizcloud/0.6.36",
		},
		{
			name: "empty version defaults to dev",
			want: "terraform-provider-zillizcloud/dev",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := providerUserAgent(tc.version); got != tc.want {
				t.Fatalf("providerUserAgent(%q) = %q, want %q", tc.version, got, tc.want)
			}
		})
	}
}

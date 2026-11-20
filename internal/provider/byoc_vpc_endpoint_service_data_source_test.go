package provider

import (
	"fmt"
	"testing"

	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestResolveByocVpcEndpointService(t *testing.T) {
	fallback := "projects/vdc-prod/regions/us-west1/serviceAttachments/zilliz-byoc-psc-dns"

	tests := []struct {
		name     string
		result   *zilliz.ByocVpcEndpointService
		err      error
		fallback string
		want     string
		wantErr  bool
	}{
		{
			name: "control plane value takes precedence",
			result: &zilliz.ByocVpcEndpointService{
				EndpointService: "projects/vdc-prod/regions/us-west1/serviceAttachments/custom-service",
			},
			fallback: fallback,
			want:     "projects/vdc-prod/regions/us-west1/serviceAttachments/custom-service",
		},
		{
			name:     "not found uses fallback",
			err:      fmt.Errorf("request failed: %w", zilliz.Error{Code: 404, Message: "Request object not found"}),
			fallback: fallback,
			want:     fallback,
		},
		{
			name:     "empty control plane value uses fallback",
			result:   &zilliz.ByocVpcEndpointService{},
			fallback: fallback,
			want:     fallback,
		},
		{
			name:    "not found without fallback remains an error",
			err:     zilliz.Error{Code: 404, Message: "Request object not found"},
			wantErr: true,
		},
		{
			name:     "other errors are not hidden",
			err:      zilliz.Error{Code: 500, Message: "Internal server error"},
			fallback: fallback,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveByocVpcEndpointService(tc.result, tc.err, tc.fallback)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveByocVpcEndpointService() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("resolveByocVpcEndpointService() = %q, want %q", got, tc.want)
			}
		})
	}
}

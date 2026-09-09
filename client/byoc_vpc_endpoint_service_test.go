package client

import (
	"net/http"
	"testing"
)

func TestUnitGetByocVpcEndpointService(t *testing.T) {
	client := newMockClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Errorf("method=%s", req.Method)
		}
		if req.URL.Path != "/v2/byoc/vpc-endpoint-service" {
			t.Errorf("path=%s", req.URL.Path)
		}
		if req.URL.Query().Get("cloudId") != "aws" || req.URL.Query().Get("region") != "ap-southeast-7" {
			t.Errorf("query=%s", req.URL.RawQuery)
		}
		return jsonResponse(t, map[string]any{
			"code": 0,
			"data": map[string]any{
				"cloudId": "aws", "region": "ap-southeast-7",
				"endpointService": "com.amazonaws.vpce.ap-southeast-7.vpce-svc-test",
			},
		}), nil
	})

	service, err := client.GetByocVpcEndpointService("aws", "ap-southeast-7")
	if err != nil {
		t.Fatalf("GetByocVpcEndpointService: %v", err)
	}
	if service.EndpointService != "com.amazonaws.vpce.ap-southeast-7.vpce-svc-test" {
		t.Fatalf("EndpointService=%q", service.EndpointService)
	}
}

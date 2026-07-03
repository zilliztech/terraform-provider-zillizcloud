package client

import (
	"encoding/json"
	"testing"
)

func TestCreateByocOpProjectRequestGCPParamJSON(t *testing.T) {
	request := CreateByocOpProjectRequest{
		CloudID:     GCP,
		RegionID:    "gcp-us-west1",
		ProjectId:   "proj-1",
		DataPlaneId: "dp-1",
		DeployType:  7,
		GCPParam: &GCPParam{
			GCPProjectID:      "customer-gcp-project",
			BucketID:          "zilliz-byoc-bucket",
			GKENodeSA:         "node@customer-gcp-project.iam.gserviceaccount.com",
			ManagementSA:      "maintenance@customer-gcp-project.iam.gserviceaccount.com",
			StorageSA:         "storage@customer-gcp-project.iam.gserviceaccount.com",
			GKEClusterName:    "zilliz-byoc-gke",
			VPCName:           "zilliz-byoc-vpc",
			PrimarySubnetName: "zilliz-byoc-primary",
			PodSubnetName:     "zilliz-byoc-pods",
			ServiceSubnetName: "zilliz-byoc-services",
			LBSubnetName:      "zilliz-byoc-lb",
			PSCEndpointIP:     stringPtr("10.10.0.10"),
			Zones:             []string{"us-west1-a", "us-west1-b"},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	gcpParam, ok := got["gcpParam"].(map[string]any)
	if !ok {
		t.Fatalf("missing gcpParam in request: %s", string(body))
	}

	want := map[string]string{
		"gcpProjectId":      "customer-gcp-project",
		"bucketId":          "zilliz-byoc-bucket",
		"gkeNodeSa":         "node@customer-gcp-project.iam.gserviceaccount.com",
		"managementSa":      "maintenance@customer-gcp-project.iam.gserviceaccount.com",
		"storageSa":         "storage@customer-gcp-project.iam.gserviceaccount.com",
		"gkeClusterName":    "zilliz-byoc-gke",
		"vpcName":           "zilliz-byoc-vpc",
		"primarySubnetName": "zilliz-byoc-primary",
		"podSubnetName":     "zilliz-byoc-pods",
		"serviceSubnetName": "zilliz-byoc-services",
		"lbSubnetName":      "zilliz-byoc-lb",
		"pscEndpointIp":     "10.10.0.10",
	}

	for key, value := range want {
		if gotValue := gcpParam[key]; gotValue != value {
			t.Fatalf("gcpParam.%s = %v, want %s", key, gotValue, value)
		}
	}

	zones, ok := gcpParam["zones"].([]any)
	if !ok || len(zones) != 2 || zones[0] != "us-west1-a" || zones[1] != "us-west1-b" {
		t.Fatalf("gcpParam.zones = %v", gcpParam["zones"])
	}

	if _, ok := got["awsParam"]; ok {
		t.Fatalf("awsParam should be omitted: %s", string(body))
	}
	if _, ok := got["azureParam"]; ok {
		t.Fatalf("azureParam should be omitted: %s", string(body))
	}
}

func stringPtr(value string) *string {
	return &value
}

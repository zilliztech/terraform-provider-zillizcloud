package client

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGcpCmekWireFields(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		param := GCPParam{GCPProjectID: "customer-project", StorageSA: "storage-gsa@customer-project.iam.gserviceaccount.com"}
		if enabled {
			param.CseSA = "cmek-gsa@customer-project.iam.gserviceaccount.com"
			param.DefaultGCPCseKeyName = "projects/key-project/locations/us-west1/keyRings/ring/cryptoKeys/key"
		}
		raw, err := json.Marshal(param)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		_, hasEmail := decoded["cseSa"]
		_, hasKey := decoded["defaultGcpCseKeyName"]
		if hasEmail != enabled || hasKey != enabled || decoded["storageSa"] != param.StorageSA {
			t.Fatalf("unexpected CMEK fields: %s", raw)
		}
		if enabled && (decoded["defaultGcpCseKeyName"] != param.DefaultGCPCseKeyName || decoded["cseSa"] != param.CseSA) {
			t.Fatalf("wrong CMEK fields: %s", raw)
		}
		if strings.Contains(string(raw), "awsCseRoleArn") {
			t.Fatalf("AWS configuration leaked into GCP: %s", raw)
		}
	}
}

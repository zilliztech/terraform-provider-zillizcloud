package client

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAzureCmekWireFields(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "legacy", true: "cmek"}[enabled], func(t *testing.T) {
			storage := AzureIdentityParam{ClientID: "storage-client", PrincipalID: "storage-principal", ResourceID: "storage-resource"}
			param := AzureParam{StorageAccountName: "storage", ContainerName: "container", StorageIdentities: []AzureIdentityParam{storage}}
			if enabled {
				param.CSEIdentity = &AzureIdentityParam{ClientID: "cse-client", PrincipalID: "cse-principal", ResourceID: "cse-resource"}
				param.AzureTenantID = "cse-tenant"
				param.DefaultAzureCseKeyIdentifier = "https://customer-kms.vault.azure.net/keys/cmek-key"
			}
			raw, err := json.Marshal(CreateByocOpProjectRequest{CloudID: Azure, AzureParam: &param})
			if err != nil {
				t.Fatal(err)
			}
			var wire map[string]any
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatal(err)
			}
			if _, ok := wire["awsParam"]; ok {
				t.Fatalf("AWS fields leaked: %s", raw)
			}
			if _, ok := wire["gcpParam"]; ok {
				t.Fatalf("GCP fields leaked: %s", raw)
			}
			azure, ok := wire["azureParam"].(map[string]any)
			if !ok {
				t.Fatalf("Azure configuration must be an object: %s", raw)
			}
			for _, field := range []string{"cseIdentity", "azureTenantId", "defaultAzureCseKeyIdentifier"} {
				if _, ok := azure[field]; ok != enabled {
					t.Fatalf("unexpected %s presence: %s", field, raw)
				}
			}
			var roundtrip CreateByocOpProjectRequest
			if err := json.Unmarshal(raw, &roundtrip); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(roundtrip.AzureParam, &param) {
				t.Fatalf("CMEK request changed on JSON round trip: %#v", roundtrip.AzureParam)
			}
			var described DescribeByocOpProjectResponse
			describedJSON, err := json.Marshal(map[string]any{"azureConfig": param})
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(describedJSON, &described); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(described.AzureConfig, param) {
				t.Fatalf("CMEK describe changed on JSON round trip: %#v", described.AzureConfig)
			}
		})
	}
}

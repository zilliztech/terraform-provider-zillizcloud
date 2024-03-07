package client

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClient_CreateCluster(t *testing.T) {
	type fields struct {
		CloudRegionId string
		HTTPClient    *http.Client
		baseUrl       string
		apiKey        string
		userAgent     string
	}
	type args struct {
		params CreateClusterParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CreateClusterResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				CloudRegionId: tt.fields.CloudRegionId,
				HTTPClient:    tt.fields.HTTPClient,
				baseUrl:       tt.fields.baseUrl,
				apiKey:        tt.fields.apiKey,
				userAgent:     tt.fields.userAgent,
			}
			got, err := c.CreateCluster(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

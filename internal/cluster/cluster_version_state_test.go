package cluster

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type failedClusterReadStore struct {
	ClusterStore
}

func (failedClusterReadStore) Get(context.Context, string) (*ClusterResourceModel, error) {
	return nil, errors.New("cluster is not visible yet")
}

func TestCreateRetainsKnownStateWhenClusterReadTimesOut(t *testing.T) {
	ctx := context.Background()
	r := &ClusterResource{store: failedClusterReadStore{}}
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)

	objectType, ok := schema.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("expected an object schema")
	}
	values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}
	values["id"] = tftypes.NewValue(tftypes.String, "in01-accepted")
	values["milvus_version"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	resp := resource.CreateResponse{State: tfsdk.State{
		Schema: schema.Schema,
		Raw:    tftypes.NewValue(objectType, values),
	}}
	var model ClusterResourceModel
	if diags := resp.State.Get(ctx, &model); diags.HasError() {
		t.Fatalf("read state: %v", diags)
	}

	waitCtx, cancel := context.WithCancel(ctx)
	cancel()
	r.tryBestUpdateStatesAfterCreation(waitCtx, &model, &model, &resp)

	var version, id types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("milvus_version"), &version)...)
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
	if resp.Diagnostics.HasError() || len(resp.Diagnostics.Warnings()) != 1 || id.ValueString() != "in01-accepted" || !version.IsNull() {
		t.Fatalf("expected a warning, retained cluster ID, and null Milvus version; got %v, %s, %s", resp.Diagnostics, id, version)
	}
}

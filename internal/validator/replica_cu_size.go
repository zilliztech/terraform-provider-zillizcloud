package validator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ReplicaCuSizeValidator validates that replica is not more than 1 when cu_size is less than 8.
type ReplicaCuSizeValidator struct{}

func (r ReplicaCuSizeValidator) Description(_ context.Context) string {
	return "Replica cannot be more than 1 when cu_size is less than 8."
}

func (r ReplicaCuSizeValidator) MarkdownDescription(_ context.Context) string {
	return "Replica cannot be more than 1 when cu_size is less than 8."
}

func (r ReplicaCuSizeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	replicaVal := req.ConfigValue.ValueInt64()

	// Get the cu_size value from the same configuration
	var cuSizeVal types.Int64
	diags := req.Config.GetAttribute(ctx, path.Root("cu_size"), &cuSizeVal)
	if diags.HasError() {
		return // Cannot validate if cu_size is not accessible
	}

	if cuSizeVal.IsNull() || cuSizeVal.IsUnknown() {
		return // Cannot validate if cu_size is not set
	}

	cuSize := cuSizeVal.ValueInt64()

	if cuSize < 8 && replicaVal > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid replica configuration",
			fmt.Sprintf("When cu_size (%d) is less than 8, replica cannot be more than 1. Current replica value: %d", cuSize, replicaVal),
		)
	}
}

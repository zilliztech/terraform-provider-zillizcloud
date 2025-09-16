package conv

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SliceToSet converts a slice of any supported type into a Terraform types.Set.
// Returns a null set when input is nil, empty set when input is empty.
func SliceToSet[T string | int64 | bool](src []T) types.Set {
	if src == nil {
		return types.SetNull(getAttrType[T]())
	}
	if len(src) == 0 {
		return types.SetValueMust(getAttrType[T](), []attr.Value{})
	}

	values := make([]attr.Value, len(src))
	for i, v := range src {
		values[i] = createAttrValue(v)
	}

	ret, _ := types.SetValue(getAttrType[T](), values)
	return ret
}

var (
	errNullSet = fmt.Errorf("null set")
)

func IsNullError(err error) bool {
	return err == errNullSet
}

// SetToSlice converts a types.Set into a slice of the specified type.
// Returns nil slice with error when input is null, empty slice when input is unknown.
func SetToSlice[T string | int64 | bool](ctx context.Context, src types.Set) ([]T, error) {
	if src.IsNull() {
		return nil, errNullSet
	}
	if src.IsUnknown() {
		return []T{}, nil
	}

	var result []T
	diags := src.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert set to slice: %s", diags.Errors()[0].Summary())
	}

	return result, nil
}

// Helper function to get the appropriate attr.Type for each supported type
func getAttrType[T string | int64 | bool]() attr.Type {
	var zero T
	switch any(zero).(type) {
	case string:
		return types.StringType
	case int64:
		return types.Int64Type
	case bool:
		return types.BoolType
	default:
		panic("unsupported type")
	}
}

// Helper function to create attr.Value from any supported type
func createAttrValue[T string | int64 | bool](v T) attr.Value {
	switch val := any(v).(type) {
	case string:
		return types.StringValue(val)
	case int64:
		return types.Int64Value(val)
	case bool:
		return types.BoolValue(val)
	default:
		panic("unsupported type")
	}
}

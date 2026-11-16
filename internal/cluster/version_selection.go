package cluster

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

var milvusVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.x$`)
var nonBlank = regexp.MustCompile(`\S`)

type versionSelectionModel struct {
	MilvusVersion types.String `tfsdk:"milvus_version"`
	Channel       types.String `tfsdk:"channel"`
}

func versionSelectionSchema(description string, channels ...string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"milvus_version": schema.StringAttribute{
				MarkdownDescription: "Milvus version series, such as `2.6.x`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(milvusVersionPattern, "Use a Milvus version series such as 2.6.x."),
				},
			},
			"channel": schema.StringAttribute{
				MarkdownDescription: "The channel of the selected Milvus version.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf(channels...)},
			},
		},
	}
}

func versionSelection(ctx context.Context, value types.Object) (*zilliz.ClusterVersionSelection, error) {
	if value.IsNull() {
		return nil, nil
	}
	if value.IsUnknown() {
		return nil, fmt.Errorf("version must be known before submitting the request")
	}
	var selection versionSelectionModel
	if diags := value.As(ctx, &selection, basetypes.ObjectAsOptions{}); diags.HasError() {
		return nil, fmt.Errorf("could not read version: %s", diags.Errors()[0].Detail())
	}
	if selection.MilvusVersion.ValueString() == "" || selection.Channel.ValueString() == "" {
		return nil, fmt.Errorf("milvus_version and channel must be known before submitting the request")
	}
	return &zilliz.ClusterVersionSelection{
		MilvusVersion: selection.MilvusVersion.ValueString(),
		Channel:       selection.Channel.ValueString(),
	}, nil
}

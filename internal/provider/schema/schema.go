package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// vmConfigSchema creates a reusable schema for VM configuration.
func vmConfigSchema(vmType string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: fmt.Sprintf("%s VM configuration", vmType),
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"vm": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Instance type for %s virtual machine", vmType),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"min_count": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("%s VM minimum instance count", vmType),
				Required:            true,
				// Note: min_count and max_count validation (min_count <= max_count)
				// should be implemented at the resource level using ValidateConfig
			},
			"max_count": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("%s VM maximum instance count", vmType),
				Required:            true,
				// Note: min_count and max_count validation (min_count <= max_count)
				// should be implemented at the resource level using ValidateConfig
			},
		},
	}
}

func coreVmConfigSchema(vmType string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: fmt.Sprintf("%s VM configuration", vmType),
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"vm": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Instance type for %s virtual machine", vmType),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"count": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("%s VM instance count", vmType),
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

var Instances = schema.SingleNestedAttribute{
	MarkdownDescription: "Instance type configuration",
	Required:            true,
	Attributes: map[string]schema.Attribute{
		"core":        coreVmConfigSchema("core"),
		"fundamental": vmConfigSchema("fundamental"),
		"search":      vmConfigSchema("search"),
		"index":       vmConfigSchema("index"),

		"auto_scaling": schema.BoolAttribute{
			MarkdownDescription: "Enable auto scaling for instances",
			Default:             booldefault.StaticBool(true),
			Computed:            true,
			Optional:            true,
		},
		"arch": schema.StringAttribute{
			MarkdownDescription: "Architecture type (X86 or ARM)",
			Default:             stringdefault.StaticString("X86"),
			Computed:            true,
			Optional:            true,

			Validators: []validator.String{
				stringvalidator.OneOf("X86", "ARM"),
			},
		},
	},
}

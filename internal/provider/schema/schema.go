package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var Instances = schema.SingleNestedAttribute{

	MarkdownDescription: "Instance type configuration",
	Required:            true,
	Attributes: map[string]schema.Attribute{
		"core_vm": schema.StringAttribute{
			MarkdownDescription: "Instance type used for the core virtual machine, which hosts Milvus Operators, Zilliz Cloud Agent, and Milvus dependencies, such as Prometheus, Etcd, Pulsar, etc. ",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("m6i.2xlarge"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"m6i.2xlarge",
				),
			},
		},
		"core_vm_min_count": schema.Int64Attribute{
			MarkdownDescription: "Core VM instance count. Defaults to 3 if not specified.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(3),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"fundamental_vm": schema.StringAttribute{
			MarkdownDescription: "Instance type used for the fundamental virtual machine, which hosts Milvus components other than the query nodes, including the proxy, datanode, index pool, and coordinators.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("m6i.2xlarge"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"m6i.2xlarge",
				),
			},
		},
		"fundamental_vm_min_count": schema.Int64Attribute{
			MarkdownDescription: "Fundamental VM instance count",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(0),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"search_vm": schema.StringAttribute{
			MarkdownDescription: "Instance type used for the search virtual machine, which hosts the query nodes.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("m6id.2xlarge"),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"m6id.2xlarge",
					"m6id.4xlarge",
				),
			},
		},
		"search_vm_min_count": schema.Int64Attribute{
			MarkdownDescription: "Search VM instance count",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(0),
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
	},
}

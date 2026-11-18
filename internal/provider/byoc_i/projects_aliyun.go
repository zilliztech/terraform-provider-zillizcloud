package byoc_op

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
)

type AliyunConfig struct {
	Region             types.String `tfsdk:"region"`
	AccountID          types.String `tfsdk:"account_id"`
	ACKClusterID       types.String `tfsdk:"ack_cluster_id"`
	VPCID              types.String `tfsdk:"vpc_id"`
	BucketID           types.String `tfsdk:"bucket_id"`
	NodeRoleARN        types.String `tfsdk:"node_role_arn"`
	MaintenanceRoleARN types.String `tfsdk:"maintenance_role_arn"`
	StorageRoleARN     types.String `tfsdk:"storage_role_arn"`
	OIDCProviderARN    types.String `tfsdk:"oidc_provider_arn"`
	EndpointID         types.String `tfsdk:"endpoint_id"`
	VSwitchIDs         types.Set    `tfsdk:"vswitch_ids"`
	SecurityGroupIDs   types.Set    `tfsdk:"security_group_ids"`
	Zones              types.Set    `tfsdk:"zones"`
}

func aliyunSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{Optional: true, MarkdownDescription: "Customer-owned Aliyun ACK resources for BYOC-I. Requires the Aliyun registration backend.", Attributes: map[string]schema.Attribute{
		"region":               schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^ali-[a-z0-9-]+$`), "must use the ali- region prefix")}, MarkdownDescription: "region"},
		"account_id":           schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9]{16}$`), "must be a 16-digit Alibaba Cloud account ID")}, MarkdownDescription: "account id"},
		"ack_cluster_id":       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "ack cluster id"},
		"vpc_id":               schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "vpc id"},
		"bucket_id":            schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "bucket id"},
		"node_role_arn":        schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "node role arn"},
		"maintenance_role_arn": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "maintenance role arn"},
		"storage_role_arn":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "storage role arn"},
		"oidc_provider_arn":    schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "oidc provider arn"},
		"endpoint_id":          schema.StringAttribute{Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, Validators: []validator.String{stringvalidator.LengthAtLeast(1)}, MarkdownDescription: "endpoint id"},
		"vswitch_ids":          schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))}, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, MarkdownDescription: "vswitch ids"},
		"security_group_ids":   schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))}, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, MarkdownDescription: "security group ids"},
		"zones":                schema.SetAttribute{Required: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))}, PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()}, MarkdownDescription: "zones"},
	}}
}

func (a *AliyunConfig) registration(ctx context.Context) (*zilliz.AliyunParam, error) {
	p := &zilliz.AliyunParam{
		AccountID:          a.AccountID.ValueString(),
		ACKClusterID:       a.ACKClusterID.ValueString(),
		VPCID:              a.VPCID.ValueString(),
		BucketID:           a.BucketID.ValueString(),
		NodeRoleARN:        a.NodeRoleARN.ValueString(),
		MaintenanceRoleARN: a.MaintenanceRoleARN.ValueString(),
		StorageRoleARN:     a.StorageRoleARN.ValueString(),
		OIDCProviderARN:    a.OIDCProviderARN.ValueString(),
		EndpointID:         optionalStringPointer(a.EndpointID),
	}
	if diags := a.VSwitchIDs.ElementsAs(ctx, &p.VSwitchIDs, false); diags.HasError() {
		return nil, fmt.Errorf("invalid Aliyun VSwitchIDs: %v", diags)
	}
	if diags := a.SecurityGroupIDs.ElementsAs(ctx, &p.SecurityGroupIDs, false); diags.HasError() {
		return nil, fmt.Errorf("invalid Aliyun SecurityGroupIDs: %v", diags)
	}
	if diags := a.Zones.ElementsAs(ctx, &p.Zones, false); diags.HasError() {
		return nil, fmt.Errorf("invalid Aliyun Zones: %v", diags)
	}
	return p, nil
}

func aliyunFromResponse(ctx context.Context, region string, p *zilliz.AliyunParam) (*AliyunConfig, error) {
	a := &AliyunConfig{
		Region:             types.StringValue(region),
		AccountID:          types.StringValue(p.AccountID),
		ACKClusterID:       types.StringValue(p.ACKClusterID),
		VPCID:              types.StringValue(p.VPCID),
		BucketID:           types.StringValue(p.BucketID),
		NodeRoleARN:        types.StringValue(p.NodeRoleARN),
		MaintenanceRoleARN: types.StringValue(p.MaintenanceRoleARN),
		StorageRoleARN:     types.StringValue(p.StorageRoleARN),
		OIDCProviderARN:    types.StringValue(p.OIDCProviderARN),
		EndpointID:         types.StringPointerValue(p.EndpointID),
	}
	valueVSwitchIDs, diagsVSwitchIDs := types.SetValueFrom(ctx, types.StringType, p.VSwitchIDs)
	if diagsVSwitchIDs.HasError() {
		return nil, fmt.Errorf("invalid returned Aliyun VSwitchIDs: %v", diagsVSwitchIDs)
	}
	a.VSwitchIDs = valueVSwitchIDs
	valueSecurityGroupIDs, diagsSecurityGroupIDs := types.SetValueFrom(ctx, types.StringType, p.SecurityGroupIDs)
	if diagsSecurityGroupIDs.HasError() {
		return nil, fmt.Errorf("invalid returned Aliyun SecurityGroupIDs: %v", diagsSecurityGroupIDs)
	}
	a.SecurityGroupIDs = valueSecurityGroupIDs
	valueZones, diagsZones := types.SetValueFrom(ctx, types.StringType, p.Zones)
	if diagsZones.HasError() {
		return nil, fmt.Errorf("invalid returned Aliyun Zones: %v", diagsZones)
	}
	a.Zones = valueZones
	return a, nil
}

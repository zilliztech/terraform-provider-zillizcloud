// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	zilliz "github.com/zilliztech/terraform-provider-zillizcloud/client"
	byoc "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider/byoc"
	byoc_op "github.com/zilliztech/terraform-provider-zillizcloud/internal/provider/byoc_i"

	cluster "github.com/zilliztech/terraform-provider-zillizcloud/internal/cluster"
)

// Ensure ZillizProvider satisfies various provider interfaces.
var _ provider.Provider = &ZillizProvider{}

// ZillizProvider defines the provider implementation.
type ZillizProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// zillizProviderModel describes the provider data model.
type zillizProviderModel struct {
	ApiKey      types.String  `tfsdk:"api_key"`
	RegionId    types.String  `tfsdk:"region_id"`
	HostAddress types.String  `tfsdk:"host_address"`
	Qps         types.Float64 `tfsdk:"qps"`
	Burst       types.Int64   `tfsdk:"burst"`
}

func (p *ZillizProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "zillizcloud"
	resp.Version = p.version
}

func (p *ZillizProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Zilliz Cloud API Key",
				Optional:            true,
				Sensitive:           true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "Zilliz Cloud Region Id",
				Optional:            true,
			},
			"host_address": schema.StringAttribute{
				MarkdownDescription: "Zilliz Cloud Host Address",
				Optional:            true,
			},
			"qps": schema.Float64Attribute{
				MarkdownDescription: "The maximum queries per second (QPS) to the Zilliz Cloud API for each resource. Defaults to 10.0.",
				Optional:            true,
			},
			"burst": schema.Int64Attribute{
				MarkdownDescription: "The maximum burst for throttle. Defaults to 10.",
				Optional:            true,
			},
		},
	}
}

func (p *ZillizProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data zillizProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := p.buildClientConfig(data, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := zilliz.NewClient(
		zilliz.WithApiKey(config.apiKey),
		zilliz.WithCloudRegionId(data.RegionId.ValueString()),
		zilliz.WithHostAddress(config.hostAddress),
		zilliz.WithRateLimiter(config.qps, config.burst),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

type clientConfig struct {
	apiKey      string
	hostAddress string
	qps         float64
	burst       int64
}

func (p *ZillizProvider) buildClientConfig(data zillizProviderModel, resp *provider.ConfigureResponse) clientConfig {
	config := clientConfig{
		apiKey:      getStringFromEnvOrConfig("ZILLIZCLOUD_API_KEY", data.ApiKey),
		hostAddress: getStringFromEnvOrConfig("ZILLIZCLOUD_HOST_ADDRESS", data.HostAddress),
		qps:         zilliz.DefaultQPS,
		burst:       int64(zilliz.DefaultQPS),
	}

	if qps, err := p.parseQPS(data.Qps); err != nil {
		resp.Diagnostics.AddError("failed to parse qps", err.Error())
	} else {
		config.qps = qps
	}

	if burst, err := p.parseBurst(data.Burst); err != nil {
		resp.Diagnostics.AddError("failed to parse burst", err.Error())
	} else {
		config.burst = burst
	}

	return config
}

func getStringFromEnvOrConfig(envVar string, configValue types.String) string {
	if !configValue.IsNull() {
		return configValue.ValueString()
	}
	return os.Getenv(envVar)
}

func (p *ZillizProvider) parseQPS(qpsValue types.Float64) (float64, error) {
	if qpsValue.IsNull() {
		if qpsEnv := os.Getenv("ZILLIZCLOUD_QPS"); qpsEnv != "" {
			return strconv.ParseFloat(qpsEnv, 64)
		}
		return zilliz.DefaultQPS, nil
	}
	return qpsValue.ValueFloat64(), nil
}

func (p *ZillizProvider) parseBurst(burstValue types.Int64) (int64, error) {
	if burstValue.IsNull() {
		if burstEnv := os.Getenv("ZILLIZCLOUD_BURST"); burstEnv != "" {
			return strconv.ParseInt(burstEnv, 10, 64)
		}
		return int64(zilliz.DefaultQPS), nil
	}
	return burstValue.ValueInt64(), nil
}

func (p *ZillizProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		cluster.NewClusterResource,
		byoc.NewBYOCProjectResource,
		byoc_op.NewBYOCOpProjectSettingsResource,
		byoc_op.NewBYOCOpProjectResource,
		byoc_op.NewBYOCOpProjectAgentResource,
		NewUserResource,
		NewUserRoleResource,
		NewDatabaseResource,
		NewCollectionResource,
		NewIndexResource,
		NewAliasResource,
		NewPartitionsResource,
	}
}

func (p *ZillizProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCloudProvidersDataSource,
		NewCloudRegionsDataSource,
		NewProjectDataSource,
		cluster.NewClustersDataSource,
		cluster.NewClusterDataSource,
		byoc.NewExternalIdDataSource,
		byoc_op.NewBYOCOpProjectSettingsData,
		NewUsersDataSource,
		NewRolesDataSource,
		NewDatabasesDataSource,
		NewCollectionsDataSource,
		NewIndexesDataSource,
		NewAliasesDataSource,
		NewPartitionsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ZillizProvider{
			version: version,
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"biot.com/terraform-provider-biot/internal/api"
	"biot.com/terraform-provider-biot/internal/resources/template"
)

type BiotProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type BiotProviderModel struct {
	BaseURL          string `tfsdk:"base_url"`
	ServiceID         string `tfsdk:"service_id"`
	ServiceSecretKey  string `tfsdk:"service_secret_key"`
}

func (p *BiotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "biot"
	resp.Version = p.version
}

func (p *BiotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Biot base URL",
				Required:            true, 
				Optional:            false,
			},
			"service_id": schema.StringAttribute{
				MarkdownDescription: "Terraform plugin service id.",
				Required:            true, 
				Optional:            false,
			},
			"service_secret_key": schema.StringAttribute{
				MarkdownDescription: "Terraform plugin service secret key.",
				Required:            true, 
				Optional:            false,
				Sensitive:           true,
			},
		},
	}
}


func (p *BiotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config BiotProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.BaseURL == "" {
		resp.Diagnostics.AddError(
			"Missing base_url",
			"Base URL must be configured in the provider.",
		)
		return
	}


	if config.ServiceID == "" {
		resp.Diagnostics.AddError(
			"Missing service_id",
			"Service ID must be configured in the provider.",
		)
		return
	}

	if config.ServiceSecretKey == "" {
		resp.Diagnostics.AddError(
			"Missing service_secret_key",
			"Service secret key must be configured in the provider.",
		)
		return
	}

	client := api.NewAPIClient(api.NewBiotSdkImpl(config.BaseURL), config.ServiceID, config.ServiceSecretKey)

	// Example client configuration for data sources and resources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *BiotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		template.NewResource,
	}
}

func (p *BiotProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{
		// NewExampleEphemeralResource,
	}
}

func (p *BiotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewExampleDataSource,
		// TODO: Add datasource for template IDs.
	}
}

func (p *BiotProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BiotProvider{
			version: version,
		}
	}
}

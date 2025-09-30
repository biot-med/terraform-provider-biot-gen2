// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	"biot.com/terraform-provider-biot-gen2/internal/resources/template"
	"biot.com/terraform-provider-biot-gen2/internal/version"
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
	ServiceID        string `tfsdk:"service_id"`
	ServiceSecretKey string `tfsdk:"service_secret_key"`
}

// Custom validators
type nonEmptyStringValidator struct{}

func (v nonEmptyStringValidator) Description(ctx context.Context) string {
	return "Ensures the string is not empty"
}

func (v nonEmptyStringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v nonEmptyStringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid empty string",
			"This field cannot be empty",
		)
	}
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
				Validators: []validator.String{
					nonEmptyStringValidator{},
				},
			},
			"service_id": schema.StringAttribute{
				MarkdownDescription: "Terraform plugin service id.",
				Required:            true,
				Optional:            false,
				Validators: []validator.String{
					nonEmptyStringValidator{},
				},
			},
			"service_secret_key": schema.StringAttribute{
				MarkdownDescription: "Terraform plugin service secret key.",
				Required:            true,
				Optional:            false,
				Sensitive:           true,
				Validators: []validator.String{
					nonEmptyStringValidator{},
				},
			},
		},
	}
}

func (p *BiotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Debug(ctx, "Starting provider configuration")

	var config BiotProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Provider configuration failed due to validation errors")
		return
	}

	biotSdk := api.NewBiotSdkImpl(config.BaseURL)
	authenticator := api.NewAuthenticatorService(biotSdk, config.ServiceID, config.ServiceSecretKey)

	client := api.NewAPIClient(biotSdk, authenticator)

	// Validate versions
	versionValidator := api.NewVersionValidator(biotSdk, authenticator)

	if err := versionValidator.Validate(ctx, p.version, version.MinimumBiotVersion); err != nil {
		switch e := err.(type) {
		case api.ValidationUnsupportedError:
			// 200 OK but status == UNSUPPORTED
			resp.Diagnostics.AddError("Versions Validation Failed", e.Error())
		case api.ValidationAPIError:
			// Non-200/transport error
			resp.Diagnostics.AddError("Error occurred while trying to validate version", e.Error())
		default:
			resp.Diagnostics.AddError("Error occurred while trying to validate version", err.Error())
		}
		return
	}

	tflog.Info(ctx, "Provider configuration completed successfully", map[string]interface{}{
		"base_url":   config.BaseURL,
		"service_id": config.ServiceID,
	})

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

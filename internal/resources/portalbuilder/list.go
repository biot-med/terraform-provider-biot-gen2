package portalbuilder

import (
	"context"
	"fmt"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	portalbuilderapi "biot.com/terraform-provider-biot-gen2/internal/api/portalbuilder"
)

var (
	_ list.ListResource              = &BiotPortalViewListResource{}
	_ list.ListResourceWithConfigure = &BiotPortalViewListResource{}
)

// NewListResource backs `list "biot_portal_view"` blocks in .tfquery.hcl files, so that
// `terraform query` can generate configuration from views designed in the BioT UI.
func NewListResource() list.ListResource {
	return &BiotPortalViewListResource{}
}

type BiotPortalViewListResource struct {
	client *api.APIClient
}

type viewListConfig struct {
	PortalType   types.String `tfsdk:"portal_type"`
	ViewType     types.String `tfsdk:"view_type"`
	TemplateName types.String `tfsdk:"template_name"`
}

func (r *BiotPortalViewListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_portal_view"
}

func (r *BiotPortalViewListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", "Expected *api.APIClient")
		return
	}

	r.client = client
}

func (r *BiotPortalViewListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists the portal builder views in BioT. All filters are optional.",
		Attributes: map[string]listschema.Attribute{
			"portal_type": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list views of this portal: " + quoted(portalbuilderapi.PortalTypes) + ".",
			},
			"view_type": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list views of this type: " + quoted(ViewTypes) + ".",
			},
			"template_name": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list the views of the template with this name.",
			},
		},
	}
}

func (r *BiotPortalViewListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		var diags diag.Diagnostics
		diags.AddError("Provider not configured",
			"The biot provider has no API client for listing portal views. This is a bug in the provider.")
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	var config viewListConfig
	if !req.Config.Raw.IsNull() {
		if diags := req.Config.Get(ctx, &config); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	portalTypes := portalbuilderapi.PortalTypes
	if known(config.PortalType) {
		portalTypes = []string{config.PortalType.ValueString()}
	}

	stream.Results = r.results(ctx, req, config, portalTypes)
}

// results lists the views of each portal, filtering locally - BioT's list API has no filters.
func (r *BiotPortalViewListResource) results(ctx context.Context, req list.ListRequest, config viewListConfig, portalTypes []string) iter.Seq[list.ListResult] {
	return func(push func(list.ListResult) bool) {
		var pushed int64
		templates := newResolver(r.client)

		for _, portalType := range portalTypes {
			views, err := r.client.PortalBuilder.ListViews(ctx, portalType)
			if err != nil {
				var diags diag.Diagnostics
				diags.AddError("API Error", fmt.Sprintf("Failed to list the views of %s: %s", portalType, err))
				push(list.ListResult{Diagnostics: diags})
				return
			}
			templates.remember(views)

			for _, view := range views {
				if known(config.ViewType) && view.Type != config.ViewType.ValueString() {
					continue
				}
				if known(config.TemplateName) && templateName(view) != config.TemplateName.ValueString() {
					continue
				}

				if !push(r.result(ctx, req, view, templates)) {
					return
				}

				pushed++
				if req.Limit > 0 && pushed >= req.Limit {
					return
				}
			}
		}
	}
}

func (r *BiotPortalViewListResource) result(ctx context.Context, req list.ListRequest, view portalbuilderapi.View, templates *resolver) list.ListResult {
	result := req.NewListResult(ctx)
	result.DisplayName = describeView(view)
	result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("portal_type"), view.PortalType)...)
	result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("id"), view.ID)...)

	// The layout needs one more call per view, so fetch it only when generating config.
	if req.IncludeResource {
		full, err := r.client.PortalBuilder.GetView(ctx, view.PortalType, view.ID)
		if err != nil {
			result.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read portal view %s: %s", view.ID, err))
			return result
		}

		model, diags := mapViewToTerraformModel(ctx, full, templates)
		result.Diagnostics.Append(diags...)
		if !diags.HasError() {
			result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
		}
	}

	return result
}

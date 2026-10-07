package action

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
	"biot.com/terraform-provider-biot-gen2/internal/resources/abac"
)

var (
	_ list.ListResource              = &BiotAbacActionListResource{}
	_ list.ListResourceWithConfigure = &BiotAbacActionListResource{}
)

// NewListResource backs `list "biot_abac_action"` blocks in .tfquery.hcl files, so that
// `terraform query` can find existing actions and generate the configuration to import them.
func NewListResource() list.ListResource {
	return &BiotAbacActionListResource{}
}

type BiotAbacActionListResource struct {
	client *api.APIClient
}

type actionListConfig struct {
	Value   types.String `tfsdk:"value"`
	BuiltIn types.Bool   `tfsdk:"built_in"`
}

func (r *BiotAbacActionListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_abac_action"
}

func (r *BiotAbacActionListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BiotAbacActionListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists the access-control (ABAC) actions in BioT. Static (`CLASS`) " +
			"actions are never listed: BioT ships them, they cannot be created, and they take no " +
			"params. All filters are optional.",
		Attributes: map[string]listschema.Attribute{
			"value": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list actions with this implementation, for example `AddSearchFilterAction`.",
			},
			"built_in": listschema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "`false` lists only your own actions, leaving out the ones BioT ships; " +
					"`true` lists only the ones BioT ships. Omit it to list both.",
			},
		},
	}
}

func (r *BiotAbacActionListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = abac.ListNotConfigured(Entity)
		return
	}

	var config actionListConfig
	if !req.Config.Raw.IsNull() {
		if diags := req.Config.Get(ctx, &config); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	// type and value are filtered by the service. built_in cannot be - action search has no
	// tags filter - so it is applied locally as results arrive.
	filter := map[string]transport.SearchFilter{
		"type": abac.ExcludeStaticTypes(),
	}
	if !config.Value.IsNull() && !config.Value.IsUnknown() {
		filter["value"] = transport.SearchFilter{Eq: config.Value.ValueString()}
	}

	stream.Results = abac.ListResults(ctx, req, Entity, abac.ListSpec[abacapi.ActionResponse]{
		Items: r.client.Abac.SearchActions(ctx, filter),
		Keep: func(action abacapi.ActionResponse) bool {
			return abac.MatchesBuiltIn(config.BuiltIn, action.Tags)
		},
		ID: func(action abacapi.ActionResponse) string { return action.ID },
		DisplayName: func(action abacapi.ActionResponse) string {
			return abac.DescribeObject(action.Value, action.Description)
		},
		Model: func(ctx context.Context, action abacapi.ActionResponse) (any, diag.Diagnostics) {
			return MapActionResponseToTerraformModel(ctx, action)
		},
	})
}

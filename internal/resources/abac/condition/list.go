package condition

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
	_ list.ListResource              = &BiotAbacConditionListResource{}
	_ list.ListResourceWithConfigure = &BiotAbacConditionListResource{}
)

// NewListResource backs `list "biot_abac_condition"` blocks in .tfquery.hcl files, so that
// `terraform query` can find existing conditions and generate the configuration to import them.
func NewListResource() list.ListResource {
	return &BiotAbacConditionListResource{}
}

type BiotAbacConditionListResource struct {
	client *api.APIClient
}

type conditionListConfig struct {
	Value   types.String `tfsdk:"value"`
	BuiltIn types.Bool   `tfsdk:"built_in"`
}

func (r *BiotAbacConditionListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_abac_condition"
}

func (r *BiotAbacConditionListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BiotAbacConditionListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists the access-control (ABAC) conditions in BioT. Static (`CLASS`) " +
			"conditions are never listed: BioT ships them, they cannot be created, and they take no " +
			"params. All filters are optional.",
		Attributes: map[string]listschema.Attribute{
			"value": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list conditions with this implementation, for example `InitiatorAttributeInParamsCondition`.",
			},
			"built_in": listschema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "`false` lists only your own conditions, leaving out the ones BioT ships; " +
					"`true` lists only the ones BioT ships. Omit it to list both.",
			},
		},
	}
}

func (r *BiotAbacConditionListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = abac.ListNotConfigured(Entity)
		return
	}

	var config conditionListConfig
	if !req.Config.Raw.IsNull() {
		if diags := req.Config.Get(ctx, &config); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	// type and value are filtered by the service. built_in cannot be - condition search has no
	// tags filter - so it is applied locally as results arrive.
	filter := map[string]transport.SearchFilter{
		"type": abac.ExcludeStaticTypes(),
	}
	if !config.Value.IsNull() && !config.Value.IsUnknown() {
		filter["value"] = transport.SearchFilter{Eq: config.Value.ValueString()}
	}

	stream.Results = abac.ListResults(ctx, req, Entity, abac.ListSpec[abacapi.ConditionResponse]{
		Items: r.client.Abac.SearchConditions(ctx, filter),
		Keep: func(condition abacapi.ConditionResponse) bool {
			return abac.MatchesBuiltIn(config.BuiltIn, condition.Tags)
		},
		ID: func(condition abacapi.ConditionResponse) string { return condition.ID },
		DisplayName: func(condition abacapi.ConditionResponse) string {
			return abac.DescribeObject(condition.Value, condition.Description)
		},
		Model: func(ctx context.Context, condition abacapi.ConditionResponse) (any, diag.Diagnostics) {
			return MapConditionResponseToTerraformModel(ctx, condition)
		},
	})
}

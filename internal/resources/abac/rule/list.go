package rule

import (
	"context"
	"fmt"
	"strings"

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
	_ list.ListResource              = &BiotAbacRuleListResource{}
	_ list.ListResourceWithConfigure = &BiotAbacRuleListResource{}
)

// NewListResource backs `list "biot_abac_rule"` blocks in .tfquery.hcl files, so that
// `terraform query` can find existing rules and generate the configuration to import them.
func NewListResource() list.ListResource {
	return &BiotAbacRuleListResource{}
}

type BiotAbacRuleListResource struct {
	client *api.APIClient
}

type ruleListConfig struct {
	APIID       types.String `tfsdk:"api_id"`
	ActionID    types.String `tfsdk:"action_id"`
	ConditionID types.String `tfsdk:"condition_id"`
	BuiltIn     types.Bool   `tfsdk:"built_in"`
}

func (r *BiotAbacRuleListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_abac_rule"
}

func (r *BiotAbacRuleListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BiotAbacRuleListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		MarkdownDescription: "Lists the access-control (ABAC) rules in BioT. All filters are optional, " +
			"and a rule must match all of the ones that are set.",
		Attributes: map[string]listschema.Attribute{
			"api_id": listschema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Only list rules that run on this API, written like `api_id` on the " +
					"rule - for example `GET/organization/v1/users/patients`.",
			},
			"action_id": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list rules that run this action.",
			},
			"condition_id": listschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Only list rules that use this condition.",
			},
			"built_in": listschema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "`false` lists only your own rules, leaving out the ones BioT ships; " +
					"`true` lists only the ones BioT ships. Omit it to list both.",
			},
		},
	}
}

func (r *BiotAbacRuleListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = abac.ListNotConfigured(Entity)
		return
	}

	var config ruleListConfig
	if !req.Config.Raw.IsNull() {
		if diags := req.Config.Get(ctx, &config); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	// api_id, action_id and condition_id are filtered by the service, which resolves them to the
	// rules that reference them. built_in is applied locally, as for conditions and actions.
	filter := map[string]transport.SearchFilter{}
	addEq := func(name string, value types.String) {
		if !value.IsNull() && !value.IsUnknown() {
			filter[name] = transport.SearchFilter{Eq: value.ValueString()}
		}
	}
	addEq("apiId", config.APIID)
	addEq("actionId", config.ActionID)
	addEq("conditionId", config.ConditionID)

	stream.Results = abac.ListResults(ctx, req, Entity, abac.ListSpec[abacapi.RuleResponse]{
		Items: r.client.Abac.SearchRules(ctx, filter),
		Keep: func(rule abacapi.RuleResponse) bool {
			return abac.MatchesBuiltIn(config.BuiltIn, rule.Tags)
		},
		ID:          func(rule abacapi.RuleResponse) string { return rule.ID },
		DisplayName: describeRule,
		Model: func(ctx context.Context, rule abacapi.RuleResponse) (any, diag.Diagnostics) {
			return MapRuleResponseToTerraformModel(ctx, rule)
		},
	})
}

// describeRuleMaxLength keeps built-in rule descriptions, which run to a paragraph, to one line
// of `terraform query` output.
const describeRuleMaxLength = 80

// describeRule is the display name for a rule. Terraform already prints the id next to it.
func describeRule(rule abacapi.RuleResponse) string {
	description := strings.Join(strings.Fields(rule.Description), " ")
	if description == "" {
		return fmt.Sprintf("runs on %d API execution point(s)", len(rule.APIExecutionPoints))
	}

	if runes := []rune(description); len(runes) > describeRuleMaxLength {
		return strings.TrimSpace(string(runes[:describeRuleMaxLength-1])) + "…"
	}

	return description
}

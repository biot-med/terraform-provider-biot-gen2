package rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
	"biot.com/terraform-provider-biot-gen2/internal/resources/abac"
)

var (
	_ resource.Resource                   = &BiotAbacRuleResource{}
	_ resource.ResourceWithConfigure      = &BiotAbacRuleResource{}
	_ resource.ResourceWithImportState    = &BiotAbacRuleResource{}
	_ resource.ResourceWithValidateConfig = &BiotAbacRuleResource{}
	_ resource.ResourceWithIdentity       = &BiotAbacRuleResource{}
	_ resource.ResourceWithModifyPlan     = &BiotAbacRuleResource{}
)

// Entity describes rules to the shared abac helpers. Rules have no implementation value, so
// there is no invalid-value code. Errors about the actions and conditions a rule references
// are handled separately, in errors.go.
var Entity = abac.Entity{
	Noun:                    "rule",
	ResourceType:            "biot_abac_rule",
	AlreadyExistsCode:       "RULE_ALREADY_EXISTS",
	NotFoundCode:            "RULE_NOT_FOUND",
	CannotDeleteBuiltInCode: "CANNOT_DELETE_BUILTIN_RULE",
}

func NewResource() resource.Resource {
	return &BiotAbacRuleResource{}
}

type BiotAbacRuleResource struct {
	client *api.APIClient
}

func (r *BiotAbacRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_abac_rule"
}

// The response of this function is passed to every rule resource when running
// create / read / update / delete. The input is the return of the ProviderConfigure function.
func (r *BiotAbacRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Every attribute inside the nested sets is Required. A Default inside a set element is
// unsafe - set elements are matched by value, so once the plan element differs from the
// configured one the default can overwrite a configured value (SOFT-9875) - and leaving them
// Optional would make an omitted value come back from the server as a mismatch (SOFT-9778).
func (r *BiotAbacRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An access-control (ABAC) rule: on the APIs it is attached to, runs " +
			"its actions when its conditions are met.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Unique id for the rule, chosen by you rather than generated " +
					"by BioT (for example `RESTRICT_PATIENT_SEARCH`). The API cannot change an " +
					"id, so changing it here destroys and recreates the rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Free-text description. Defaults to an empty string.",
			},
			"action_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				MarkdownDescription: "Ids of the actions the rule runs. At least one is required. " +
					"Refer to the resource (`biot_abac_action.<name>.id`) rather than typing the " +
					"id, so Terraform creates the action first, and updates this rule if the " +
					"action's id changes.",
			},
			"conditions": schema.SetNestedAttribute{
				Optional: true,
				Computed: true,
				// An empty default, rather than leaving it Optional, so that a rule without
				// conditions round-trips: the API cannot tell "omitted" from "empty".
				Default: setdefault.StaticValue(types.SetValueMust(conditionObjectType, []attr.Value{})),
				MarkdownDescription: "Conditions that decide whether the rule's actions run. " +
					"Each condition may be listed once. Defaults to none.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
							MarkdownDescription: "Id of the condition. Refer to the resource " +
								"(`biot_abac_condition.<name>.id`) rather than typing the id.",
						},
						"inverted": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether to negate the condition.",
						},
					},
				},
			},
			"api_execution_points": schema.SetNestedAttribute{
				Required: true,
				MarkdownDescription: "The APIs the rule runs on. A new rule needs at least one. A rule " +
					"that already exists in BioT with none - BioT allows clearing them on update - can " +
					"be imported and managed with `api_execution_points = []`, but not recreated that " +
					"way.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_id": schema.StringAttribute{
							Required: true,
							MarkdownDescription: "The HTTP method followed directly by the API's " +
								"path template, with no space - for example " +
								"`GET/organization/v1/users/patients/{id}`.",
						},
						"api_execution_point": schema.StringAttribute{
							Required: true,
							MarkdownDescription: "`PRE_REQUEST` to run before the API handles " +
								"the request, or `POST_REQUEST` to run on its response.",
						},
						"order": schema.Int64Attribute{
							Required: true,
							MarkdownDescription: "Order of this rule among the rules on the " +
								"same API and execution point. BioT's own rules use `1`.",
						},
						"enabled": schema.BoolAttribute{
							Required:            true,
							MarkdownDescription: "Whether the rule runs on this API.",
						},
					},
				},
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				MarkdownDescription: "Tags applied to the rule. Removing this from your " +
					"configuration keeps the tags BioT already has; set it to `[]` to clear " +
					"them.\n\nThe `<<BuiltIn>>` tag that BioT puts on its own rules is not " +
					"included here and cannot be set - use the `built_in` attribute instead.",
			},
			"built_in": schema.BoolAttribute{
				Computed: true,
				MarkdownDescription: "Whether BioT ships this rule. Built-in rules can be imported " +
					"and updated, but not destroyed.",
				PlanModifiers: []planmodifier.Bool{
					// Whether a rule is built-in never changes, so keep the known value rather
					// than replanning it as "(known after apply)" on every change.
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// IdentitySchema identifies a rule by its id. It is what `terraform query` returns for each
// rule it finds, and what an `import { identity = { id = "..." } }` block takes.
func (r *BiotAbacRuleResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The id of the rule.",
			},
		},
	}
}

// ValidateConfig turns what the service would reject into plan-time errors - see validate.go.
func (r *BiotAbacRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config TerraformAbacRule

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	abac.RejectBuiltInTag(config.Tags, Entity, &resp.Diagnostics)
	validateActionIDs(config.ActionIDs, &resp.Diagnostics)
	validateConditions(config.Conditions, &resp.Diagnostics)
	validateAPIExecutionPoints(config.APIExecutionPoints, &resp.Diagnostics)
}

// ModifyPlan enforces what only creating a rule requires - see validateNewRuleExecutionPoints.
//
// A null prior state covers every create, replacements included: when a rule is replaced -
// an id change, `-replace`, or `replace_triggered_by` - Terraform plans the new object a second
// time with no prior state, and that is the plan this check rejects. (resp.RequiresReplace is
// no help here: the framework always hands resource-level ModifyPlan an empty one.)
func (r *BiotAbacRuleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || !req.State.Raw.IsNull() {
		return // destroy, or an update of an existing rule
	}

	var points types.Set
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("api_execution_points"), &points)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateNewRuleExecutionPoints(points, &resp.Diagnostics)
}

func (r *BiotAbacRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TerraformAbacRule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest, diags := MapTerraformRuleToCreateRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Abac.CreateRule(ctx, createRequest)
	if err != nil {
		addRuleError(&resp.Diagnostics, "create", plan.ID.ValueString(), err)
		return
	}

	model, diags := MapRuleResponseToTerraformModel(ctx, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), model.ID)...)
}

func (r *BiotAbacRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TerraformAbacRule

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.Abac.GetRule(ctx, state.ID.ValueString())
	if err != nil {
		// A GET can only 404 on the rule itself, so not-found here is always drift.
		if transport.IsNotFound(err) {
			// Gone from the backend - drop it from state so Terraform plans a create.
			resp.State.RemoveResource(ctx)
			return
		}
		addRuleError(&resp.Diagnostics, "read", state.ID.ValueString(), err)
		return
	}

	model, diags := MapRuleResponseToTerraformModel(ctx, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), model.ID)...)
}

func (r *BiotAbacRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TerraformAbacRule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest, diags := MapTerraformRuleToUpdateRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unlike Read, a 404 here is not necessarily drift: it is also how the service reports a
	// reference to an action or condition that does not exist. addRuleError tells them apart.
	response, err := r.client.Abac.UpdateRule(ctx, plan.ID.ValueString(), updateRequest)
	if err != nil {
		addRuleError(&resp.Diagnostics, "update", plan.ID.ValueString(), err)
		return
	}

	model, diags := MapRuleResponseToTerraformModel(ctx, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, path.Root("id"), model.ID)...)
}

func (r *BiotAbacRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TerraformAbacRule

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Abac.DeleteRule(ctx, state.ID.ValueString())
	if err != nil {
		if transport.IsNotFound(err) {
			// Already gone - deleting it is exactly what we wanted.
			return
		}
		addRuleError(&resp.Diagnostics, "delete", state.ID.ValueString(), err)
	}
}

// Rules are imported by their id - either `terraform import <address> <id>`, or an import
// block with `id = "..."` or `identity = { id = "..." }`.
func (r *BiotAbacRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

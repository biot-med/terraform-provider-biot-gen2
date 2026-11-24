// Package portalbuilder holds the biot_portal_view resource, which applies a portal builder
// view's layout as raw JSON - typically one designed in the BioT UI and copied to other
// environments.
package portalbuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	portalbuilderapi "biot.com/terraform-provider-biot-gen2/internal/api/portalbuilder"
	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
	biotplanmodifiers "biot.com/terraform-provider-biot-gen2/internal/resources/biot_plan_modifiers"
)

var (
	_ resource.Resource                   = &BiotPortalViewResource{}
	_ resource.ResourceWithConfigure      = &BiotPortalViewResource{}
	_ resource.ResourceWithImportState    = &BiotPortalViewResource{}
	_ resource.ResourceWithValidateConfig = &BiotPortalViewResource{}
	_ resource.ResourceWithIdentity       = &BiotPortalViewResource{}
)

// ViewTypes are the view types BioT has, in the settings service's ViewTypeEnum.
var ViewTypes = []string{
	"ENTITY_LIST",
	"TEMPLATE_EXPAND",
	"TEMPLATE_PREVIEW",
	"TEMPLATE_LIST",
	"MAIN",
	"OPERATIONAL_ENTITY_LIST",
	"OPERATIONAL_TEMPLATE_PREVIEW",
}

func NewResource() resource.Resource {
	return &BiotPortalViewResource{}
}

type BiotPortalViewResource struct {
	client *api.APIClient
}

func (r *BiotPortalViewResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_portal_view"
}

func (r *BiotPortalViewResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BiotPortalViewResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Which view this resource manages. Changing any of these points the resource at another
	// view; nothing is deleted, since BioT owns the views.
	selector := func() []planmodifier.String {
		return []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "The layout of a portal builder view: which attributes, measurements " +
			"and tabs it shows.\n\n" +
			"BioT creates views by itself - one per portal and view type for each template or " +
			"entity - so this resource does not create one. It finds the existing view and " +
			"applies `config_json` to it. Destroying the resource only stops Terraform managing " +
			"the view; the view and its layout stay in BioT.\n\n" +
			"The intended workflow is to design the layout in the BioT portal builder, generate " +
			"this resource from it with `terraform query`, and apply it to other environments. " +
			"Manage the view's template with `biot_template` in the same configuration and refer " +
			"to it, since template ids differ between environments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The id BioT gave the view. It differs between environments.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"portal_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The portal the view belongs to: " + quoted(portalbuilderapi.PortalTypes) + ".",
				PlanModifiers:       selector(),
			},
			"view_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The type of the view: " + quoted(ViewTypes) + ".",
				PlanModifiers:       selector(),
			},
			"template_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The name of the template of a template view (`TEMPLATE_*` and " +
					"`OPERATIONAL_TEMPLATE_PREVIEW`), for example `patient`. Names are used rather " +
					"than ids because they are the same in every environment. The template must exist " +
					"before the view can be managed: if it is a `biot_template` in the same " +
					"configuration, add it to `depends_on`.",
				PlanModifiers: selector(),
			},
			"template_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The id of the view's template in this environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"entity_type_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The entity of an entity view (`ENTITY_LIST` and " +
					"`OPERATIONAL_ENTITY_LIST`), for example `patient`. Template views have one too, " +
					"but there `template_name` is enough to find the view.",
				PlanModifiers: selector(),
			},
			"config_json": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "The view's layout as JSON, in the body format of BioT's update " +
					"view API: an object with the sections `attributes`, `measurements` and `tabs`. " +
					"Generate it with `terraform query` rather than writing it by hand.\n\n" +
					"The whole layout is replaced on apply: a section left out is cleared. Template " +
					"tabs refer to their template by name - `template = { name = \"patient\" }` - and " +
					"the provider swaps in this environment's id when applying. Conditional values " +
					"hold their code snippet's code - `codeSnippet = \"user._template.name === 'nurse'\"` " +
					"- and the provider creates the snippet when applying.",
				PlanModifiers: []planmodifier.String{
					biotplanmodifiers.JsonNormalizePlanModifier{},
					keepEquivalentStateModifier{},
				},
			},
		},
	}
}

// IdentitySchema identifies a view by its portal and id - BioT's API needs both to read a view.
func (r *BiotPortalViewResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"portal_type": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The portal the view belongs to.",
			},
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
				Description:       "The id of the view.",
			},
		},
	}
}

func (r *BiotPortalViewResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config TerraformPortalView

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateOneOf(config.PortalType, path.Root("portal_type"), portalbuilderapi.PortalTypes, &resp.Diagnostics)
	validateOneOf(config.ViewType, path.Root("view_type"), ViewTypes, &resp.Diagnostics)

	if !config.ConfigJSON.IsNull() && !config.ConfigJSON.IsUnknown() {
		if err := validateConfigJSON(config.ConfigJSON.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("config_json"), "Invalid config_json", err.Error())
		}
	}
}

func (r *BiotPortalViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TerraformPortalView

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	templates := newResolver(r.client)

	view, diags := r.findView(ctx, plan, templates)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(view.ID)
	plan.TemplateID = optionalString(view.TemplateID())
	// Fill in whichever selectors were left unset; set ones already match the view.
	if !known(plan.TemplateName) {
		plan.TemplateName = optionalString(templateName(view))
	}
	if !known(plan.EntityTypeName) {
		plan.EntityTypeName = optionalString(view.EntityTypeName)
	}

	r.apply(ctx, plan, templates, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	setIdentity(ctx, plan, resp.Identity, &resp.Diagnostics)
}

func (r *BiotPortalViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TerraformPortalView

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	view, err := r.client.PortalBuilder.GetView(ctx, state.PortalType.ValueString(), state.ID.ValueString())
	if err != nil {
		if transport.IsNotFound(err) {
			// BioT deletes a view with its template, or when its entity stops supporting the
			// view type. Drop it so the next plan looks for the view again.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read portal view %s: %s", state.ID.ValueString(), err))
		return
	}

	model, diags := mapViewToTerraformModel(ctx, view, newResolver(r.client))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the configured form when BioT's is equivalent, so nulls and empty lists BioT adds
	// do not show as drift.
	if !state.ConfigJSON.IsNull() && equivalentJSON(state.ConfigJSON.ValueString(), model.ConfigJSON.ValueString()) {
		model.ConfigJSON = state.ConfigJSON
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
	setIdentity(ctx, model, resp.Identity, &resp.Diagnostics)
}

func (r *BiotPortalViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TerraformPortalView

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.apply(ctx, plan, newResolver(r.client), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	setIdentity(ctx, plan, resp.Identity, &resp.Diagnostics)
}

// Delete leaves the view in BioT, which has no API to delete one - Terraform just stops managing it.
func (r *BiotPortalViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// ImportState takes `<portal_type>/<view_id>`, or an identity block with both.
func (r *BiotPortalViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	portalType, viewID := "", ""

	if req.ID != "" {
		parts := strings.Split(req.ID, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			resp.Diagnostics.AddError("Invalid import id",
				fmt.Sprintf("Expected <portal_type>/<view_id>, for example ORGANIZATION_PORTAL/123e4567-e89b-12d3-a456-426614174000, got %q.", req.ID))
			return
		}
		portalType, viewID = parts[0], parts[1]
	} else {
		var identity struct {
			PortalType types.String `tfsdk:"portal_type"`
			ID         types.String `tfsdk:"id"`
		}
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		portalType, viewID = identity.PortalType.ValueString(), identity.ID.ValueString()
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("portal_type"), portalType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), viewID)...)
}

// findView looks up the view a new resource manages, by portal, view type and whichever of
// template_name and entity_type_name are set.
func (r *BiotPortalViewResource) findView(ctx context.Context, plan TerraformPortalView, templates *resolver) (portalbuilderapi.View, diag.Diagnostics) {
	var diags diag.Diagnostics

	views, err := r.client.PortalBuilder.ListViews(ctx, plan.PortalType.ValueString())
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to list the views of %s: %s", plan.PortalType.ValueString(), err))
		return portalbuilderapi.View{}, diags
	}
	templates.remember(views)

	var matches []portalbuilderapi.View
	for _, view := range views {
		if view.Type != plan.ViewType.ValueString() {
			continue
		}
		if known(plan.TemplateName) && templateName(view) != plan.TemplateName.ValueString() {
			continue
		}
		if known(plan.EntityTypeName) && view.EntityTypeName != plan.EntityTypeName.ValueString() {
			continue
		}
		matches = append(matches, view)
	}

	wanted := describeSelector(plan)
	switch len(matches) {
	case 1:
		return matches[0], diags
	case 0:
		if known(plan.TemplateName) {
			_, err := templates.templateID(ctx, plan.TemplateName.ValueString())
			if errors.Is(err, errTemplateNotFound) {
				diags.AddAttributeError(path.Root("template_name"), "Template not found",
					templateNotFoundError(plan.TemplateName.ValueString(), "template_name"))
				return portalbuilderapi.View{}, diags
			}
		}
		diags.AddError("Portal view not found",
			fmt.Sprintf("BioT has no %s. Check that the view type is supported for this template or entity - "+
				"the portal builder in the BioT UI shows which views exist.", wanted))
	default:
		var candidates []string
		for _, view := range matches {
			candidates = append(candidates, describeView(view))
		}
		diags.AddError("Portal view is ambiguous",
			fmt.Sprintf("%d views match %s. Set template_name or entity_type_name to pick one:\n  %s",
				len(matches), wanted, strings.Join(candidates, "\n  ")))
	}

	return portalbuilderapi.View{}, diags
}

// apply sends config_json to the view. BioT may change what it stores - for example, release
// flags make it ignore some fields - so a mismatch is reported as a warning: the next plan
// will show it as a difference.
func (r *BiotPortalViewResource) apply(ctx context.Context, plan TerraformPortalView, templates *resolver, diags *diag.Diagnostics) {
	// Learn the snippets the view uses now, so unchanged code keeps its snippet.
	current, err := r.client.PortalBuilder.GetView(ctx, plan.PortalType.ValueString(), plan.ID.ValueString())
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to read portal view %s: %s", plan.ID.ValueString(), err))
		return
	}
	if err := templates.rememberSnippets(ctx, current); err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to read the code snippets of portal view %s: %s", plan.ID.ValueString(), err))
		return
	}

	config, err := templates.toBioT(ctx, plan.ConfigJSON.ValueString())
	if err != nil {
		diags.AddAttributeError(path.Root("config_json"), "Invalid config_json", err.Error())
		return
	}

	view, err := r.client.PortalBuilder.UpdateView(ctx, plan.PortalType.ValueString(), plan.ID.ValueString(), json.RawMessage(config))
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to update portal view %s: %s", plan.ID.ValueString(), err))
		return
	}

	stored, err := configJSONFromView(view)
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to read back portal view %s: %s", plan.ID.ValueString(), err))
		return
	}

	if !equivalentJSON(config, stored) {
		diags.AddWarning("BioT changed the portal view's layout",
			fmt.Sprintf("BioT saved portal view %s differently from config_json, so the next plan will show a "+
				"difference. This usually means a feature behind a release flag is off in this environment.\n\n"+
				"BioT stored:\n%s", plan.ID.ValueString(), stored))
	}
}

func mapViewToTerraformModel(ctx context.Context, view portalbuilderapi.ViewResponse, templates *resolver) (TerraformPortalView, diag.Diagnostics) {
	var diags diag.Diagnostics

	config, err := configJSONFromView(view)
	if err == nil {
		config, err = templates.toPortable(ctx, config)
	}
	if err != nil {
		diags.AddError("API Error", fmt.Sprintf("Failed to encode the layout of portal view %s: %s", view.ID, err))
		return TerraformPortalView{}, diags
	}

	return TerraformPortalView{
		ID:             types.StringValue(view.ID),
		PortalType:     types.StringValue(view.PortalType),
		ViewType:       types.StringValue(view.Type),
		TemplateID:     optionalString(view.TemplateID()),
		TemplateName:   optionalString(templateName(view.View)),
		EntityTypeName: optionalString(view.EntityTypeName),
		ConfigJSON:     types.StringValue(config),
	}, diags
}

func setIdentity(ctx context.Context, model TerraformPortalView, identity *tfsdk.ResourceIdentity, diags *diag.Diagnostics) {
	diags.Append(identity.SetAttribute(ctx, path.Root("portal_type"), model.PortalType)...)
	diags.Append(identity.SetAttribute(ctx, path.Root("id"), model.ID)...)
}

func validateOneOf(value types.String, at path.Path, allowed []string, diags *diag.Diagnostics) {
	if !known(value) || slices.Contains(allowed, value.ValueString()) {
		return
	}
	diags.AddAttributeError(at, "Invalid value",
		fmt.Sprintf("%q is not one of %s.", value.ValueString(), quoted(allowed)))
}

func describeSelector(plan TerraformPortalView) string {
	description := fmt.Sprintf("%s view in %s", plan.ViewType.ValueString(), plan.PortalType.ValueString())
	if known(plan.TemplateName) {
		description += " for template " + plan.TemplateName.ValueString()
	}
	if known(plan.EntityTypeName) {
		description += " for entity " + plan.EntityTypeName.ValueString()
	}
	return description
}

func describeView(view portalbuilderapi.View) string {
	description := fmt.Sprintf("%s %s", view.PortalType, view.Type)
	if view.Template != nil {
		description += fmt.Sprintf(" - template %s", view.Template.Name)
	} else if view.EntityTypeName != "" {
		description += fmt.Sprintf(" - entity %s", view.EntityTypeName)
	}
	return description
}

func templateName(view portalbuilderapi.View) string {
	if view.Template == nil {
		return ""
	}
	return view.Template.Name
}

func known(value types.String) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func quoted(values []string) string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = "`" + value + "`"
	}
	return strings.Join(out, ", ")
}

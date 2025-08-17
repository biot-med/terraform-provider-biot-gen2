package template

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"biot.com/terraform-provider-biot/internal/api"
	biotplanmodifiers "biot.com/terraform-provider-biot/internal/resources/biot_plan_modifiers"
)

func NewResource() resource.Resource {
	return &BiotTemplateResource{}
}

type BiotTemplateResource struct {
	client *api.APIClient
}

func (r *BiotTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "biot_template"
}

// The response of this function is passed to every template resouce when running create / read / update / delete.
// The input in this function is the return of the ProviderConfigure function.
func (r *BiotTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BiotTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the template.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"owner_organization_id": schema.StringAttribute{
				Optional: true,
			},
			"analytics_db_configuration": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"entity_type": schema.StringAttribute{
				Optional: true,
			},
			"parent_template_id": schema.StringAttribute{
				Optional: true,
			},
			// TODO: Do we want ? is it helping the user in any way to have them ? (same todo in to_terraform_mapper and sdk_template_model)
			// "removable": schema.BoolAttribute{
			// 	Computed: true,
			// },
			// "creation_time": schema.StringAttribute{
			// 	Computed: true,
			// },
			// "last_modified_time": schema.StringAttribute{
			// 	Computed: true,
			// },
			"builtin_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Builtin attributes associated with the template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: attributeSchema(),
				},
				PlanModifiers: []planmodifier.Set{
					biotplanmodifiers.CopyIDFromStateByNameSetModifier{},
				},
			},
			"custom_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Custom attributes associated with the template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: attributeSchema(),
				},
				PlanModifiers: []planmodifier.Set{
					biotplanmodifiers.CopyIDFromStateByNameSetModifier{},
				},
			},
			"template_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Template attributes associated with the template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: templateAttributeSchema(),
				},
				PlanModifiers: []planmodifier.Set{
					biotplanmodifiers.CopyIDFromStateByNameSetModifier{},
				},
			},
		},
	}
}

func attributeSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			// PlanModifiers: []planmodifier.String {
			// 	biotplanmodifiers.CopyIDFromStateByNameStringModifier{},
			// },
		},
		"display_name": schema.StringAttribute{Optional: true},
		"phi":          schema.BoolAttribute{Optional: true},
		"name":         schema.StringAttribute{Required: true},
		"type":         schema.StringAttribute{Required: true},
		"category":     schema.StringAttribute{Optional: true},
		"base_path":    schema.StringAttribute{Optional: true},

		"reference_configuration": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"uniquely":                               schema.BoolAttribute{Optional: true},
				"referenced_side_attribute_name":         schema.StringAttribute{Optional: true},
				"referenced_side_attribute_display_name": schema.StringAttribute{Optional: true},
				"valid_templates_to_reference": schema.ListAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
				"entity_type": schema.StringAttribute{Optional: true},
			},
		},

		"link_configuration": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"entity_type_name": schema.StringAttribute{Optional: true},
				"template_id":      schema.StringAttribute{Optional: true},
				"attribute_id":     schema.StringAttribute{Optional: true},
			},
		},

		"validation": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"mandatory":     schema.BoolAttribute{Optional: true},
				"default_value": schema.StringAttribute{Optional: true},
				"min":           schema.Int64Attribute{Optional: true},
				"max":           schema.Int64Attribute{Optional: true},
				"regex":         schema.StringAttribute{Optional: true},
			},
		},

		"numeric_meta_data": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"units":       schema.StringAttribute{Optional: true},
				"upper_range": schema.Int64Attribute{Optional: true},
				"lower_range": schema.Int64Attribute{Optional: true},
				"sub_type":    schema.StringAttribute{Optional: true},
			},
		},

		"analytics_db_configuration": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{Optional: true},
			},
		},

		"selectable_values": schema.SetNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id":           schema.StringAttribute{Computed: true},
					"name":         schema.StringAttribute{Required: true},
					"display_name": schema.StringAttribute{Optional: true},
				},
			},
		},
	}
}

func templateAttributeSchema() map[string]schema.Attribute {
	base := attributeSchema()

	// VERY IMPORTANT: Clone the base attribute map to avoid mutating the original
	attrSchema := make(map[string]schema.Attribute, len(base)+2)
	maps.Copy(attrSchema, base)

	attrSchema["value_json"] = schema.StringAttribute{
		Optional:    true,
		Description: "Value as JSON string (used instead of DynamicAttribute due to Terraform limitations)",
		PlanModifiers: []planmodifier.String{
			biotplanmodifiers.JsonNormalizePlanModifier{},
		},
	}

	attrSchema["organization_selection"] = schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"allowed": schema.BoolAttribute{
				Optional: true,
			},
			"configuration": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"selected": schema.SetNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
					"all": schema.BoolAttribute{
						Optional: true,
					},
				},
			},
		},
	}
	return attrSchema
}

func (r *BiotTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TerraformTemplate

	// TODO: Implement destructive logic properly. (and probably not in "read" ...)
	// forceEnv := os.Getenv("TF_FORCE")
	// if forceEnv != "true" {
	// 	resp.Diagnostics.AddError(
	// 		"Force Required",
	// 		"This action is dangerous. Set TF_FORCE=true to proceed.",
	// 	)
	// 	return
	// }

	// Read current local state, throws error if failed to load state.
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.client
	getTemplateResponse, err := client.GetTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.SpecificErrorCodes.NotFound) {
			// The template is not exist in the backend, removing it from local state.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read template: %s", err))
		return
	}

	// Update state
	templateModel := mapTemplateResponseToTerrformModel(ctx, getTemplateResponse)
	diags = resp.State.Set(ctx, templateModel)
	resp.Diagnostics.Append(diags...)
}

func (r *BiotTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TerraformTemplate

	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := MapTerraformTemplateToCreateRequest(plan)
	response, err := r.client.CreateTemplate(ctx, createRequest)

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create template: %s", err))
		return
	}

	diags = resp.State.Set(ctx, mapTemplateResponseToTerrformModel(ctx, response))
	resp.Diagnostics.Append(diags...)
}

func (r *BiotTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TerraformTemplate
	var state TerraformTemplate

	req.Plan.Get(ctx, &plan)
	req.State.Get(ctx, &state)

	updateRequest := MapTerraformTemplateToUpdateRequest(plan)
	response, err := r.client.UpdateTemplate(ctx, state.ID.ValueString(), updateRequest)

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update template: %s", err))
		return
	}

	diags := resp.State.Set(ctx, mapTemplateResponseToTerrformModel(ctx, response))
	resp.Diagnostics.Append(diags...)
}

func (r *BiotTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TerraformTemplate

	req.State.Get(ctx, &state)

	client := r.client
	err := client.DeleteTemplate(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete template: %s", err))
	}
}

// Import state Works with entity-type:template-name (instead of ID)
func (r *BiotTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			`Expected format: "entity-type:template-name" (e.g., "caregiver:doctor")`,
		)
		return
	}

	entityType := idParts[0]
	templateName := idParts[1]
	tflog.Debug(ctx, fmt.Sprintf("Template importState: Going to import entity type: [%s], template name: [%s]", entityType, templateName))

	templateResponse, err := r.client.GetTemplateByTypeAndName(ctx, entityType, templateName)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to import template [entityType: %q and name: %q]", entityType, templateName),
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Template importState: Successfully got template with name: [%s], and id [%s]", templateResponse.Name, templateResponse.ID))

	tfModel := mapTemplateResponseToTerrformModel(ctx, templateResponse)

	diags := resp.State.Set(ctx, tfModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

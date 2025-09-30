package template

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	biotplanmodifiers "biot.com/terraform-provider-biot-gen2/internal/resources/biot_plan_modifiers"
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
			"builtin_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Builtin attributes associated with the template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: builtinAttributeSchema(),
				},
				PlanModifiers: []planmodifier.Set{
					biotplanmodifiers.CopyIDFromStateByNameSetModifier{},
				},
			},
			"custom_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Description: "Custom attributes associated with the template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: customAttributeSchema(),
				},
				PlanModifiers: []planmodifier.Set{
					biotplanmodifiers.CopyIDFromStateByNameSetModifier{},
				},
			},
			"template_attributes": schema.SetNestedAttribute{
				Optional:    true,
				Computed:    true,
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
			// Plan-modifier is implemeneted from the "parent" attribute.
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

func builtinAttributeSchema() map[string]schema.Attribute {
	base := attributeSchema()

	// VERY IMPORTANT: Clone the base attribute map to avoid mutating the original
	attrSchema := make(map[string]schema.Attribute, len(base)+2)
	maps.Copy(attrSchema, base)

	attrSchema["analytics_db_configuration"] = schema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Optional: true},
		},
	}

	return attrSchema
}

func customAttributeSchema() map[string]schema.Attribute {
	base := attributeSchema()

	// VERY IMPORTANT: Clone the base attribute map to avoid mutating the original
	attrSchema := make(map[string]schema.Attribute, len(base)+2)
	maps.Copy(attrSchema, base)

	attrSchema["analytics_db_configuration"] = schema.SingleNestedAttribute{
		Optional: true,
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Optional: true},
		},
	}

	return attrSchema
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
	}
	return attrSchema
}

func (r *BiotTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TerraformTemplate

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

	createRequest := MapTerraformTemplateToCreateRequest(ctx, plan)
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

	forceUpdateString := os.Getenv("TF_FORCE_UPDATE")
	var forceUpdate = false
	var err error

	if forceUpdateString != "" {
		forceUpdate, err = strconv.ParseBool(forceUpdateString)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid TF_FORCE_UPDATE value",
				fmt.Sprintf("Value [%q] is not a valid boolean (expected: true / false)", forceUpdateString),
			)
			return
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Invalid TF_FORCE_UPDATE value: [%q], should be boolean (true / false)", forceUpdateString)
		return
	}

	updateRequest := MapTerraformTemplateToUpdateRequest(ctx, plan)
	response, err := r.client.UpdateTemplate(ctx, state.ID.ValueString(), updateRequest, forceUpdate)

	if err != nil {
		if apiError, ok := api.ConvertAPIError(err); ok && apiError.Code == "CUSTOM_ATTRIBUTE_IN_USE" {
			formatCustomAttributeInUseError(apiError, resp)
		} else {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update template: %s", err))
		}
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

func formatCustomAttributeInUseError(apiError api.APIError, resp *resource.UpdateResponse) {
	// Extract attribute names from the details
	var attributeNames []string
	for _, attr := range apiError.Details.Attributes {
		attributeNames = append(attributeNames, attr.Name)
	}

	// Create the custom message
	var attributeList string
	if len(attributeNames) == 1 {
		attributeList = attributeNames[0]
	} else if len(attributeNames) == 2 {
		attributeList = fmt.Sprintf("%s and %s", attributeNames[0], attributeNames[1])
	} else {
		// Handle case with more than 2 attributes
		lastIdx := len(attributeNames) - 1
		attributeList = fmt.Sprintf("%s and %s",
			strings.Join(attributeNames[:lastIdx], ", "),
			attributeNames[lastIdx])
	}

	warningMessage := fmt.Sprintf(`You have made changes to the following observation attributes:
%s
If you choose to apply these changes ALL observation data will be deleted (including observations that were not changed). To apply the changes run:

TF_FORCE_UPDATE=true terraform apply`, attributeList)

	resp.Diagnostics.AddError("DESTRUCTIVE CHANGE WARNING", warningMessage)
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
	tflog.Debug(ctx, "Starting template import", map[string]interface{}{
		"entity_type":   entityType,
		"template_name": templateName,
	})

	templateResponse, err := r.client.GetTemplateByTypeAndName(ctx, entityType, templateName)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to import template [entityType: %q and name: %q]", entityType, templateName),
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Successfully retrieved template for import", map[string]interface{}{
		"template_name": templateResponse.Name,
		"template_id":   templateResponse.ID,
	})

	tfModel := mapTemplateResponseToTerrformModel(ctx, templateResponse)

	diags := resp.State.Set(ctx, tfModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

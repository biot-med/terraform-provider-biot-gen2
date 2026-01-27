package template

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TerraformTemplate struct {
	ID                       types.String                       `tfsdk:"id"`
	Name                     types.String                       `tfsdk:"name"`
	DisplayName              types.String                       `tfsdk:"display_name"`
	Description              types.String                       `tfsdk:"description"`
	OwnerOrganizationID      types.String                       `tfsdk:"owner_organization_id"`
	EntityTypeName           types.String                       `tfsdk:"entity_type"`
	AnalyticsDbConfiguration *TerraformAnalyticsDbConfiguration `tfsdk:"analytics_db_configuration"`
	ParentTemplateID         types.String                       `tfsdk:"parent_template_id"`
	BuiltInAttributes        []TerraformBuiltinAttribute               `tfsdk:"builtin_attributes"`
	CustomAttributes         []TerraformCustomAttribute               `tfsdk:"custom_attributes"`
	TemplateAttributes       []TerraformTemplateAttribute       `tfsdk:"template_attributes"`
}

type BaseTerraformAttribute struct {
	Name                     types.String                       `tfsdk:"name"`
	BasePath                 types.String                       `tfsdk:"base_path"`
	ID                       types.String                       `tfsdk:"id"`
	DisplayName              types.String                       `tfsdk:"display_name"`
	Phi                      types.Bool                         `tfsdk:"phi"`
	ReferenceConfiguration   *TerraformReferenceConfiguration   `tfsdk:"reference_configuration"`
	LinkConfiguration        *TerraformLinkConfiguration        `tfsdk:"link_configuration"`
	Validation               *TerraformValidation               `tfsdk:"validation"`
	NumericMetaData          *TerraformNumericMetaData          `tfsdk:"numeric_meta_data"`
	Type                     types.String                       `tfsdk:"type"`
	Category                 types.String                       `tfsdk:"category"`
	SelectableValues         []TerraformSelectableValue         `tfsdk:"selectable_values"`
}

type TerraformBuiltinAttribute struct {
	BaseTerraformAttribute

	AnalyticsDbConfiguration *TerraformAnalyticsDbConfiguration `tfsdk:"analytics_db_configuration"`
}

type TerraformCustomAttribute struct {
	BaseTerraformAttribute

	AnalyticsDbConfiguration *TerraformAnalyticsDbConfiguration `tfsdk:"analytics_db_configuration"`
}

type TerraformTemplateAttribute struct {
	BaseTerraformAttribute

	Value                 types.String                    `tfsdk:"value_json"`
	OrganizationSelection *TerraformOrganizationSelection `tfsdk:"organization_selection"`
}

type TerraformAnalyticsDbConfiguration struct {
	Name types.String `tfsdk:"name"`
}

type TerraformParentTemplate struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Name        types.String `tfsdk:"name"`
}

type TerraformReferenceConfiguration struct {
	Uniquely                           types.Bool     `tfsdk:"uniquely"`
	ReferencedSideAttributeName        types.String   `tfsdk:"referenced_side_attribute_name"`
	ReferencedSideAttributeDisplayName types.String   `tfsdk:"referenced_side_attribute_display_name"`
	ValidTemplatesToReference          []types.String `tfsdk:"valid_templates_to_reference"`
	EntityType                         types.String   `tfsdk:"entity_type"`
}

type TerraformLinkConfiguration struct {
	EntityTypeName types.String `tfsdk:"entity_type_name"`
	TemplateID     types.String `tfsdk:"template_id"`
	AttributeID    types.String `tfsdk:"attribute_id"`
}

type TerraformValidation struct {
	Mandatory    types.Bool   `tfsdk:"mandatory"`
	DefaultValue types.String `tfsdk:"default_value"`
	Min          types.Number  `tfsdk:"min"`
	Max          types.Number  `tfsdk:"max"`
	Regex        types.String `tfsdk:"regex"`
}

type TerraformValidationMetadata struct {
	MandatoryReadOnly types.Bool `tfsdk:"mandatory_read_only"`
	SystemMandatory   types.Bool `tfsdk:"system_mandatory"`
	PhiReadOnly       types.Bool `tfsdk:"phi_read_only"`
}

type TerraformNumericMetaData struct {
	Units      types.String `tfsdk:"units"`
	UpperRange types.Number  `tfsdk:"upper_range"`
	LowerRange types.Number  `tfsdk:"lower_range"`
	SubType    types.String `tfsdk:"sub_type"`
}

type TerraformSelectableValue struct {
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	ID          types.String `tfsdk:"id"`
}

type TerraformOrganizationSelection struct {
	Selected []TerraformIDWrapper `tfsdk:"selected"`
	All      types.Bool           ` tfsdk:"all"`
}
type TerraformIDWrapper struct {
	ID types.String `tfsdk:"id"`
}

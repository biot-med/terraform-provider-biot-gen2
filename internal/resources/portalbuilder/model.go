package portalbuilder

import "github.com/hashicorp/terraform-plugin-framework/types"

type TerraformPortalView struct {
	ID             types.String `tfsdk:"id"`
	PortalType     types.String `tfsdk:"portal_type"`
	ViewType       types.String `tfsdk:"view_type"`
	TemplateID     types.String `tfsdk:"template_id"`
	TemplateName   types.String `tfsdk:"template_name"`
	EntityTypeName types.String `tfsdk:"entity_type_name"`
	ConfigJSON     types.String `tfsdk:"config_json"`
}

package template

import (
	"context"

	"biot.com/terraform-provider-biot/internal/api"
	"biot.com/terraform-provider-biot/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func mapTemplateResponseToTerrformModel(ctx context.Context, resp api.TemplateResponse) TerraformTemplate {
	builtInAttrs := []TerraformBuiltinAttribute{}

	for _, attr := range resp.BuiltInAttributes {
		builtInAttrs = append(builtInAttrs, mapBuiltinAttributeResponseToTerraformAttribute(ctx, attr))
	}

	customAttrs := []TerraformCustomAttribute{}
	for _, attr := range resp.CustomAttributes {
		customAttrs = append(customAttrs, mapCustomAttributeResponseToTerraformAttribute(ctx, attr))
	}

	templateAttrs := []TerraformTemplateAttribute{}
	for _, attr := range resp.TemplateAttributes {
		templateAttrs = append(templateAttrs, mapTemplateAttributeResponseToTerrformAttribute(ctx, attr))
	}

	template := TerraformTemplate{
		ID:                  types.StringValue(resp.ID),
		Name:                types.StringValue(resp.Name),
		DisplayName:         types.StringValue(resp.DisplayName),
		Description:         utils.StringOrNullPtr(resp.Description),
		OwnerOrganizationID: utils.StringOrNullPtr(resp.OwnerOrganizationID),
		EntityTypeName:      types.StringValue(resp.EntityTypeName),

		AnalyticsDbConfiguration: mapToTerraformAnalyticsDbConfiguration(ctx, resp.AnalyticsDbConfiguration),

		ParentTemplateID: mapToTerraformParentTemplateID(ctx, resp.ParentTemplate),

		BuiltInAttributes:  builtInAttrs,
		CustomAttributes:   customAttrs,
		TemplateAttributes: templateAttrs,
	}

	return template
}

func mapBuiltinAttributeResponseToTerraformAttribute(ctx context.Context, attr api.BuiltinAttributeResponse) TerraformBuiltinAttribute {
	base := mapAttributeResponseToTerrformAttribute(ctx, attr.BaseAttributeResponse)
	
	return TerraformBuiltinAttribute{
		BaseTerraformAttribute: base,
		AnalyticsDbConfiguration: mapToTerraformAnalyticsDbConfiguration(ctx, attr.AnalyticsDbConfiguration),
	}
}

func mapCustomAttributeResponseToTerraformAttribute(ctx context.Context, attr api.CustomAttributeResponse) TerraformCustomAttribute {
	base := mapAttributeResponseToTerrformAttribute(ctx, attr.BaseAttributeResponse)
	
	return TerraformCustomAttribute{
		BaseTerraformAttribute: base,
		AnalyticsDbConfiguration: mapToTerraformAnalyticsDbConfiguration(ctx, attr.AnalyticsDbConfiguration),
	}
}

func mapAttributeResponseToTerrformAttribute(ctx context.Context, attr api.BaseAttributeResponse) BaseTerraformAttribute {
	return BaseTerraformAttribute{
		Name:                     types.StringValue(attr.Name),
		BasePath:                 utils.StringOrNullPtr(attr.BasePath),
		ID:                       types.StringValue(attr.ID),
		DisplayName:              types.StringValue(attr.DisplayName),
		Phi:                      types.BoolValue(attr.Phi),
		Type:                     types.StringValue(attr.Type),
		Category:                 mapToTerraformCategory(ctx, attr.Category),
		SelectableValues:         mapToTerraformSelectableValues(ctx, attr.Type, attr.SelectableValues),
		ReferenceConfiguration:   mapToTerraformReferenceConfiguration(ctx, attr.ReferenceConfiguration),
		LinkConfiguration:        mapToTerraformLinkConfiguration(ctx, attr.LinkConfiguration),
		Validation:               mapToTerraformValidation(ctx, attr.Validation),
		NumericMetaData:          mapToTerraformNumericMetaData(ctx, attr.NumericMetaData),
	}
}

func mapTemplateAttributeResponseToTerrformAttribute(ctx context.Context, attr api.TemplateAttributeResponse) TerraformTemplateAttribute {
	base := mapAttributeResponseToTerrformAttribute(ctx, attr.BaseAttributeResponse)

	return TerraformTemplateAttribute{
		BaseTerraformAttribute: base,
		Value:                	utils.InterfaceToJsonString(ctx, "value", attr.Value),
		OrganizationSelection: 	mapToTerraformOrganizationSelection(ctx, attr.OrganizationSelection),
	}
}

func mapToTerraformAnalyticsDbConfiguration(ctx context.Context, adbConfiguration *api.AnalyticsDbConfiguration) *TerraformAnalyticsDbConfiguration {
	if adbConfiguration == nil {
		return nil
	}

	return &TerraformAnalyticsDbConfiguration{
		Name: types.StringValue(adbConfiguration.Name),
	}
}

func mapToTerraformCategory(ctx context.Context, category *api.Category) basetypes.StringValue {
	if category == nil {
		return types.StringNull()
	}

	return types.StringValue(category.Name)
}

func mapToTerraformSelectableValues(ctx context.Context, attributeType string, selectableValues []api.SelectableValue) []TerraformSelectableValue {
	// It is important that we do NOT return nil here, otherwise terraform will detect changes where there are none.
	result := []TerraformSelectableValue{}

	// Timezone and Locale attributes are hard coded VERY LONG array we want to ignore them.
	if attributeType == "TIMEZONE" || attributeType == "LOCALE" {
		return []TerraformSelectableValue{}
	}

	for _, sv := range selectableValues {
		result = append(result, TerraformSelectableValue{
			Name:        types.StringValue(sv.Name),
			DisplayName: types.StringValue(sv.DisplayName),
			ID:          utils.StringOrNull(sv.ID),
		})
	}

	return result
}

func mapToTerraformReferenceConfiguration(ctx context.Context, referenceConfiguration *api.ReferenceConfiguration) *TerraformReferenceConfiguration {
	if referenceConfiguration == nil {
		return nil
	}

	return &TerraformReferenceConfiguration{
		Uniquely:                           types.BoolValue(referenceConfiguration.Uniquely),
		ReferencedSideAttributeName:        types.StringValue(referenceConfiguration.ReferencedSideAttributeName),
		ReferencedSideAttributeDisplayName: types.StringValue(referenceConfiguration.ReferencedSideAttributeDisplayName),
		ValidTemplatesToReference:          utils.ConvertStringList(referenceConfiguration.ValidTemplatesToReference),
		EntityType:                         types.StringValue(referenceConfiguration.EntityType),
	}
}

func mapToTerraformLinkConfiguration(ctx context.Context, linkConfiguration *api.LinkConfiguration) *TerraformLinkConfiguration {
	if linkConfiguration == nil {
		return nil
	}

	return &TerraformLinkConfiguration{
		TemplateID:     types.StringValue(linkConfiguration.TemplateID),
		EntityTypeName: types.StringValue(linkConfiguration.EntityTypeName),
		AttributeID:    types.StringValue(linkConfiguration.AttributeID),
	}
}

func mapToTerraformValidation(ctx context.Context, validation *api.Validation) *TerraformValidation {
	if validation == nil {
		return nil
	}

	return &TerraformValidation{
		Mandatory:    utils.BoolOrNullPtr(validation.Mandatory),
		DefaultValue: utils.StringOrNullPtr(validation.DefaultValue),
		Min:          utils.Int64OrNullPtr(validation.Min),
		Max:          utils.Int64OrNullPtr(validation.Max),
		Regex:        utils.StringOrNullPtr(validation.Regex),
	}
}

func mapToTerraformNumericMetaData(ctx context.Context, numericMetaData *api.NumericMetaData) *TerraformNumericMetaData {
	if numericMetaData == nil {
		return nil
	}

	upperRange := utils.Int64OrNullPtr(numericMetaData.UpperRange)
	lowerRange := utils.Int64OrNullPtr(numericMetaData.LowerRange)

	return &TerraformNumericMetaData{
		Units:      utils.StringOrNullPtr(numericMetaData.Units),
		UpperRange: upperRange,
		LowerRange: lowerRange,
		SubType:    utils.StringOrNullPtr(numericMetaData.SubType),
	}
}

func mapToTerraformParentTemplateID(ctx context.Context, parentTemplate *api.ParentTemplate) types.String {
	if parentTemplate == nil {
		return types.StringNull()
	}
	return types.StringValue(parentTemplate.ID)
}

func mapToTerraformOrganizationSelection(ctx context.Context, organizationSelection *api.OrganizationSelection) *TerraformOrganizationSelection {
	if organizationSelection == nil {
		return nil
	}

	return &TerraformOrganizationSelection{
		Allowed:       types.BoolValue(organizationSelection.Allowed),
		Configuration: mapToOrganizationSelectionConfiguration(ctx, organizationSelection.Configuration),
	}
}

func mapToOrganizationSelectionConfiguration(ctx context.Context, organizationSelectionConfiguration *api.OrganizationSelectionConfiguration) *TerraformOrganizationSelectionConfiguration {
	if organizationSelectionConfiguration == nil {
		return nil
	}

	return &TerraformOrganizationSelectionConfiguration{
		All:      types.BoolValue(organizationSelectionConfiguration.All),
		Selected: mapToTerraformIDWrappers(ctx, organizationSelectionConfiguration.Selected),
	}
}

func mapToTerraformIDWrappers(ctx context.Context, apiWrappers []api.IDWrapper) []TerraformIDWrapper {
	var result []TerraformIDWrapper

	for _, idWrapper := range apiWrappers {
		result = append(result, TerraformIDWrapper{
			ID: types.StringValue(idWrapper.ID),
		})
	}

	return result
}

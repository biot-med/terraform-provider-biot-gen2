package template

import (
	"context"
	"encoding/json"

	templateapi "biot.com/terraform-provider-biot-gen2/internal/api/template"
	"biot.com/terraform-provider-biot-gen2/internal/utils"
)

func MapTerraformTemplateToCreateRequest(ctx context.Context, t TerraformTemplate) templateapi.CreateTemplateRequest {
	return templateapi.CreateTemplateRequest{
		BaseTemplate: templateapi.BaseTemplate{
			DisplayName:              utils.StringOrEmpty(t.DisplayName),
			Name:                     utils.StringOrEmpty(t.Name),
			Description:              utils.StringOrNilPtr(t.Description),
			OwnerOrganizationID:      utils.StringOrNilPtr(t.OwnerOrganizationID),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(ctx, t.AnalyticsDbConfiguration),
		},
		EntityType:         utils.StringOrEmpty(t.EntityTypeName),
		ParentTemplateID:   utils.StringOrNilPtr(t.ParentTemplateID),
		BuiltInAttributes:  mapBuiltinAttributes(ctx, t.BuiltInAttributes),
		CustomAttributes:   mapCustomAttributes(ctx, t.CustomAttributes),
		TemplateAttributes: mapTemplateAttributes(ctx, t.TemplateAttributes),
	}
}

func MapTerraformTemplateToUpdateRequest(ctx context.Context, t TerraformTemplate) templateapi.UpdateTemplateRequest {
	return templateapi.UpdateTemplateRequest{
		BaseTemplate: templateapi.BaseTemplate{
			DisplayName:              utils.StringOrEmpty(t.DisplayName),
			Name:                     utils.StringOrEmpty(t.Name),
			Description:              utils.StringOrNilPtr(t.Description),
			OwnerOrganizationID:      utils.StringOrNilPtr(t.OwnerOrganizationID),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(ctx, t.AnalyticsDbConfiguration),
		},
		ParentTemplateID:   utils.StringOrNilPtr(t.ParentTemplateID),
		BuiltInAttributes:  mapBuiltinAttributes(ctx, t.BuiltInAttributes),
		CustomAttributes:   mapCustomAttributes(ctx, t.CustomAttributes),
		TemplateAttributes: mapTemplateAttributes(ctx, t.TemplateAttributes),
	}
}

func mapAnalyticsDbConfig(ctx context.Context, c *TerraformAnalyticsDbConfiguration) *templateapi.AnalyticsDbConfiguration {
	if c == nil || c.Name.IsNull() || c.Name.IsUnknown() {
		return nil
	}

	return &templateapi.AnalyticsDbConfiguration{
		Name: c.Name.ValueString(),
	}
}

func mapBaseAttribute(ctx context.Context, attr BaseTerraformAttribute) templateapi.BaseAttribute {
	return templateapi.BaseAttribute{
		Name:                   attr.Name.ValueString(),
		BasePath:               utils.StringOrNilPtr(attr.BasePath),
		ID:                     attr.ID.ValueString(),
		DisplayName:            attr.DisplayName.ValueString(),
		Phi:                    attr.Phi.ValueBool(),
		ReferenceConfiguration: mapReferenceConfiguration(attr.ReferenceConfiguration),
		LinkConfiguration:      mapLinkConfiguration(attr.LinkConfiguration),
		Validation:             mapValidation(attr.Validation),
		NumericMetaData:        mapNumericMetaData(attr.NumericMetaData),
		UiConfiguration:        mapUiConfiguration(attr.UiConfiguration),
		Type:                   attr.Type.ValueString(),
		SelectableValues:       mapSelectableValues(attr.Name.ValueString(), attr.SelectableValues),
	}
}

func mapCustomAttributes(ctx context.Context, attrs []TerraformCustomAttribute) []templateapi.CustomAttributeRequest {
	result := []templateapi.CustomAttributeRequest{}
	for _, attr := range attrs {
		result = append(result, templateapi.CustomAttributeRequest{
			BaseAttribute:            mapBaseAttribute(ctx, attr.BaseTerraformAttribute),
			Category:                 attr.Category.ValueString(),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(ctx, attr.AnalyticsDbConfiguration),
		})
	}
	return result
}

func mapBuiltinAttributes(ctx context.Context, attrs []TerraformBuiltinAttribute) []templateapi.BuiltinAttributeRequest {
	result := []templateapi.BuiltinAttributeRequest{}
	for _, attr := range attrs {
		result = append(result, templateapi.BuiltinAttributeRequest{
			BaseAttribute:            mapBaseAttribute(ctx, attr.BaseTerraformAttribute),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(ctx, attr.AnalyticsDbConfiguration),
		})
	}
	return result
}

func mapTemplateAttributes(ctx context.Context, attrs []TerraformTemplateAttribute) []templateapi.TemplateAttributeRequest {
	result := make([]templateapi.TemplateAttributeRequest, 0, len(attrs))

	for _, attr := range attrs {
		var value interface{}

		if !attr.Value.IsNull() && !attr.Value.IsUnknown() {
			jsonStr := attr.Value.ValueString()
			var decoded map[string]interface{}
			json.Unmarshal([]byte(jsonStr), &decoded)
			if v, ok := decoded["value"]; ok {
				value = v
			} else {
				// "value" key missing, fallback to nil or whole map
				value = nil
			}
		} else {
			value = nil
		}

		result = append(result, templateapi.TemplateAttributeRequest{
			BaseAttribute:                      mapBaseAttribute(ctx, attr.BaseTerraformAttribute),
			Value:                              value,
			OrganizationSelectionConfiguration: mapOrgSelection(attr.OrganizationSelection),
		})
	}

	return result
}

func mapReferenceConfiguration(rc *TerraformReferenceConfiguration) *templateapi.ReferenceConfiguration {
	if rc == nil {
		return nil
	}

	return &templateapi.ReferenceConfiguration{
		Uniquely:                           rc.Uniquely.ValueBool(),
		ReferencedSideAttributeName:        rc.ReferencedSideAttributeName.ValueString(),
		ReferencedSideAttributeDisplayName: rc.ReferencedSideAttributeDisplayName.ValueString(),
		ValidTemplatesToReference:          utils.ConvertTerraformStringList(rc.ValidTemplatesToReference),
		EntityType:                         rc.EntityType.ValueString(),
	}
}

func mapLinkConfiguration(lc *TerraformLinkConfiguration) *templateapi.LinkConfiguration {
	if lc == nil {
		return nil
	}
	return &templateapi.LinkConfiguration{
		EntityTypeName: lc.EntityTypeName.ValueString(),
		TemplateID:     lc.TemplateID.ValueString(),
		AttributeID:    lc.AttributeID.ValueString(),
	}
}

func mapValidation(v *TerraformValidation) *templateapi.Validation {
	if v == nil {
		return nil
	}

	validation := &templateapi.Validation{
		Mandatory: utils.BoolOrNilPtr(v.Mandatory),
		Unique:    utils.BoolOrNilPtr(v.Unique),
	}

	if !v.DefaultValue.IsNull() && !v.DefaultValue.IsUnknown() {
		validation.DefaultValue = utils.StringOrNilPtr(v.DefaultValue)
	}

	if !v.Min.IsNull() && !v.Min.IsUnknown() {
		validation.Min = utils.Float64OrNilPtr(v.Min)
	}

	if !v.Max.IsNull() && !v.Max.IsUnknown() {
		validation.Max = utils.Float64OrNilPtr(v.Max)
	}

	if !v.Regex.IsNull() && !v.Regex.IsUnknown() {
		validation.Regex = utils.StringOrNilPtr(v.Regex)
	}

	if len(v.SupportedMimeTypes) > 0 {
		validation.SupportedMimeTypes = utils.ConvertTerraformStringList(v.SupportedMimeTypes)
	}

	return validation
}

func mapNumericMetaData(numericMetaData *TerraformNumericMetaData) *templateapi.NumericMetaData {
	if numericMetaData == nil {
		return nil
	}

	return &templateapi.NumericMetaData{
		Units:      utils.StringOrNilPtr(numericMetaData.Units),
		UpperRange: utils.Float64OrNilPtr(numericMetaData.UpperRange),
		LowerRange: utils.Float64OrNilPtr(numericMetaData.LowerRange),
		SubType:    utils.StringOrNilPtr(numericMetaData.SubType),
	}
}

func mapUiConfiguration(uiConfiguration *TerraformUiConfiguration) *templateapi.UiConfiguration {
	if uiConfiguration == nil {
		return nil
	}

	result := &templateapi.UiConfiguration{}

	if uiConfiguration.Date != nil {
		result.Date = &templateapi.DateUiConfiguration{
			DateStyle: utils.StringOrNilPtr(uiConfiguration.Date.DateStyle),
		}
	}

	if uiConfiguration.DateTime != nil {
		result.DateTime = &templateapi.DateTimeUiConfiguration{
			DateStyle: utils.StringOrNilPtr(uiConfiguration.DateTime.DateStyle),
			TimeStyle: utils.StringOrNilPtr(uiConfiguration.DateTime.TimeStyle),
		}
	}

	return result
}

func mapSelectableValues(attributeType string, vals []TerraformSelectableValue) []templateapi.SelectableValue {
	result := []templateapi.SelectableValue{}
	if attributeType == "TIMEZONE" || attributeType == "LOCALE" {
		return []templateapi.SelectableValue{}
	}

	for _, val := range vals {
		result = append(result, templateapi.SelectableValue{
			Name:        val.Name.ValueString(),
			DisplayName: val.DisplayName.ValueString(),
			ID:          utils.StringOrEmpty(val.ID),
		})
	}
	return result
}

func mapOrgSelection(organizationSelection *TerraformOrganizationSelection) *templateapi.OrganizationSelectionConfiguration {
	if organizationSelection == nil {
		return nil
	}

	selected := make([]templateapi.IDWrapper, len(organizationSelection.Selected))
	for i, s := range organizationSelection.Selected {
		selected[i] = templateapi.IDWrapper{ID: s.ID.ValueString()}
	}

	return &templateapi.OrganizationSelectionConfiguration{
		Selected: selected,
		All:      organizationSelection.All.ValueBool(),
	}
}

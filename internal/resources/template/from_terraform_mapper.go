package template

import (
	"encoding/json"

	"biot.com/terraform-provider-biot/internal/api"
	"biot.com/terraform-provider-biot/internal/utils"
)

func MapTerraformTemplateToCreateRequest(t TerraformTemplate) api.CreateTemplateRequest {
	return api.CreateTemplateRequest{
		BaseTemplate: api.BaseTemplate{
			DisplayName:              utils.StringOrEmpty(t.DisplayName),
			Name:                     utils.StringOrEmpty(t.Name),
			Description:              utils.StringOrNilPtr(t.Description),
			OwnerOrganizationID:      utils.StringOrNilPtr(t.OwnerOrganizationID),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(t.AnalyticsDbConfiguration),
		},
		EntityType:         utils.StringOrEmpty(t.EntityTypeName),
		ParentTemplateID:   utils.StringOrNilPtr(t.ParentTemplateID),
		BuiltInAttributes:  mapBaseAttributes(t.BuiltInAttributes),
		CustomAttributes:   mapCustomAttributes(t.CustomAttributes),
		TemplateAttributes: mapTemplateAttributes(t.TemplateAttributes),
	}
}

func MapTerraformTemplateToUpdateRequest(t TerraformTemplate) api.UpdateTemplateRequest {
	return api.UpdateTemplateRequest{
		BaseTemplate: api.BaseTemplate{
			DisplayName:              utils.StringOrEmpty(t.DisplayName),
			Name:                     utils.StringOrEmpty(t.Name),
			Description:              utils.StringOrNilPtr(t.Description),
			OwnerOrganizationID:      utils.StringOrNilPtr(t.OwnerOrganizationID),
			AnalyticsDbConfiguration: mapAnalyticsDbConfig(t.AnalyticsDbConfiguration),
		},
		ParentTemplateID:   utils.StringOrNilPtr(t.ParentTemplateID),
		BuiltInAttributes:  mapBaseAttributes(t.BuiltInAttributes),
		CustomAttributes:   mapCustomAttributes(t.CustomAttributes),
		TemplateAttributes: mapTemplateAttributes(t.TemplateAttributes),
	}
}

func mapAnalyticsDbConfig(c *TerraformAnalyticsDbConfiguration) *api.AnalyticsDbConfiguration {
	if c == nil || c.Name.IsNull() || c.Name.IsUnknown() {
		return nil
	}

	return &api.AnalyticsDbConfiguration{
		Name: c.Name.ValueString(),
	}
}

func mapBaseAttributes(attrs []TerraformAttribute) []api.BaseAttribute {
	result := []api.BaseAttribute{}
	for _, attr := range attrs {
		result = append(result, mapBaseAttribute(attr))
	}
	return result
}

func mapBaseAttribute(attr TerraformAttribute) api.BaseAttribute {
	return api.BaseAttribute{
		Name:                     attr.Name.ValueString(),
		// BasePath:                 stringOrNilPtr(attr.BasePath),
		ID:                       attr.ID.ValueString(),
		DisplayName:              attr.DisplayName.ValueString(),
		Phi:                      attr.Phi.ValueBool(),
		ReferenceConfiguration:   mapReferenceConfiguration(attr.ReferenceConfiguration),
		LinkConfiguration:        mapLinkConfiguration(attr.LinkConfiguration),
		Validation:               mapValidation(attr.Validation),
		NumericMetaData:          mapNumericMetaData(attr.NumericMetaData),
		Type:                     attr.Type.ValueString(),
		SelectableValues:         mapSelectableValues(attr.Name.ValueString(), attr.SelectableValues),
		AnalyticsDbConfiguration: mapAnalyticsDbConfig(attr.AnalyticsDbConfiguration),
	}
}

func mapCustomAttributes(attrs []TerraformAttribute) []api.CustomAttributeRequest {
	result := []api.CustomAttributeRequest{}
	for _, attr := range attrs {
		result = append(result, api.CustomAttributeRequest{
			BaseAttribute: mapBaseAttribute(attr),
			Category:      attr.Category.ValueString(), // extract Go string
		})
	}
	return result
}

func mapTemplateAttributes(attrs []TerraformTemplateAttribute) []api.CreateTemplateAttribute {
    result := make([]api.CreateTemplateAttribute, 0, len(attrs))

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

        result = append(result, api.CreateTemplateAttribute{
            BaseAttribute:         mapBaseAttribute(attr.TerraformAttribute),
            Value:                 value,
            OrganizationSelection: mapOrgSelection(attr.OrganizationSelection),
        })
    }

    return result
}

func mapReferenceConfiguration(rc *TerraformReferenceConfiguration) *api.ReferenceConfiguration {
    if rc == nil {
        return nil
    }

    return &api.ReferenceConfiguration{
        Uniquely:                           rc.Uniquely.ValueBool(),
        ReferencedSideAttributeName:        rc.ReferencedSideAttributeName.ValueString(),
        ReferencedSideAttributeDisplayName: rc.ReferencedSideAttributeDisplayName.ValueString(),
        ValidTemplatesToReference:          utils.ConvertTerraformStringList(rc.ValidTemplatesToReference),
        EntityType:                         rc.EntityType.ValueString(),
    }
}

func mapLinkConfiguration(lc *TerraformLinkConfiguration) *api.LinkConfiguration {
	if lc == nil {
		return nil
	}
	return &api.LinkConfiguration{
		EntityTypeName: lc.EntityTypeName.ValueString(),
		TemplateID:     lc.TemplateID.ValueString(),
		AttributeID:    lc.AttributeID.ValueString(),
	}
}

func mapValidation(v *TerraformValidation) *api.Validation {
	if v == nil {
		return nil
	}

	validation := &api.Validation{
		Mandatory: utils.BoolOrNilPtr(v.Mandatory),
	}

	if !v.DefaultValue.IsNull() && !v.DefaultValue.IsUnknown() {
		validation.DefaultValue = utils.StringOrNilPtr(v.DefaultValue)
	}

	if !v.Min.IsNull() && !v.Min.IsUnknown() {
		validation.Min = utils.Int64OrNilPtr(v.Min)
	}

	if !v.Max.IsNull() && !v.Max.IsUnknown() {
		validation.Max = utils.Int64OrNilPtr(v.Max)
	}

	if !v.Regex.IsNull() && !v.Regex.IsUnknown() {
		validation.Regex = utils.StringOrNilPtr(v.Regex)
	}

	return validation
}

func mapNumericMetaData(nm *TerraformNumericMetaData) *api.NumericMetaData {
	if nm == nil {
		return nil
	}

	var upperRange *int64
	if !nm.UpperRange.IsNull() && !nm.UpperRange.IsUnknown() {
		val := nm.UpperRange.ValueInt64()
		upperRange = &val
	}

	var lowerRange *int64
	if !nm.LowerRange.IsNull() && !nm.LowerRange.IsUnknown() {
		val := nm.LowerRange.ValueInt64()
		lowerRange = &val
	}

	return &api.NumericMetaData{
		Units:      utils.StringOrEmpty(nm.Units),
		UpperRange: upperRange,
		LowerRange: lowerRange,
		SubType:    utils.StringOrEmpty(nm.SubType),
	}
}

func mapSelectableValues(attributeType string, vals []TerraformSelectableValue) []api.SelectableValue {
	result := []api.SelectableValue{}
	if attributeType == "TIMEZONE" || attributeType == "LOCALE" {
		return []api.SelectableValue{}
	}

	for _, val := range vals {
		result = append(result, api.SelectableValue{
			Name:        val.Name.ValueString(),
			DisplayName: val.DisplayName.ValueString(),
			ID:          utils.StringOrEmpty(val.ID),
		})
	}
	return result
}

func mapOrgSelection(sel *TerraformOrganizationSelection) *api.OrganizationSelection {
	if sel == nil {
		return nil
	}
	return &api.OrganizationSelection{
		Allowed:       sel.Allowed.ValueBool(),
		Configuration: mapOrgSelectionConfig(sel.Configuration),
	}
}

func mapOrgSelectionConfig(cfg *TerraformOrganizationSelectionConfiguration) *api.OrganizationSelectionConfiguration {
	if cfg == nil {
		return nil
	}
	selected := make([]api.IDWrapper, len(cfg.Selected))
	for i, s := range cfg.Selected {
		selected[i] = api.IDWrapper{ID: s.ID.ValueString()}
	}
	return &api.OrganizationSelectionConfiguration{
		Selected: selected,
		All:      cfg.All.ValueBool(),
	}
}


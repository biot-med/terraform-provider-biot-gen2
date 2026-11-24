package portalbuilder

import "encoding/json"

type ListViewsResponse struct {
	Views []View `json:"views"`
}

// View identifies a view. A template view has both Template and EntityTypeName (the template's
// entity), an entity view has only EntityTypeName, and the MAIN view has neither.
type View struct {
	ID             string           `json:"id"`
	Type           string           `json:"type"`
	PortalType     string           `json:"portalType"`
	EntityTypeName string           `json:"entityTypeName,omitempty"`
	Template       *TemplateDetails `json:"template,omitempty"`
}

func (v View) TemplateID() string {
	if v.Template == nil {
		return ""
	}

	return v.Template.ID
}

type TemplateDetails struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// ViewResponse is a view together with its configuration. Every section mixes what is selected
// with what could be selected ("available"); only the selected part is modelled, and kept raw,
// since the provider passes it through without interpreting it.
type ViewResponse struct {
	View

	Attributes   []AttributesGroup     `json:"attributes,omitempty"`
	Measurements *MeasurementsResponse `json:"measurements,omitempty"`
	Tabs         *TabsResponse         `json:"tabs,omitempty"`
}

type AttributesGroup struct {
	Classification *string           `json:"classification,omitempty"`
	Selected       []json.RawMessage `json:"selected"`
}

type MeasurementsResponse struct {
	Selected json.RawMessage `json:"selected,omitempty"`
}

type TabsResponse struct {
	Selected []json.RawMessage `json:"selected,omitempty"`
}

// ViewConfig is the body of an update: the selected items of each section. It is what the
// provider stores as config_json, so a view read back from BioT can be applied elsewhere as is.
type ViewConfig struct {
	Attributes   []AttributesGroup `json:"attributes,omitempty"`
	Measurements json.RawMessage   `json:"measurements,omitempty"`
	Tabs         []json.RawMessage `json:"tabs,omitempty"`
}

// ConfigKeys are the top-level keys an update accepts.
var ConfigKeys = []string{"attributes", "measurements", "tabs"}

// SelectedConfig extracts the selected items of a view, in the shape an update takes. Attribute
// groups with nothing selected are dropped: BioT lists every group that has attributes available,
// selected or not, and including empty ones would only add noise.
func (r ViewResponse) SelectedConfig() ViewConfig {
	var config ViewConfig

	for _, group := range r.Attributes {
		if len(group.Selected) > 0 {
			config.Attributes = append(config.Attributes, group)
		}
	}

	if r.Measurements != nil && len(r.Measurements.Selected) > 0 && string(r.Measurements.Selected) != "null" {
		config.Measurements = r.Measurements.Selected
	}

	if r.Tabs != nil {
		config.Tabs = r.Tabs.Selected
	}

	return config
}

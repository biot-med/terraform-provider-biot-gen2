package api

type BaseTemplate struct {
	DisplayName              string                     `json:"displayName"`
	Name                     string                     `json:"name"`
	Description              *string                    `json:"description"`
	OwnerOrganizationID      *string                    `json:"ownerOrganizationId"`
	AnalyticsDbConfiguration *AnalyticsDbConfiguration  `json:"analyticsDbConfiguration"`
}

type UpdateTemplateRequest struct {
	BaseTemplate
	ParentTemplateID   *string                   `json:"parentTemplateId"`
	BuiltInAttributes  []BaseAttribute           `json:"builtInAttributes"`
	CustomAttributes   []CustomAttributeRequest  `json:"customAttributes"`
	TemplateAttributes []CreateTemplateAttribute `json:"templateAttributes"`
}

type CreateTemplateRequest struct {
	BaseTemplate
	EntityType     	   string                    `json:"entityType"`
	ParentTemplateID   *string                   `json:"parentTemplateId"`
	BuiltInAttributes  []BaseAttribute           `json:"builtInAttributes"`
	CustomAttributes   []CustomAttributeRequest  `json:"customAttributes"`
	TemplateAttributes []CreateTemplateAttribute `json:"templateAttributes"`
}

type TemplateResponse struct {
	BaseTemplate
	ID                 string                      `json:"id"`
	// TODO: Do we want ? (same todo in to_terraform_mapper and resource schema and TerraformTemplate)
	// Removable          bool                        `json:"removable"`
	// CreationTime       string                      `json:"creationTime"`
	// LastModifiedTime   string                      `json:"lastModifiedTime"`
	EntityTypeName     string                      `json:"entityTypeName"`
	ParentTemplate     *ParentTemplate             `json:"parentTemplate"`
	BuiltInAttributes  []BaseAttributeResponse     `json:"builtInAttributes"`
	CustomAttributes   []BaseAttributeResponse     `json:"customAttributes"`
	TemplateAttributes []TemplateAttributeResponse `json:"templateAttributes"`
}

// ********* All below are for search: *****************
type SearchTemplatesResponse struct {
	Data     []TemplateResponse `json:"data"`
	Metadata SearchMetadata     `json:"metadata"`
}

type SearchMetadata struct {
	Sort           []string               `json:"sort"`
	Filter         map[string]FilterEntry `json:"filter"`
	Page           PageMetadata           `json:"page"`
	FreeTextSearch *string                `json:"freeTextSearch"` // nullable
}

type FilterEntry struct {
	In     []string               `json:"in"`
	NotIn  []string               `json:"notIn"`
	Filter map[string]interface{} `json:"filter"` // generic inner filter
}

type PageMetadata struct {
	TotalResults int `json:"totalResults"`
	Page         int `json:"page"`
	Limit        int `json:"limit"`
}

// ***************************************************

/* Template DTO Models */

type BaseAttribute struct {
	Name                     string                    `json:"name"`
	// BasePath                 *string                   `json:"basePath"`
	ID                       string                    `json:"id"`
	DisplayName              string                    `json:"displayName"`
	Phi                      bool                      `json:"phi"`
	ReferenceConfiguration   *ReferenceConfiguration   `json:"referenceConfiguration"`
	LinkConfiguration        *LinkConfiguration        `json:"linkConfiguration"`
	Validation               *Validation               `json:"validation"`
	NumericMetaData          *NumericMetaData          `json:"numericMetaData"`
	Type                     string                    `json:"type"`
	SelectableValues         []SelectableValue         `json:"selectableValues"`
	AnalyticsDbConfiguration *AnalyticsDbConfiguration `json:"analyticsDbConfiguration"`
	ValidationMetadata       *ValidationMetadata       `json:"validationMetadata,omitempty"`
	ReadOnly                 bool                      `json:"readOnly,omitempty"`
}

type CustomAttributeRequest struct {
	BaseAttribute
	Category string `json:"category"`
}

type BaseAttributeResponse struct {
	BaseAttribute
	Type                     string                    `json:"type"`
	Name                     string                    `json:"name"`
	// BasePath                 *string                   `json:"basePath"`
	Category *Category `json:"category"`
}

type CreateTemplateAttribute struct {
	BaseAttribute
	Value                 interface{}             `json:"value"`
	OrganizationSelection *OrganizationSelection `json:"organizationSelection,omitempty"`
}

type TemplateAttributeResponse struct {
	BaseAttributeResponse
	Value                 interface{}             `json:"value"`
	OrganizationSelection *OrganizationSelection `json:"organizationSelection,omitempty"`
}

type AnalyticsDbConfiguration struct {
	Name string `json:"name"`
}

type ParentTemplate struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
}

type ReferenceConfiguration struct {
	Uniquely                           bool     `json:"uniquely"`
	ReferencedSideAttributeName        string   `json:"referencedSideAttributeName"`
	ReferencedSideAttributeDisplayName string   `json:"referencedSideAttributeDisplayName"`
	ValidTemplatesToReference          []string `json:"validTemplatesToReference"`
	EntityType                         string   `json:"entityType"`
}

type LinkConfiguration struct {
	EntityTypeName string `json:"entityTypeName"`
	TemplateID     string `json:"templateId"`
	AttributeID    string `json:"attributeId"`
}

type Validation struct {
	Mandatory    *bool   `json:"mandatory"`
	DefaultValue *string `json:"defaultValue"`
	Min          *int64  `json:"min"`
	Max          *int64  `json:"max"`
	Regex        *string `json:"regex"`
}

type ValidationMetadata struct {
	MandatoryReadOnly bool `json:"mandatoryReadOnly"`
	SystemMandatory   bool `json:"systemMandatory"`
	PhiReadOnly       bool `json:"phiReadOnly"`
}

type NumericMetaData struct {
    Units      string
    UpperRange *int64
    LowerRange *int64
    SubType    string
}

type Category struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type SelectableValue struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ID          string `json:"id,omitempty"`
}

type OrganizationSelection struct {
	Allowed       bool                                `json:"allowed"`
	Configuration *OrganizationSelectionConfiguration `json:"configuration"`
}

type OrganizationSelectionConfiguration struct {
	Selected []IDWrapper `json:"selected"`
	All      bool        `json:"all"`
}

type IDWrapper struct {
	ID string `json:"id""`
}

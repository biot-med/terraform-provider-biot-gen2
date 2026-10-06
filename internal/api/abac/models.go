package abac

// Request / response models for the BioT access-control (ABAC) service.
//
// Update requests intentionally omit `id` and `value`: the service maps both with
// `ignore = true` on PATCH, so neither can be changed after creation. The resources
// mark them as RequiresReplace rather than silently no-op'ing a change.
//
// None of the update fields carry `omitempty`. The service treats an absent field and an
// explicit null identically ("leave unchanged"), so omitting an empty value would make it
// impossible to clear a description, empty a tag set, or reset params. Callers must pass
// non-nil maps and slices for the same reason - a nil one marshals to null and is ignored.

type CreateConditionRequest struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

type UpdateConditionRequest struct {
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

type ConditionResponse struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

type CreateActionRequest struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

type UpdateActionRequest struct {
	Description string                 `json:"description"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

type ActionResponse struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Value       string                 `json:"value"`
	Params      map[string]interface{} `json:"params"`
	Tags        []string               `json:"tags"`
}

// Rules reference actions and conditions by id. The service also wraps each action id in
// an object of its own, which RuleActionRef mirrors.
type RuleActionRef struct {
	ID string `json:"id"`
}

type RuleConditionRef struct {
	ID       string `json:"id"`
	Inverted bool   `json:"inverted"`
}

// RuleAPIExecutionPoint binds a rule to one API. APIID is the HTTP method followed directly by
// the path template, e.g. "GET/organization/v1/users/patients/{id}".
//
// The response form carries an extra derived `tags` field per execution point; it is copied
// from the rule's own tags and is not settable, so it is deliberately not modelled here.
type RuleAPIExecutionPoint struct {
	APIID          string `json:"apiId"`
	ExecutionPoint string `json:"apiExecutionPoint"`
	Order          int64  `json:"order"`
	Enabled        bool   `json:"enabled"`
}

type CreateRuleRequest struct {
	ID                 string                  `json:"id"`
	Description        string                  `json:"description"`
	Actions            []RuleActionRef         `json:"actions"`
	Conditions         []RuleConditionRef      `json:"conditions"`
	APIExecutionPoints []RuleAPIExecutionPoint `json:"apiExecutionPoints"`
	Tags               []string                `json:"tags"`
}

// UpdateRuleRequest replaces actions, conditions and execution points wholesale whenever they
// are present, which they always are - see the note at the top of this file.
type UpdateRuleRequest struct {
	Description        string                  `json:"description"`
	Actions            []RuleActionRef         `json:"actions"`
	Conditions         []RuleConditionRef      `json:"conditions"`
	APIExecutionPoints []RuleAPIExecutionPoint `json:"apiExecutionPoints"`
	Tags               []string                `json:"tags"`
}

type RuleResponse struct {
	ID                 string                  `json:"id"`
	Description        string                  `json:"description"`
	Actions            []RuleActionRef         `json:"actions"`
	Conditions         []RuleConditionRef      `json:"conditions"`
	APIExecutionPoints []RuleAPIExecutionPoint `json:"apiExecutionPoints"`
	Tags               []string                `json:"tags"`
}

// ErrorDetails is the access-control shape of transport.APIError.Details. Recover it with
// apiError.DecodeDetails(&details).
type ErrorDetails struct {
	ID           string   `json:"id"`
	Value        string   `json:"value"`
	ActionIDs    []string `json:"actionIds"`
	ConditionIDs []string `json:"conditionIds"`
}

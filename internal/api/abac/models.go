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

// ErrorDetails is the access-control shape of transport.APIError.Details. Recover it with
// apiError.DecodeDetails(&details).
type ErrorDetails struct {
	ID           string   `json:"id"`
	Value        string   `json:"value"`
	ActionIDs    []string `json:"actionIds"`
	ConditionIDs []string `json:"conditionIds"`
}

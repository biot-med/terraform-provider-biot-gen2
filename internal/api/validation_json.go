package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// UnmarshalJSON implements custom JSON unmarshaling for Validation
// to handle defaultValue as any JSON type (string, number, object, array)
// and convert it to a string representation
func (v *Validation) UnmarshalJSON(data []byte) error {
	// Define a temporary struct with interface{} for defaultValue
	type validationAlias struct {
		Mandatory    *bool       `json:"mandatory"`
		DefaultValue interface{} `json:"defaultValue"`
		Min          *float64    `json:"min"`
		Max          *float64    `json:"max"`
		Regex        *string     `json:"regex"`
	}

	var alias validationAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	// Copy simple fields
	v.Mandatory = alias.Mandatory
	v.Min = alias.Min
	v.Max = alias.Max
	v.Regex = alias.Regex

	// Convert defaultValue to string if present
	if alias.DefaultValue != nil {
		var jsonStr string
		// If it's already a string, store it as-is (don't double-encode)
		if str, ok := alias.DefaultValue.(string); ok {
			jsonStr = str
		} else if num, ok := alias.DefaultValue.(float64); ok {
			// If it's a number, convert to plain string representation
			// Use %g to avoid scientific notation for large numbers, but prefer integer format
			if num == float64(int64(num)) {
				jsonStr = fmt.Sprintf("%.0f", num)
			} else {
				jsonStr = fmt.Sprintf("%g", num)
			}
		} else {
			// For objects and arrays, JSON-encode them
			jsonBytes, err := json.Marshal(alias.DefaultValue)
			if err != nil {
				return err
			}
			jsonStr = string(jsonBytes)
		}
		v.DefaultValue = &jsonStr
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for Validation
// to send defaultValue as the actual JSON value (not a JSON string)
// by parsing the stored string representation
func (v Validation) MarshalJSON() ([]byte, error) {
	// Define a temporary struct with interface{} for defaultValue
	type validationAlias struct {
		Mandatory    *bool       `json:"mandatory,omitempty"`
		DefaultValue interface{} `json:"defaultValue,omitempty"`
		Min          *float64    `json:"min,omitempty"`
		Max          *float64    `json:"max,omitempty"`
		Regex        *string     `json:"regex,omitempty"`
	}

	alias := validationAlias{
		Mandatory: v.Mandatory,
		Min:       v.Min,
		Max:       v.Max,
		Regex:     v.Regex,
	}

	// Parse the stored string to get the actual value
	if v.DefaultValue != nil && *v.DefaultValue != "" {
		// First, try to parse as JSON (for objects and arrays)
		var defaultValue interface{}
		if err := json.Unmarshal([]byte(*v.DefaultValue), &defaultValue); err == nil {
			// Successfully parsed as JSON - use the parsed value
			alias.DefaultValue = defaultValue
		} else {
			// Not valid JSON, try to parse as number
			if num, err := strconv.ParseFloat(*v.DefaultValue, 64); err == nil {
				// It's a numeric string, send as number
				alias.DefaultValue = num
			} else {
				// Not a number either, treat as plain string
				alias.DefaultValue = *v.DefaultValue
			}
		}
	}

	return json.Marshal(alias)
}

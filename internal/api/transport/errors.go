package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrNotFound is returned for any 404. It is wrapped around the parsed APIError rather than
// replacing it, because BioT uses 404 for more than one thing: the access-control service
// returns it both for "the object you asked for is missing" and for "an object you
// referenced is missing", and only the error code tells those apart.
//
// So errors.Is(err, ErrNotFound) detects drift, and AsAPIError still recovers the code.
var ErrNotFound = errors.New("resource not found")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// APIError is BioT's standard error envelope. Details is left raw because its shape is
// service-specific; each API package decodes it into its own type via DecodeDetails.
type APIError struct {
	Code        string          `json:"code"`
	Message     string          `json:"message"`
	ServiceName string          `json:"serviceName"`
	TraceID     string          `json:"traceId"`
	Environment string          `json:"environment"`
	Details     json.RawMessage `json:"details"`
}

func (e APIError) Error() string {
	msg, code, traceId := e.Message, e.Code, e.TraceID

	if msg == "" {
		msg = "unknown error message"
	}
	if code == "" {
		code = "unknown error code"
	}
	if traceId == "" {
		traceId = "unknown trace-id"
	}

	return fmt.Sprintf("server error (code: [%s], traceId: [%s]): [%s]", code, traceId, msg)
}

// DecodeDetails decodes the service-specific details payload into target. A response with no
// details leaves target untouched and returns nil.
func (e APIError) DecodeDetails(target any) error {
	if len(e.Details) == 0 {
		return nil
	}

	return json.Unmarshal(e.Details, target)
}

// AsAPIError recovers an APIError from anywhere in an error chain, including from underneath
// the ErrNotFound wrapping.
func AsAPIError(err error) (APIError, bool) {
	var apiError APIError
	ok := errors.As(err, &apiError)

	return apiError, ok
}

// ParseAPIError reads the error envelope from a non-2xx response body.
func ParseAPIError(response *http.Response) error {
	var apiError APIError
	//nolint:errcheck // a body we cannot decode still yields a usable error via the defaults below
	json.NewDecoder(response.Body).Decode(&apiError)

	if apiError.Message == "" {
		apiError.Message = "unknown error message"
	}
	if apiError.Code == "" {
		apiError.Code = "unknown error code"
	}
	if apiError.TraceID == "" {
		apiError.TraceID = "unknown trace-id"
	}

	return apiError
}

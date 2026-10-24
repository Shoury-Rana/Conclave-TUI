package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// APIError represents an HTTP error returned by the Conclave REST API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// parseErrorResponse inspects the HTTP response body returned on a non-2xx status code
// and formats it into a human-readable APIError.
func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	// 1. Try to unmarshal as JSON map: { "error": "...", "detail": "...", "field": [...] }
	var jsonMap map[string]any
	if err := json.Unmarshal(body, &jsonMap); err == nil {
		// Look for standard "error" field (e.g. from Signup/Login views)
		if errMsg, ok := jsonMap["error"].(string); ok && errMsg != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: errMsg}
		}

		// Look for standard Django REST Framework "detail" field
		if detailMsg, ok := jsonMap["detail"].(string); ok && detailMsg != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: detailMsg}
		}

		// Look for field validation errors (e.g. { "email": ["..."], "password": ["..."] })
		var fieldErrors []string
		for field, val := range jsonMap {
			switch v := val.(type) {
			case string:
				fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", field, v))
			case []any:
				var msgs []string
				for _, item := range v {
					msgs = append(msgs, fmt.Sprint(item))
				}
				fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", field, strings.Join(msgs, ", ")))
			}
		}
		if len(fieldErrors) > 0 {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    strings.Join(fieldErrors, " | "),
			}
		}
	}

	// 2. Fallback to raw response text if not structured JSON
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    trimmed,
	}
}

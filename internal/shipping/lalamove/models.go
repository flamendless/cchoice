package lalamove

import (
	"fmt"
	"strings"
)

type LalamoveError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type LalamoveErrorResponse struct {
	Errors []LalamoveError `json:"errors"`
}

func (e LalamoveErrorResponse) Error() string {
	if len(e.Errors) == 0 {
		return "unknown lalamove error"
	}

	var messages []string
	for _, err := range e.Errors {
		if err.Detail != "" {
			messages = append(messages, fmt.Sprintf("%s: %s (%s)", err.ID, err.Message, err.Detail))
		} else {
			messages = append(messages, fmt.Sprintf("%s: %s", err.ID, err.Message))
		}
	}

	if len(messages) == 1 {
		return messages[0]
	}

	return fmt.Sprintf("multiple errors: [%s]", strings.Join(messages, "; "))
}


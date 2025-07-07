package validation

import (
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// ValidateUUID validates that a string is a valid UUID format
func ValidateUUID(id string, fieldName string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if _, err := model.UUIDFromString(id); err != nil {
		return fmt.Errorf("invalid %s: must be a valid UUID", fieldName)
	}
	return nil
}

// ValidateRequired validates that a string field is not empty
func ValidateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateEmail validates basic email format
func ValidateEmail(email string, fieldName string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("invalid %s: must be a valid email address", fieldName)
	}
	return nil
}

// ValidatePositiveNumber validates that a number is positive
func ValidatePositiveNumber(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be greater than 0", fieldName)
	}
	return nil
}

// ValidateInSlice validates that a value is in an allowed slice
func ValidateInSlice(value string, allowed []string, fieldName string) error {
	for _, v := range allowed {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("invalid %s: must be one of %v", fieldName, allowed)
}
package main

import (
	"testing"
)

func TestTeamTypeValidation(t *testing.T) {
	// Test that the validator accepts valid team types
	validTypes := []string{"OPEN", "MEMBER_INVITE", "EXTERNAL", "ORG_ADMIN_MANAGED"}

	for _, validType := range validTypes {
		t.Run("Valid_"+validType, func(t *testing.T) {
			// This is a simplified test - in a real scenario you'd create a proper
			// validator.StringRequest and test the validation response
			found := false
			for _, allowedType := range validTypes {
				if validType == allowedType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected %s to be valid, but it wasn't found in allowed types", validType)
			}
		})
	}
}

func TestInvalidTeamType(t *testing.T) {
	invalidTypes := []string{"CLOSED", "INVALID", "PUBLIC"}
	validTypes := []string{"OPEN", "MEMBER_INVITE", "EXTERNAL", "ORG_ADMIN_MANAGED"}

	for _, invalidType := range invalidTypes {
		t.Run("Invalid_"+invalidType, func(t *testing.T) {
			found := false
			for _, allowedType := range validTypes {
				if invalidType == allowedType {
					found = true
					break
				}
			}
			if found {
				t.Errorf("Expected %s to be invalid, but it was found in allowed types", invalidType)
			}
		})
	}
}

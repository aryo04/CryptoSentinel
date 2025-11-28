package unit

import (
	"testing"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/naming"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

func TestAgentNamingConventions(t *testing.T) {
	t.Run("DefaultValidator", func(t *testing.T) {
		validator := naming.NewDefaultValidator()

		// Test valid names
		validNames := []string{
			"security-scanner",
			"data-processor-v2",
			"my-agent",
			"blockchain-validator",
		}

		for _, name := range validNames {
			result := validator.ValidateName(name)
			if !result.IsValid {
				t.Errorf("Expected '%s' to be valid, but got errors: %v", name, result.Errors)
			}
		}

		// Test invalid names
		invalidNames := []string{
			"a",           // too short
			"123invalid",  // starts with number
			"system",      // reserved
			"agent@name!", // invalid characters
		}

		for _, name := range invalidNames {
			result := validator.ValidateName(name)
			if result.IsValid {
				t.Errorf("Expected '%s' to be invalid, but validation passed", name)
			}
		}
	})

	t.Run("StrictValidator", func(t *testing.T) {
		validator := naming.NewStrictValidator()

		// Test valid strict names
		validNames := []string{
			"security-scanner-agent",
			"data-processor-agent",
		}

		for _, name := range validNames {
			result := validator.ValidateName(name)
			if !result.IsValid {
				t.Errorf("Expected '%s' to be valid with strict rules, but got errors: %v", name, result.Errors)
			}
		}

		// Test invalid strict names
		invalidNames := []string{
			"SecurityAgent",    // uppercase
			"security_agent",   // underscores
			"security-scanner", // missing -agent suffix
		}

		for _, name := range invalidNames {
			result := validator.ValidateName(name)
			if result.IsValid {
				t.Errorf("Expected '%s' to be invalid with strict rules, but validation passed", name)
			}
		}
	})
}

func TestNameNormalization(t *testing.T) {
	validator := naming.NewDefaultValidator()

	tests := []struct {
		input    string
		expected string
	}{
		{"Security Scanner", "securityscanner"},
		{"DATA_PROCESSOR", "data_processor"},
		{"agent@name!", "agentname"},
		{"MyAwesome-Agent", "myawesome-agent"},
	}

	for _, tt := range tests {
		result := validator.NormalizeName(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeName('%s') = '%s', expected '%s'", tt.input, result, tt.expected)
		}
	}
}

func TestNameGeneration(t *testing.T) {
	validator := naming.NewDefaultValidator()

	tests := []struct {
		baseName string
		purpose  string
		contains []string
	}{
		{"security", "scanner", []string{"security", "scanner"}},
		{"data", "analyzer", []string{"data", "analyzer"}},
		{"", "scanner", []string{"scanner"}},
		{"auth", "", []string{"auth"}},
		{"", "", []string{"custom"}},
	}

	for _, tt := range tests {
		result := validator.GenerateName(tt.baseName, tt.purpose)

		// Validate the generated name
		validation := validator.ValidateName(result)
		if !validation.IsValid {
			t.Errorf("Generated name '%s' is not valid: %v", result, validation.Errors)
		}

		// Check if it contains expected parts
		for _, expected := range tt.contains {
			if !contains(result, expected) {
				t.Errorf("Generated name '%s' should contain '%s'", result, expected)
			}
		}
	}
}

func TestSuggestions(t *testing.T) {
	validator := naming.NewDefaultValidator()

	invalidNames := []string{
		"123InvalidAgent",
		"agent@name!",
		"a",
	}

	for _, invalidName := range invalidNames {
		suggestions := validator.SuggestNames(invalidName, 3)

		if len(suggestions) == 0 {
			t.Errorf("Expected suggestions for '%s', but got none", invalidName)
			continue
		}

		// Validate that suggestions are actually valid
		for _, suggestion := range suggestions {
			result := validator.ValidateName(suggestion)
			if !result.IsValid {
				t.Errorf("Suggestion '%s' for '%s' is not valid: %v", suggestion, invalidName, result.Errors)
			}
		}
	}
}

func TestAgentConfigIntegration(t *testing.T) {
	validator := naming.NewDefaultValidator()

	// Valid config
	validConfig := &types.AgentConfig{
		Name: "security-scanner-v2",
		NamingRules: &types.AgentNamingRules{
			MaxLength:        30,
			MinLength:        5,
			CaseSensitive:    false,
			AllowNumbers:     true,
			AllowHyphens:     true,
			AllowUnderscores: false,
		},
	}

	result := validator.ValidateAgentConfig(validConfig)
	if !result.IsValid {
		t.Errorf("Expected valid config to pass validation, but got errors: %v", result.Errors)
	}

	// Invalid config
	invalidConfig := &types.AgentConfig{
		Name: "InvalidName123!",
		NamingRules: &types.AgentNamingRules{
			MaxLength:        20,
			MinLength:        8,
			CaseSensitive:    true,
			AllowNumbers:     false,
			AllowHyphens:     true,
			AllowUnderscores: false,
			RequiredSuffix:   "-agent",
		},
	}

	result = validator.ValidateAgentConfig(invalidConfig)
	if result.IsValid {
		t.Error("Expected invalid config to fail validation")
	}
}

func TestReservedNames(t *testing.T) {
	validator := naming.NewDefaultValidator()

	reservedNames := []string{
		"system", "admin", "root", "coordinator",
		"teneo", "protocol", "network",
		"api", "gateway", "proxy",
		"agent", "bot", "service",
		"test", "demo", "example",
	}

	for _, name := range reservedNames {
		result := validator.ValidateName(name)
		if result.IsValid {
			t.Errorf("Reserved name '%s' should not be valid", name)
		}
	}
}

func TestCustomRules(t *testing.T) {
	// Use the built-in strict validator instead of creating custom rules manually
	validator := naming.NewStrictValidator()

	// Should be valid with strict rules
	validName := "custom-security-agent"
	result := validator.ValidateName(validName)
	if !result.IsValid {
		t.Errorf("Expected '%s' to be valid with strict rules, but got errors: %v", validName, result.Errors)
	}

	// Should be invalid - missing -agent suffix
	invalidName := "security-scanner"
	result = validator.ValidateName(invalidName)
	if result.IsValid {
		t.Errorf("Expected '%s' to be invalid without required suffix", invalidName)
	}
}

// helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package naming

import (
	"testing"
)

func TestValidateAgentName(t *testing.T) {
	tests := []struct {
		name           string
		agentName      string
		rules          *AgentNamingRules
		expectedValid  bool
		expectedErrors int
	}{
		{
			name:           "valid default name",
			agentName:      "my-test-agent",
			rules:          DefaultAgentNamingRules,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "valid name with numbers",
			agentName:      "agent-v2",
			rules:          DefaultAgentNamingRules,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "valid name with underscores",
			agentName:      "data_processor_agent",
			rules:          DefaultAgentNamingRules,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "too short",
			agentName:      "ab",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "too long",
			agentName:      "this-is-a-very-long-agent-name-that-exceeds-the-maximum-length-allowed",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "reserved name",
			agentName:      "system",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "starts with number",
			agentName:      "123agent",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "invalid characters",
			agentName:      "agent@name!",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "empty name",
			agentName:      "",
			rules:          DefaultAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "strict rules - valid",
			agentName:      "security-scanner-agent",
			rules:          StrictAgentNamingRules,
			expectedValid:  true,
			expectedErrors: 0,
		},
		{
			name:           "strict rules - missing suffix",
			agentName:      "security-scanner",
			rules:          StrictAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 1,
		},
		{
			name:           "strict rules - uppercase not allowed",
			agentName:      "Security-Agent",
			rules:          StrictAgentNamingRules,
			expectedValid:  false,
			expectedErrors: 2, // pattern + missing suffix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAgentName(tt.agentName, tt.rules)
			
			if result.IsValid != tt.expectedValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectedValid, result.IsValid)
			}
			
			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("expected %d errors, got %d errors: %v", tt.expectedErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

func TestNormalizeAgentName(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		rules          *AgentNamingRules
		expectedOutput string
	}{
		{
			name:           "simple normalization",
			input:          "My Test Agent",
			rules:          DefaultAgentNamingRules,
			expectedOutput: "mytestagent",
		},
		{
			name:           "remove invalid characters",
			input:          "agent@name!",
			rules:          DefaultAgentNamingRules,
			expectedOutput: "agentname",
		},
		{
			name:           "handle underscores to hyphens",
			input:          "data_processor_agent",
			rules:          DefaultAgentNamingRules,
			expectedOutput: "data_processor_agent",
		},
		{
			name:           "truncate long names",
			input:          "this-is-a-very-long-agent-name-that-exceeds-maximum",
			rules:          DefaultAgentNamingRules,
			expectedOutput: "this-is-a-very-long-agent-name-that-exceeds-maximu",
		},
		{
			name:           "strict rules with suffix",
			input:          "security-scanner",
			rules:          StrictAgentNamingRules,
			expectedOutput: "security-scanner-agent",
		},
		{
			name:           "pad short names",
			input:          "ai",
			rules:          DefaultAgentNamingRules,
			expectedOutput: "ai1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAgentName(tt.input, tt.rules)
			
			if result != tt.expectedOutput {
				t.Errorf("expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

func TestGenerateAgentName(t *testing.T) {
	tests := []struct {
		name       string
		baseName   string
		purpose    string
		rules      *AgentNamingRules
		shouldContain []string
	}{
		{
			name:     "generate with base and purpose",
			baseName: "security",
			purpose:  "scanner",
			rules:    DefaultAgentNamingRules,
			shouldContain: []string{"security", "scanner"},
		},
		{
			name:     "generate with only purpose",
			baseName: "",
			purpose:  "scanner",
			rules:    DefaultAgentNamingRules,
			shouldContain: []string{"scanner"},
		},
		{
			name:     "generate default when empty",
			baseName: "",
			purpose:  "",
			rules:    DefaultAgentNamingRules,
			shouldContain: []string{"custom"},
		},
		{
			name:     "generate with strict rules",
			baseName: "data",
			purpose:  "processor",
			rules:    StrictAgentNamingRules,
			shouldContain: []string{"data", "processor", "agent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateAgentName(tt.baseName, tt.purpose, tt.rules)
			
			// validate the generated name
			validation := ValidateAgentName(result, tt.rules)
			if !validation.IsValid {
				t.Errorf("generated name '%s' is not valid: %v", result, validation.Errors)
			}
			
			// check if it contains expected parts
			for _, expected := range tt.shouldContain {
				if !containsSubstring(result, expected) {
					t.Errorf("generated name '%s' should contain '%s'", result, expected)
				}
			}
		})
	}
}

func TestReservedNames(t *testing.T) {
	reservedNames := []string{
		"system", "admin", "root", "coordinator", "teneo", "protocol",
		"api", "agent", "test", "demo",
	}

	for _, name := range reservedNames {
		t.Run("reserved_"+name, func(t *testing.T) {
			result := ValidateAgentName(name, DefaultAgentNamingRules)
			if result.IsValid {
				t.Errorf("reserved name '%s' should not be valid", name)
			}
		})
	}
}

func TestCaseSensitivity(t *testing.T) {
	tests := []struct {
		name           string
		agentName      string
		caseSensitive  bool
		expectedNormal string
	}{
		{
			name:           "case insensitive",
			agentName:      "MyTestAgent",
			caseSensitive:  false,
			expectedNormal: "mytestagent",
		},
		{
			name:           "case sensitive",
			agentName:      "MyTestAgent",
			caseSensitive:  true,
			expectedNormal: "MyTestAgent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := &AgentNamingRules{
				MaxLength:      50,
				MinLength:      3,
				AllowedPattern: DefaultAgentNamingRules.AllowedPattern,
				ReservedNames:  DefaultAgentNamingRules.ReservedNames,
				CaseSensitive:  tt.caseSensitive,
				AllowNumbers:   true,
				AllowHyphens:   true,
				AllowUnderscores: true,
			}
			
			result := ValidateAgentName(tt.agentName, rules)
			if result.NormalizedName != tt.expectedNormal {
				t.Errorf("expected normalized name '%s', got '%s'", tt.expectedNormal, result.NormalizedName)
			}
		})
	}
}

// helper function
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
		 (s[:len(substr)] == substr || 
		  s[len(s)-len(substr):] == substr ||
		  hasSubstring(s, substr))))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

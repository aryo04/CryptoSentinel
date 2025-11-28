package naming

import (
	"fmt"
	"regexp"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// AgentNameValidator provides validation functionality for agent names
type AgentNameValidator struct {
	rules *AgentNamingRules
}

// NewAgentNameValidator creates a new validator with the specified rules
func NewAgentNameValidator(rules *AgentNamingRules) *AgentNameValidator {
	if rules == nil {
		rules = DefaultAgentNamingRules
	}
	return &AgentNameValidator{rules: rules}
}

// NewDefaultValidator creates a validator with default rules
func NewDefaultValidator() *AgentNameValidator {
	return NewAgentNameValidator(DefaultAgentNamingRules)
}

// NewStrictValidator creates a validator with strict rules
func NewStrictValidator() *AgentNameValidator {
	return NewAgentNameValidator(StrictAgentNamingRules)
}

// ValidateName validates an agent name and returns validation result
func (v *AgentNameValidator) ValidateName(name string) *types.AgentNameValidation {
	result := ValidateAgentName(name, v.rules)

	return &types.AgentNameValidation{
		IsValid:        result.IsValid,
		NormalizedName: result.NormalizedName,
		Errors:         result.Errors,
		Warnings:       result.Warnings,
	}
}

// NormalizeName normalizes an agent name according to the validator's rules
func (v *AgentNameValidator) NormalizeName(name string) string {
	return NormalizeAgentName(name, v.rules)
}

// GenerateName generates a valid agent name based on base name and purpose
func (v *AgentNameValidator) GenerateName(baseName, purpose string) string {
	return GenerateAgentName(baseName, purpose, v.rules)
}

// ValidateAgentConfig validates the agent name in an AgentConfig
func (v *AgentNameValidator) ValidateAgentConfig(config *types.AgentConfig) *types.AgentNameValidation {
	if config == nil {
		return &types.AgentNameValidation{
			IsValid: false,
			Errors:  []string{"agent config cannot be nil"},
		}
	}

	// use config-specific naming rules if provided
	rules := v.rules
	if config.NamingRules != nil {
		rules = convertToAgentNamingRules(config.NamingRules)
	}

	result := ValidateAgentName(config.Name, rules)

	return &types.AgentNameValidation{
		IsValid:        result.IsValid,
		NormalizedName: result.NormalizedName,
		Errors:         result.Errors,
		Warnings:       result.Warnings,
	}
}

// SuggestNames suggests alternative valid names based on an invalid name
func (v *AgentNameValidator) SuggestNames(invalidName string, count int) []string {
	if count <= 0 {
		count = 3
	}

	suggestions := make([]string, 0, count)

	// try normalization first
	normalized := v.NormalizeName(invalidName)
	if validation := v.ValidateName(normalized); validation.IsValid {
		suggestions = append(suggestions, normalized)
	}

	// generate variations
	baseName := extractBaseName(invalidName)
	purposes := []string{"agent", "bot", "service", "handler", "processor"}

	for i, purpose := range purposes {
		if len(suggestions) >= count {
			break
		}

		generated := v.GenerateName(baseName, purpose)
		if validation := v.ValidateName(generated); validation.IsValid {
			// avoid duplicates
			if !contains(suggestions, generated) {
				suggestions = append(suggestions, generated)
			}
		}

		// try with numbers
		if len(suggestions) < count && v.rules.AllowNumbers {
			numberedName := fmt.Sprintf("%s%d", generated, i+1)
			if validation := v.ValidateName(numberedName); validation.IsValid {
				if !contains(suggestions, numberedName) {
					suggestions = append(suggestions, numberedName)
				}
			}
		}
	}

	return suggestions
}

// GetRules returns the current naming rules
func (v *AgentNameValidator) GetRules() *AgentNamingRules {
	return v.rules
}

// UpdateRules updates the validator's naming rules
func (v *AgentNameValidator) UpdateRules(rules *AgentNamingRules) {
	if rules != nil {
		v.rules = rules
	}
}

// convertToAgentNamingRules converts types.AgentNamingRules to internal AgentNamingRules
func convertToAgentNamingRules(rules *types.AgentNamingRules) *AgentNamingRules {
	if rules == nil {
		return DefaultAgentNamingRules
	}

	// convert reserved names slice to map
	reservedMap := make(map[string]bool)
	for _, name := range rules.ReservedNames {
		reservedMap[name] = true
	}

	// if no reserved names provided, use defaults
	if len(reservedMap) == 0 {
		reservedMap = getReservedNames()
	}

	// create pattern based on rules
	pattern := createPatternFromRules(rules)

	return &AgentNamingRules{
		MaxLength:        getIntOrDefault(rules.MaxLength, 50),
		MinLength:        getIntOrDefault(rules.MinLength, 3),
		AllowedPattern:   pattern,
		ReservedNames:    reservedMap,
		RequiredPrefix:   rules.RequiredPrefix,
		RequiredSuffix:   rules.RequiredSuffix,
		CaseSensitive:    rules.CaseSensitive,
		AllowNumbers:     rules.AllowNumbers,
		AllowHyphens:     rules.AllowHyphens,
		AllowUnderscores: rules.AllowUnderscores,
	}
}

// createPatternFromRules creates a regex pattern based on naming rules
func createPatternFromRules(rules *types.AgentNamingRules) *regexp.Regexp {
	// build character class
	charClass := "a-zA-Z"

	if rules.AllowNumbers {
		charClass += "0-9"
	}

	var middleChars string
	if rules.AllowHyphens {
		middleChars += "\\-"
	}
	if rules.AllowUnderscores {
		middleChars += "_"
	}

	// construct pattern: starts with letter, middle can have allowed chars, ends with letter or number
	var pattern string
	if middleChars != "" {
		pattern = fmt.Sprintf("^[a-zA-Z][%s%s]*[%s]$", charClass, middleChars, charClass)
	} else {
		pattern = fmt.Sprintf("^[%s]+$", charClass)
	}

	// case sensitivity
	if !rules.CaseSensitive {
		pattern = "(?i)" + pattern
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// fallback to default pattern
		return DefaultAgentNamingRules.AllowedPattern
	}

	return compiled
}

// helper functions

func getIntOrDefault(value, defaultValue int) int {
	if value <= 0 {
		return defaultValue
	}
	return value
}

func extractBaseName(name string) string {
	// simple extraction - remove common suffixes and prefixes
	cleanName := name

	// remove common prefixes
	prefixes := []string{"teneo-", "agent-", "bot-", "service-"}
	for _, prefix := range prefixes {
		if len(cleanName) > len(prefix) && cleanName[:len(prefix)] == prefix {
			cleanName = cleanName[len(prefix):]
			break
		}
	}

	// remove common suffixes
	suffixes := []string{"-agent", "-bot", "-service", "-handler"}
	for _, suffix := range suffixes {
		if len(cleanName) > len(suffix) && cleanName[len(cleanName)-len(suffix):] == suffix {
			cleanName = cleanName[:len(cleanName)-len(suffix)]
			break
		}
	}

	return cleanName
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

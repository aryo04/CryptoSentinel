package naming

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// AgentNamingRules defines the naming conventions for agents
type AgentNamingRules struct {
	MaxLength       int
	MinLength       int
	AllowedPattern  *regexp.Regexp
	ReservedNames   map[string]bool
	RequiredPrefix  string
	RequiredSuffix  string
	CaseSensitive   bool
	AllowNumbers    bool
	AllowHyphens    bool
	AllowUnderscores bool
}

// ValidationResult represents the result of name validation
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	NormalizedName string `json:"normalized_name,omitempty"`
}

// Default naming rules for Teneo agents
var DefaultAgentNamingRules = &AgentNamingRules{
	MaxLength:        50,
	MinLength:        3,
	AllowedPattern:   regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-_]*[a-zA-Z0-9]$`),
	ReservedNames:    getReservedNames(),
	RequiredPrefix:   "",
	RequiredSuffix:   "",
	CaseSensitive:    false,
	AllowNumbers:     true,
	AllowHyphens:     true,
	AllowUnderscores: true,
}

// Strict naming rules for production environments
var StrictAgentNamingRules = &AgentNamingRules{
	MaxLength:        30,
	MinLength:        5,
	AllowedPattern:   regexp.MustCompile(`^[a-z][a-z0-9\-]*[a-z0-9]$`),
	ReservedNames:    getReservedNames(),
	RequiredPrefix:   "",
	RequiredSuffix:   "-agent",
	CaseSensitive:    true,
	AllowNumbers:     true,
	AllowHyphens:     true,
	AllowUnderscores: false,
}

// getReservedNames returns a map of reserved agent names
func getReservedNames() map[string]bool {
	reserved := map[string]bool{
		// System reserved
		"system":      true,
		"admin":       true,
		"root":        true,
		"coordinator": true,
		"manager":     true,
		"supervisor":  true,
		"monitor":     true,
		
		// Protocol reserved
		"teneo":       true,
		"protocol":    true,
		"network":     true,
		"blockchain":  true,
		"validator":   true,
		"consensus":   true,
		
		// Service reserved
		"api":         true,
		"gateway":     true,
		"proxy":       true,
		"load-balancer": true,
		"health":      true,
		"metrics":     true,
		"logging":     true,
		
		// Common terms
		"agent":       true,
		"bot":         true,
		"service":     true,
		"handler":     true,
		"processor":   true,
		"worker":      true,
		"client":      true,
		"server":      true,
		
		// Test/Development
		"test":        true,
		"demo":        true,
		"example":     true,
		"sample":      true,
		"mock":        true,
		"stub":        true,
		"dev":         true,
		"debug":       true,
	}
	
	return reserved
}

// ValidateAgentName validates an agent name against the specified rules
func ValidateAgentName(name string, rules *AgentNamingRules) *ValidationResult {
	if rules == nil {
		rules = DefaultAgentNamingRules
	}
	
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}
	
	// normalize name for validation
	normalizedName := strings.TrimSpace(name)
	if !rules.CaseSensitive {
		normalizedName = strings.ToLower(normalizedName)
	}
	result.NormalizedName = normalizedName
	
	// check if name is empty
	if normalizedName == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "agent name cannot be empty")
		return result
	}
	
	// check length constraints
	if len(normalizedName) < rules.MinLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("agent name must be at least %d characters long", rules.MinLength))
	}
	
	if len(normalizedName) > rules.MaxLength {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("agent name must not exceed %d characters", rules.MaxLength))
	}
	
	// check pattern
	if !rules.AllowedPattern.MatchString(normalizedName) {
		result.IsValid = false
		result.Errors = append(result.Errors, "agent name contains invalid characters or format")
	}
	
	// check reserved names
	if rules.ReservedNames[normalizedName] {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("'%s' is a reserved name and cannot be used", normalizedName))
	}
	
	// check required prefix
	if rules.RequiredPrefix != "" && !strings.HasPrefix(normalizedName, rules.RequiredPrefix) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("agent name must start with '%s'", rules.RequiredPrefix))
	}
	
	// check required suffix
	if rules.RequiredSuffix != "" && !strings.HasSuffix(normalizedName, rules.RequiredSuffix) {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("agent name must end with '%s'", rules.RequiredSuffix))
	}
	
	// additional character validations
	if !rules.AllowNumbers && containsNumbers(normalizedName) {
		result.IsValid = false
		result.Errors = append(result.Errors, "agent name cannot contain numbers")
	}
	
	if !rules.AllowHyphens && strings.Contains(normalizedName, "-") {
		result.IsValid = false
		result.Errors = append(result.Errors, "agent name cannot contain hyphens")
	}
	
	if !rules.AllowUnderscores && strings.Contains(normalizedName, "_") {
		result.IsValid = false
		result.Errors = append(result.Errors, "agent name cannot contain underscores")
	}
	
	// generate warnings for best practices
	if startsWithNumber(normalizedName) {
		result.Warnings = append(result.Warnings, "agent name should not start with a number")
	}
	
	if containsConsecutiveSpecialChars(normalizedName) {
		result.Warnings = append(result.Warnings, "avoid consecutive special characters for better readability")
	}
	
	if len(normalizedName) > 30 {
		result.Warnings = append(result.Warnings, "consider shorter names for better usability")
	}
	
	if isAbbreviation(normalizedName) {
		result.Warnings = append(result.Warnings, "consider using descriptive names instead of abbreviations")
	}
	
	return result
}

// NormalizeAgentName normalizes an agent name according to rules
func NormalizeAgentName(name string, rules *AgentNamingRules) string {
	if rules == nil {
		rules = DefaultAgentNamingRules
	}
	
	// trim whitespace
	normalized := strings.TrimSpace(name)
	
	// handle case sensitivity
	if !rules.CaseSensitive {
		normalized = strings.ToLower(normalized)
	}
	
	// replace invalid characters with allowed ones
	if !rules.AllowUnderscores && rules.AllowHyphens {
		normalized = strings.ReplaceAll(normalized, "_", "-")
	} else if !rules.AllowHyphens && rules.AllowUnderscores {
		normalized = strings.ReplaceAll(normalized, "-", "_")
	}
	
	// remove invalid characters
	var result strings.Builder
	for _, r := range normalized {
		if isValidCharacter(r, rules) {
			result.WriteRune(r)
		}
	}
	
	normalized = result.String()
	
	// ensure it starts with a letter
	if len(normalized) > 0 && !unicode.IsLetter(rune(normalized[0])) {
		normalized = "agent-" + normalized
	}
	
	// ensure it doesn't end with special characters
	normalized = strings.TrimRight(normalized, "-_")
	
	// add required prefix/suffix
	if rules.RequiredPrefix != "" && !strings.HasPrefix(normalized, rules.RequiredPrefix) {
		normalized = rules.RequiredPrefix + normalized
	}
	
	if rules.RequiredSuffix != "" && !strings.HasSuffix(normalized, rules.RequiredSuffix) {
		normalized = normalized + rules.RequiredSuffix
	}
	
	// ensure length constraints
	if len(normalized) > rules.MaxLength {
		normalized = normalized[:rules.MaxLength]
		// ensure it doesn't end with special character after truncation
		normalized = strings.TrimRight(normalized, "-_")
	}
	
	// ensure minimum length
	if len(normalized) < rules.MinLength {
		// pad with default suffix if too short
		for len(normalized) < rules.MinLength {
			if rules.AllowNumbers {
				normalized += "1"
			} else {
				normalized += "x"
			}
		}
	}
	
	return normalized
}

// GenerateAgentName generates a valid agent name based on a base name and purpose
func GenerateAgentName(baseName, purpose string, rules *AgentNamingRules) string {
	if rules == nil {
		rules = DefaultAgentNamingRules
	}
	
	var parts []string
	
	if baseName != "" {
		parts = append(parts, baseName)
	}
	
	if purpose != "" {
		parts = append(parts, purpose)
	}
	
	if len(parts) == 0 {
		parts = append(parts, "custom")
	}
	
	// join parts with appropriate separator
	separator := "-"
	if !rules.AllowHyphens && rules.AllowUnderscores {
		separator = "_"
	}
	
	generated := strings.Join(parts, separator)
	return NormalizeAgentName(generated, rules)
}

// helper functions

func containsNumbers(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func startsWithNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	return unicode.IsDigit(rune(s[0]))
}

func containsConsecutiveSpecialChars(s string) bool {
	prev := ' '
	for _, r := range s {
		if (r == '-' || r == '_') && (prev == '-' || prev == '_') {
			return true
		}
		prev = r
	}
	return false
}

func isAbbreviation(s string) bool {
	// simple heuristic: if it's short and mostly uppercase
	if len(s) <= 4 {
		upperCount := 0
		for _, r := range s {
			if unicode.IsUpper(r) {
				upperCount++
			}
		}
		return float64(upperCount)/float64(len(s)) > 0.6
	}
	return false
}

func isValidCharacter(r rune, rules *AgentNamingRules) bool {
	if unicode.IsLetter(r) {
		return true
	}
	
	if rules.AllowNumbers && unicode.IsDigit(r) {
		return true
	}
	
	if rules.AllowHyphens && r == '-' {
		return true
	}
	
	if rules.AllowUnderscores && r == '_' {
		return true
	}
	
	return false
}

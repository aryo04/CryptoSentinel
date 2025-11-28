# Agent Naming Conventions

This document defines the standardized naming conventions for agents in the Teneo Agent SDK to ensure consistency and uniformity across the platform.

## Overview

Agent naming conventions help maintain a consistent ecosystem where agents are easily identifiable, searchable, and manageable. The SDK provides validation, normalization, and generation utilities to enforce these conventions.

## Naming Rules

### Default Rules

The default naming rules provide a balanced approach suitable for most use cases:

- **Length**: 3-50 characters
- **Format**: Must start with a letter, can contain letters, numbers, hyphens, and underscores
- **Case**: Case-insensitive (normalized to lowercase)
- **Pattern**: `^[a-zA-Z][a-zA-Z0-9\-_]*[a-zA-Z0-9]$`

### Strict Rules

For production environments requiring more stringent naming:

- **Length**: 5-30 characters
- **Format**: Must start with a letter, lowercase only, hyphens allowed, no underscores
- **Required Suffix**: `-agent`
- **Case**: Case-sensitive (must be lowercase)
- **Pattern**: `^[a-z][a-z0-9\-]*[a-z0-9]$`

## Reserved Names

The following names are reserved and cannot be used for agents:

### System Reserved
- `system`, `admin`, `root`, `coordinator`, `manager`, `supervisor`, `monitor`

### Protocol Reserved  
- `teneo`, `protocol`, `network`, `blockchain`, `validator`, `consensus`

### Service Reserved
- `api`, `gateway`, `proxy`, `load-balancer`, `health`, `metrics`, `logging`

### Common Terms
- `agent`, `bot`, `service`, `handler`, `processor`, `worker`, `client`, `server`

### Test/Development
- `test`, `demo`, `example`, `sample`, `mock`, `stub`, `dev`, `debug`

## Usage Examples

### Basic Validation

```go
import "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/naming"

// Create validator with default rules
validator := naming.NewDefaultValidator()

// Validate a name
result := validator.ValidateName("my-security-agent")
if result.IsValid {
    fmt.Printf("Valid name: %s\n", result.NormalizedName)
} else {
    fmt.Printf("Invalid name. Errors: %v\n", result.Errors)
}
```

### Using Strict Rules

```go
// Create validator with strict rules
validator := naming.NewStrictValidator()

// Validate with strict rules
result := validator.ValidateName("SecurityScanner")
if !result.IsValid {
    fmt.Printf("Errors: %v\n", result.Errors)
    // Output: Errors: [agent name contains invalid characters or format, agent name must end with '-agent']
}

// Normalize the name
normalized := validator.NormalizeName("SecurityScanner")
// Output: security-scanner-agent
```

### Custom Rules

```go
import "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"

// Define custom rules
customRules := &naming.AgentNamingRules{
    MaxLength:        25,
    MinLength:        5,
    AllowedPattern:   regexp.MustCompile(`^[a-z][a-z0-9]*[a-z0-9]$`),
    ReservedNames:    map[string]bool{"forbidden": true},
    RequiredPrefix:   "custom-",
    RequiredSuffix:   "",
    CaseSensitive:    true,
    AllowNumbers:     true,
    AllowHyphens:     false,
    AllowUnderscores: false,
}

validator := naming.NewAgentNameValidator(customRules)
```

### Agent Configuration Integration

```go
import "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"

// Configure agent with naming rules
config := &types.AgentConfig{
    Name: "my-agent",
    NamingRules: &types.AgentNamingRules{
        MaxLength:        30,
        MinLength:        5,
        CaseSensitive:    false,
        AllowNumbers:     true,
        AllowHyphens:     true,
        AllowUnderscores: false,
        RequiredSuffix:   "-agent",
    },
}

// Validate the configuration
validator := naming.NewDefaultValidator()
result := validator.ValidateAgentConfig(config)
```

### Name Generation

```go
// Generate valid names
validator := naming.NewDefaultValidator()

// Generate from base name and purpose
name := validator.GenerateName("security", "scanner")
// Output: security-scanner

// Get suggestions for invalid names
suggestions := validator.SuggestNames("123InvalidName!", 3)
// Output: [agent-invalidname, invalidname-agent, invalidname-bot]
```

## Best Practices

### ✅ Good Examples

```
security-scanner-agent
data-processor
api-gateway-v2
content-analyzer
blockchain-validator
ml-inference-engine
```

### ❌ Bad Examples

```
SecurityScanner         // Should be lowercase with hyphens
123agent               // Cannot start with number  
agent@name             // Invalid characters
a                      // Too short
very-long-agent-name-that-exceeds-maximum-length  // Too long
system                 // Reserved name
```

### Naming Guidelines

1. **Be Descriptive**: Use names that clearly indicate the agent's purpose
   - Good: `fraud-detection-agent`
   - Bad: `agent1`

2. **Use Hyphens for Separation**: Prefer hyphens over underscores for readability
   - Good: `sentiment-analysis-bot`
   - Bad: `sentiment_analysis_bot`

3. **Follow Hierarchical Naming**: Use logical grouping for related agents
   - `trading-risk-analyzer`
   - `trading-signal-generator`
   - `trading-portfolio-manager`

4. **Avoid Abbreviations**: Use full words when possible
   - Good: `document-processor`
   - Bad: `doc-proc`

5. **Include Version When Needed**: For multiple versions of the same agent
   - `fraud-detector-v2`
   - `legacy-data-importer`

## Validation Features

### Error Types

The validator provides detailed error messages:

- **Length violations**: Name too short or too long
- **Character violations**: Invalid characters used
- **Pattern violations**: Doesn't match required format
- **Reserved names**: Attempting to use reserved names
- **Prefix/suffix violations**: Missing required prefix or suffix

### Warnings

The validator also provides warnings for best practices:

- Names starting with numbers
- Consecutive special characters
- Very long names (even if within limits)
- Abbreviations detected

### Normalization

The normalizer automatically:

- Converts to appropriate case
- Replaces invalid characters with valid ones
- Adds required prefixes/suffixes
- Ensures proper length
- Removes invalid characters

## Integration with Agent SDK

### Automatic Validation

When creating agents, the SDK automatically validates names:

```go
// This will validate the agent name during creation
agent, err := sdk.NewAgent(&types.AgentConfig{
    Name: "my-custom-agent",
    // ... other config
})
```

### Configuration-Based Rules

Different environments can use different naming rules:

```yaml
# development.yaml
agent:
  naming_rules:
    max_length: 50
    case_sensitive: false
    required_suffix: ""

# production.yaml  
agent:
  naming_rules:
    max_length: 30
    case_sensitive: true
    required_suffix: "-agent"
```

## CLI Commands

The Teneo CLI provides commands for name validation:

```bash
# Validate a name
teneo agent validate-name "my-agent-name"

# Generate suggestions
teneo agent suggest-names "InvalidName123"

# Check naming rules
teneo agent naming-rules --environment production
```

## Migration Guide

### Updating Existing Agents

If you have existing agents with non-compliant names:

1. **Assess Current Names**: Run validation on existing names
2. **Plan Migration**: Use suggestion tools to find compliant alternatives  
3. **Update Gradually**: Migrate agents during maintenance windows
4. **Use Aliases**: Maintain backward compatibility with name aliases

### Example Migration

```go
// Check existing agents
existingAgents := []string{"Agent1", "DATA_PROC", "sys-monitor"}

validator := naming.NewStrictValidator()
for _, name := range existingAgents {
    result := validator.ValidateName(name)
    if !result.IsValid {
        suggestions := validator.SuggestNames(name, 3)
        fmt.Printf("Agent '%s' needs migration. Suggestions: %v\n", name, suggestions)
    }
}

// Output:
// Agent 'Agent1' needs migration. Suggestions: [agent1-agent, agent-handler, agent-bot]
// Agent 'DATA_PROC' needs migration. Suggestions: [data-proc-agent, dataproc-agent, data-processor-agent]  
// Agent 'sys-monitor' needs migration. Suggestions: [sys-monitor-agent, system-monitor-agent, monitor-agent]
```

## API Reference

### Core Functions

- `ValidateAgentName(name string, rules *AgentNamingRules) *ValidationResult`
- `NormalizeAgentName(name string, rules *AgentNamingRules) string`
- `GenerateAgentName(baseName, purpose string, rules *AgentNamingRules) string`

### Validator Methods

- `NewAgentNameValidator(rules *AgentNamingRules) *AgentNameValidator`
- `ValidateName(name string) *AgentNameValidation`
- `NormalizeName(name string) string`
- `GenerateName(baseName, purpose string) string`
- `SuggestNames(invalidName string, count int) []string`
- `ValidateAgentConfig(config *AgentConfig) *AgentNameValidation`

### Configuration Types

- `AgentNamingRules`: Defines naming rules and constraints
- `AgentNameValidation`: Contains validation results with errors and warnings
- `ValidationResult`: Internal validation result structure

## Testing

The naming conventions include comprehensive tests:

```bash
# Run naming convention tests
go test ./pkg/naming/...

# Run with coverage
go test -cover ./pkg/naming/...

# Benchmark validation performance
go test -bench=. ./pkg/naming/...
```

## Conclusion

Standardized naming conventions ensure that the Teneo agent ecosystem remains organized, searchable, and maintainable. By following these guidelines and using the provided validation tools, developers can create agents that integrate seamlessly with the platform while maintaining consistency across all deployments.

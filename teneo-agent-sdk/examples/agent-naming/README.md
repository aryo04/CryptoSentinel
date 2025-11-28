# Agent Naming Conventions Example

This example demonstrates the standardized agent naming conventions implemented in the Teneo Agent SDK.

## Overview

The Agent SDK provides comprehensive naming validation, normalization, and generation tools to ensure consistent agent naming across the platform.

## Features Demonstrated

- **Basic Name Validation**: Validate agent names against default rules
- **Name Normalization**: Clean and standardize agent names
- **Name Generation**: Generate valid names from base components
- **Custom Rules**: Use strict production rules for different environments
- **Suggestions**: Get suggestions for invalid names
- **Config Integration**: Validate names in agent configurations

## Running the Example

```bash
cd examples/agent-naming
go run main.go
```

## Key Features

### Validation Rules

- **Length**: 3-50 characters (default), 5-30 (strict)
- **Format**: Must start with letter, can contain letters, numbers, hyphens, underscores
- **Reserved Names**: System, protocol, and service names are protected
- **Case Sensitivity**: Configurable (default: case-insensitive)

### Normalization

- Converts to appropriate case
- Removes or replaces invalid characters
- Ensures proper length constraints
- Adds required prefixes/suffixes

### Generation

- Create names from base name + purpose
- Handles empty inputs gracefully
- Applies all validation rules automatically

## Usage in Code

```go
import "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/naming"

// Basic validation
validator := naming.NewDefaultValidator()
result := validator.ValidateName("my-agent")

// Strict validation
strictValidator := naming.NewStrictValidator()
result := strictValidator.ValidateName("my-agent")

// Name generation
name := validator.GenerateName("security", "scanner")
// Returns: "security-scanner"

// Get suggestions
suggestions := validator.SuggestNames("InvalidName!", 3)
```

## Configuration

Agent configurations can include naming rules:

```go
config := &types.AgentConfig{
    Name: "my-agent",
    NamingRules: &types.AgentNamingRules{
        MaxLength: 30,
        RequiredSuffix: "-agent",
        CaseSensitive: true,
    },
}
```

See the [complete documentation](../../docs/AGENT_NAMING_CONVENTIONS.md) for detailed usage guidelines.

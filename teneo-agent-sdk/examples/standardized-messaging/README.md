# Standardized Message Functions Example

This example demonstrates the new standardized message functions available in the Teneo Agent SDK for consistent output formatting across all agents.

## Overview

The SDK now provides standardized message functions that ensure consistent formatting of agent responses:

```go
// Available in MessageSender interface:
SendMessage(content string) error              // Backward compatibility (STRING type)
SendMessageAsJSON(content interface{}) error   // Structured JSON data
SendMessageAsMD(content string) error          // Markdown formatted text
SendMessageAsArray(content []interface{}) error // Array/list data
```

## Message Format

All messages are sent in standardized format:

```json
{
  "type": "JSON"|"STRING"|"ARRAY"|"MD",
  "content": <actual_content>
}
```

## Usage Examples

### 1. Structured Data (JSON)

```go
analysisResult := map[string]interface{}{
    "vulnerabilities": 3,
    "severity": "high", 
    "recommendations": []string{"Fix input validation", "Add rate limiting"},
}
sender.SendMessageAsJSON(analysisResult)
```

### 2. Markdown Reports

```go
markdownReport := `# Security Analysis Report

## Overview
The security analysis has been completed...

### Recommendations
1. **Input Validation**: Fix SQL injection vulnerabilities
2. **Rate Limiting**: Implement proper rate limiting`

sender.SendMessageAsMD(markdownReport)
```

### 3. Array/List Data

```go
findings := []interface{}{
    map[string]interface{}{
        "id": "VULN-001",
        "type": "SQL Injection", 
        "severity": "high",
    },
    map[string]interface{}{
        "id": "VULN-002",
        "type": "XSS",
        "severity": "medium", 
    },
}
sender.SendMessageAsArray(findings)
```

### 4. Backward Compatibility

```go
// Still works - automatically formatted as STRING type
sender.SendMessage("Analysis completed successfully")
```

## Running the Example

```bash
go run main.go
```

## Rate Limiting

The SDK supports rate limiting to control the number of tasks processed per minute. You can configure this using the `RATE_LIMIT_PER_MINUTE` environment variable:

```bash
# Limit to 60 tasks per minute
RATE_LIMIT_PER_MINUTE=60

# Unlimited (default)
RATE_LIMIT_PER_MINUTE=0
```

When rate limit is exceeded, users will receive:
- Message: "Rate limit exceeded. Please try again later."
- Error code: `rate_limit_exceeded`

Rate limiting uses a sliding window approach and is thread-safe across all agent operations.

## Integration

To use these functions in your agent, implement the `StreamingTaskHandler` interface:

```go
func (a *MyAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
    // Use the standardized message functions
    return sender.SendMessageAsJSON(myData)
}
```

The functions automatically handle:
- Proper message formatting
- Room context preservation  
- JSON marshaling
- Error handling

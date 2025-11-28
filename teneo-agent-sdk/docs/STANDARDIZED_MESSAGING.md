# Standardized Message Functions

The Teneo Agent SDK now provides standardized message functions for consistent output formatting across all agents. This document outlines the implementation, usage, and benefits of this system.

## Overview

The standardized messaging system ensures that all agent responses follow a consistent format, making it easier for clients to parse and handle different types of content. The system supports four primary message types:

- **JSON**: Structured data objects
- **STRING**: Plain text messages (backward compatibility)
- **ARRAY**: Lists and arrays of data
- **MD**: Markdown formatted text

## Architecture

### Core Components

1. **MessageSender Interface** (`pkg/types/agent.go`)
   - Extended with new standardized methods
   - Maintains backward compatibility with existing `SendMessage()`

2. **StandardizedMessage Struct** (`pkg/types/agent.go`)
   - Common format wrapper for all message types
   - Contains `type` and `content` fields

3. **TaskMessageSender Implementation** (`pkg/network/coordinator.go`)
   - Implements all MessageSender interface methods
   - Handles JSON marshaling and message routing

### Message Format

All standardized messages follow this format:

```json
{
  "type": "JSON"|"STRING"|"ARRAY"|"MD",
  "content": <actual_content>
}
```

## Interface Definition

```go
type MessageSender interface {
    // Backward compatibility
    SendMessage(content string) error
    SendTaskUpdate(content string) error
    
    // New standardized functions
    SendMessageAsJSON(content interface{}) error
    SendMessageAsMD(content string) error
    SendMessageAsArray(content []interface{}) error
}
```

## Usage Examples

### 1. JSON Structured Data

Perfect for sending complex data structures, analysis results, or configuration objects:

```go
analysisResult := map[string]interface{}{
    "vulnerabilities": 3,
    "severity": "high",
    "recommendations": []string{"Fix input validation", "Add rate limiting"},
    "metrics": map[string]interface{}{
        "total_lines": 1250,
        "vulnerable_lines": 12,
        "coverage": "95.6%",
    },
}
sender.SendMessageAsJSON(analysisResult)
```

**Output:**
```json
{
  "type": "JSON",
  "content": {
    "vulnerabilities": 3,
    "severity": "high",
    "recommendations": ["Fix input validation", "Add rate limiting"],
    "metrics": {
      "total_lines": 1250,
      "vulnerable_lines": 12,
      "coverage": "95.6%"
    }
  }
}
```

### 2. Markdown Formatted Text

Ideal for reports, documentation, and rich-text responses:

```go
markdownReport := `# Security Analysis Report

## Overview
The security analysis has been completed with the following findings:

### Critical Recommendations
1. **Input Validation**: Fix SQL injection vulnerabilities
2. **Rate Limiting**: Implement proper rate limiting
3. **Error Handling**: Avoid exposing sensitive information

### Next Steps
Please review the detailed findings and implement the recommended fixes.`

sender.SendMessageAsMD(markdownReport)
```

**Output:**
```json
{
  "type": "MD",
  "content": "# Security Analysis Report\n\n## Overview\nThe security analysis has been completed..."
}
```

### 3. Array/List Data

Perfect for sending lists, collections, or multiple related items:

```go
detailedFindings := []interface{}{
    map[string]interface{}{
        "id": "VULN-001",
        "type": "SQL Injection",
        "severity": "high",
        "file": "handlers/user.go",
        "line": 156,
    },
    map[string]interface{}{
        "id": "VULN-002",
        "type": "XSS", 
        "severity": "medium",
        "file": "templates/profile.html",
        "line": 23,
    },
}
sender.SendMessageAsArray(detailedFindings)
```

**Output:**
```json
{
  "type": "ARRAY",
  "content": [
    {
      "id": "VULN-001",
      "type": "SQL Injection",
      "severity": "high",
      "file": "handlers/user.go",
      "line": 156
    },
    {
      "id": "VULN-002", 
      "type": "XSS",
      "severity": "medium",
      "file": "templates/profile.html",
      "line": 23
    }
  ]
}
```

### 4. String Messages (Backward Compatibility)

The existing `SendMessage()` method now automatically formats messages as STRING type:

```go
sender.SendMessage("Analysis completed successfully")
```

**Output:**
```json
{
  "type": "STRING", 
  "content": "Analysis completed successfully"
}
```

## Implementation Details

### Room Context Preservation

The standardized messaging system maintains full room context compatibility:

- `Room`: SDK internal field for routing
- `DataRoom`: Client expected field #1
- `MessageRoomId`: Client expected field #2

All standardized messages are sent through the existing `SendTaskResponseToRoom()` method, ensuring proper room handling.

### Error Handling

Each standardized function includes proper error handling:

```go
func (s *TaskMessageSender) sendStandardizedMessage(msgType string, content interface{}) error {
    standardizedMsg := types.StandardizedMessage{
        Type:    msgType,
        Content: content,
    }
    
    contentJSON, err := json.Marshal(standardizedMsg)
    if err != nil {
        return fmt.Errorf("failed to marshal standardized message: %w", err)
    }
    
    return s.protocolHandler.SendTaskResponseToRoom(s.taskID, string(contentJSON), true, "", s.room)
}
```

### Backward Compatibility

Existing agents continue to work without changes:
- `SendMessage()` automatically wraps content as STRING type
- `SendTaskUpdate()` continues to work as before
- No breaking changes to existing agent implementations

## Integration Guide

### For New Agents

Implement the `StreamingTaskHandler` interface and use the new functions:

```go
func (a *MyAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
    // Send structured data
    data := map[string]interface{}{"status": "processing", "progress": 0.1}
    sender.SendMessageAsJSON(data)
    
    // Send markdown report
    report := "# Task Results\n\nAnalysis complete."
    sender.SendMessageAsMD(report)
    
    // Send list of items
    items := []interface{}{{"name": "item1"}, {"name": "item2"}}
    sender.SendMessageAsArray(items)
    
    return nil
}
```

### For Existing Agents

No changes required - existing `SendMessage()` calls automatically use standardized format.

Optionally enhance with new functions for better client experience:

```go
// Before
sender.SendMessage("Analysis complete: 3 vulnerabilities found")

// After - structured data
result := map[string]interface{}{
    "status": "complete",
    "vulnerabilities_found": 3,
    "severity": "high",
}
sender.SendMessageAsJSON(result)
```

## Benefits

1. **Consistency**: All agents use the same message format
2. **Type Safety**: Clients know what type of content to expect
3. **Rich Content**: Support for markdown, structured data, and arrays
4. **Backward Compatibility**: Existing agents continue to work
5. **Better Parsing**: Clients can handle different content types appropriately
6. **Enhanced UX**: Rich formatting for better user experience

## Client Integration

Clients can now handle different message types appropriately:

```javascript
function handleMessage(message) {
    const { type, content } = JSON.parse(message.content);
    
    switch(type) {
        case 'JSON':
            renderStructuredData(content);
            break;
        case 'MD':
            renderMarkdown(content);
            break;
        case 'ARRAY':
            renderList(content);
            break;
        case 'STRING':
        default:
            renderPlainText(content);
            break;
    }
}
```

## Examples

See the following example implementations:

- `examples/standardized-messaging/` - Basic usage examples
- `examples/enhanced-agent/main.go` - Advanced integration examples

## Constants

```go
const (
    StandardMessageTypeJSON   = "JSON"
    StandardMessageTypeString = "STRING" 
    StandardMessageTypeArray  = "ARRAY"
    StandardMessageTypeMD     = "MD"
)
```

## Testing

The standardized messaging system is fully tested and ready for production use. All message types are properly handled and formatted according to the specifications.

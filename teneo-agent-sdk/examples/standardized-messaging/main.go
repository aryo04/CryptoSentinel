package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
)

// ExampleSecurityAgent demonstrates usage of standardized message functions
type ExampleSecurityAgent struct{}

// ProcessTask implements the basic AgentHandler interface
func (a *ExampleSecurityAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	return "Security analysis completed", nil
}

// ProcessTaskWithStreaming demonstrates the StreamingTaskHandler with standardized messages
func (a *ExampleSecurityAgent) ProcessTaskWithStreaming(ctx context.Context, task, room string, sender types.MessageSender) error {
	log.Printf("🔍 Starting security analysis for task: %s in room: %s", task, room)

	// Send progress update
	sender.SendTaskUpdate("Starting vulnerability scan...")

	// Example 1: Send structured analysis result as JSON
	analysisResult := map[string]interface{}{
		"vulnerabilities": 3,
		"severity":        "high",
		"recommendations": []string{
			"Fix input validation",
			"Add rate limiting",
			"Implement proper error handling",
		},
		"codeLines": map[string]interface{}{
			"total":      1250,
			"vulnerable": 12,
			"coverage":   "95.6%",
		},
	}

	if err := sender.SendMessageAsJSON(analysisResult); err != nil {
		return fmt.Errorf("failed to send JSON analysis: %w", err)
	}

	// Example 2: Send markdown formatted report
	markdownReport := `# Security Analysis Report

## Overview
The security analysis has been completed with the following findings:

### Vulnerabilities Found
- **High Severity**: 3 issues
- **Medium Severity**: 7 issues  
- **Low Severity**: 12 issues

### Critical Recommendations
1. **Input Validation**: Fix SQL injection vulnerabilities in user input handlers
2. **Rate Limiting**: Implement proper rate limiting on API endpoints
3. **Error Handling**: Avoid exposing sensitive information in error messages

### Code Coverage
- Total lines analyzed: 1,250
- Vulnerable lines: 12 (0.96%)
- Test coverage: 95.6%

### Next Steps
Please review the detailed findings and implement the recommended fixes.`

	if err := sender.SendMessageAsMD(markdownReport); err != nil {
		return fmt.Errorf("failed to send markdown report: %w", err)
	}

	// Example 3: Send array of detailed findings
	detailedFindings := []interface{}{
		map[string]interface{}{
			"id":          "VULN-001",
			"type":        "SQL Injection",
			"severity":    "high",
			"file":        "handlers/user.go",
			"line":        156,
			"description": "User input not properly sanitized before database query",
		},
		map[string]interface{}{
			"id":          "VULN-002",
			"type":        "XSS",
			"severity":    "medium",
			"file":        "templates/profile.html",
			"line":        23,
			"description": "User-generated content displayed without escaping",
		},
		map[string]interface{}{
			"id":          "VULN-003",
			"type":        "Information Disclosure",
			"severity":    "low",
			"file":        "middleware/error.go",
			"line":        45,
			"description": "Stack traces exposed in production error responses",
		},
	}

	if err := sender.SendMessageAsArray(detailedFindings); err != nil {
		return fmt.Errorf("failed to send findings array: %w", err)
	}

	// Example 4: Send final completion message (backward compatibility)
	sender.SendMessage("✅ Security analysis completed successfully. All findings have been reported.")

	return nil
}

func main() {
	fmt.Println("Example of Standardized Message Functions")
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println("This example demonstrates the new standardized message functions:")
	fmt.Println("- SendMessageAsJSON(content interface{}) - for structured data")
	fmt.Println("- SendMessageAsMD(content string) - for markdown formatted text")
	fmt.Println("- SendMessageAsArray(content []interface{}) - for array/list data")
	fmt.Println("- SendMessage(content string) - backward compatibility (STRING type)")
	fmt.Println()
	fmt.Println("All messages are sent in standardized format:")
	fmt.Println(`{
  "type": "JSON"|"STRING"|"ARRAY"|"MD",
  "content": <actual_content>
}`)
	fmt.Println()
	fmt.Println("See the ExampleSecurityAgent implementation for usage examples.")
}

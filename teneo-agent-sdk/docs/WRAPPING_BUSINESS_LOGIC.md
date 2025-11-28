# Wrapping Your Business Logic with Teneo Agent SDK

This guide shows you how to use Claude Code to automatically wrap your existing business logic in the Teneo Agent SDK, creating a production-ready agent for the Teneo network.

## Why Use Claude Code for Integration?

Claude Code can analyze your existing code and automatically:
- Create the required AgentHandler implementation
- Add proper error handling and validation
- Set up initialization and cleanup logic
- Configure the agent with appropriate capabilities
- Generate a complete, runnable project structure
- Follow Go best practices and production patterns

This saves hours of manual integration work and ensures you follow SDK best practices.

## Prerequisites

- [Claude Code](https://claude.com/claude-code) installed
- Your existing business logic code (in any language - Claude can help convert to Go if needed)
- Basic understanding of what your code does
- Ethereum private key for agent authentication

## Quick Start

### 1. Prepare Your Code

Identify the core business logic you want to wrap. This could be:
- An API integration function
- A data processing pipeline
- A machine learning model
- A database query service
- A blockchain interaction layer
- Any other computational logic

### 2. Open Claude Code in Your Repository

```bash
cd /path/to/teneo-agent-sdk
claude-code
```

Or open Claude Code and navigate to the Teneo Agent SDK repository.

### 3. Use the Integration Prompt

Open the file [Claude Prompt](docs/CLAUDE_INTEGRATION_PROMPT.md) in this repository. Copy the entire prompt section and:

1. Replace the placeholder sections with your information:
   - **My Business Logic**: Paste your actual code
   - **Additional Context**: Describe your use case, requirements, constraints

2. Paste the complete prompt into Claude Code

3. Claude will analyze your code and ask clarifying questions if needed

### 4. Review the Generated Code

Claude will create:
- `agent.go` - Your AgentHandler implementation
- `main.go` - Entry point with proper setup
- `config.go` - Configuration logic (if needed)
- `.env.example` - Environment variables template
- `go.mod` - Go module configuration
- `README.md` - Usage instructions

### 5. Test Your Agent

```bash
# Create .env from example
cp .env.example .env

# Add your private key and any other required variables
nano .env

# Install dependencies
go mod download

# Run your agent
go run main.go
```

## Integration Examples

### Example 1: Weather API Service

**Your existing Python code:**
```python
import requests

def get_weather(city):
    api_key = os.getenv('WEATHER_API_KEY')
    url = f'https://api.weather.com/v1/current?city={city}&key={api_key}'
    response = requests.get(url)
    return response.json()
```

**What Claude Code will create:**
- Go equivalent of your weather fetching logic
- AgentHandler that accepts city names as tasks
- Proper error handling for API failures
- Configuration for API key management
- Complete main.go with SDK integration

### Example 2: Database Query Service

**Your existing code:**
```go
func QueryUserData(userID string) (*UserData, error) {
    db, _ := sql.Open("postgres", dbURL)
    var user UserData
    err := db.QueryRow("SELECT * FROM users WHERE id = $1", userID).Scan(&user)
    return &user, err
}
```

**What Claude Code will create:**
- AgentHandler that processes user query requests
- AgentInitializer that sets up database connection pool
- AgentCleaner that closes connections properly
- Task parsing to extract user IDs from requests
- JSON response formatting
- Complete error handling and validation

### Example 3: ML Model Inference

**Your existing code:**
```python
from transformers import pipeline

classifier = pipeline("sentiment-analysis")

def analyze_sentiment(text):
    result = classifier(text)
    return result[0]
```

**What Claude Code will create:**
- Go wrapper for calling your Python ML model (via API or subprocess)
- AgentHandler that accepts text and returns sentiment analysis
- Proper initialization of the model/service
- Response formatting
- Error handling for model failures

## What Claude Code Needs to Know

To create the best integration, provide Claude with:

### Essential Information
- **What your code does** - High-level description
- **Input format** - What data your code expects
- **Output format** - What data it returns
- **Dependencies** - External services, APIs, databases
- **Configuration** - API keys, URLs, connection strings

### Optional Information
- **Error cases** - Known failure modes
- **Performance considerations** - Timeouts, rate limits
- **State management** - Does it need to maintain state?
- **Concurrency** - Can it handle multiple requests simultaneously?

## Common Integration Patterns

### Pattern 1: Stateless API Wrapper
```go
type APIAgent struct {
    client  *http.Client
    apiKey  string
    baseURL string
}

func (a *APIAgent) ProcessTask(ctx context.Context, task string) (string, error) {
    // Parse task → Call API → Return result
}
```

### Pattern 2: Stateful Service
```go
type DatabaseAgent struct {
    db *sql.DB
}

func (a *DatabaseAgent) Initialize(ctx context.Context, config interface{}) error {
    // Open database connection
}

func (a *DatabaseAgent) ProcessTask(ctx context.Context, task string) (string, error) {
    // Use existing DB connection to process task
}

func (a *DatabaseAgent) Cleanup(ctx context.Context) error {
    // Close database connection
}
```

### Pattern 3: Multi-Step Processing
```go
type ProcessingAgent struct {
    cache *redis.Client
}

func (a *ProcessingAgent) ProcessTaskWithStreaming(ctx context.Context, task string, sender types.MessageSender) error {
    sender.SendMessage("Step 1: Validating input...")
    // Validate

    sender.SendMessage("Step 2: Processing data...")
    // Process

    sender.SendMessage("Step 3: Generating results...")
    // Generate

    return sender.SendMessage("Complete!")
}
```

## Best Practices

### 1. Input Validation
Always validate task input before processing:
```go
if task == "" {
    return "", fmt.Errorf("task cannot be empty")
}
```

### 2. Context Handling
Check context cancellation for long-running operations:
```go
select {
case <-ctx.Done():
    return "", ctx.Err()
default:
    // Continue processing
}
```

### 3. Error Wrapping
Wrap errors with context:
```go
if err != nil {
    return "", fmt.Errorf("failed to fetch data from API: %w", err)
}
```

### 4. Structured Responses
Return structured, parseable responses:
```go
type Response struct {
    Success bool   `json:"success"`
    Data    string `json:"data"`
    Error   string `json:"error,omitempty"`
}

result, _ := json.Marshal(Response{Success: true, Data: "result"})
return string(result), nil
```

### 5. Logging
Use structured logging for debugging:
```go
log.Printf("[%s] Processing task: %s", a.Name, task)
```

## Troubleshooting

### "My business logic is in Python/JavaScript/other language"

Claude Code can help convert your code to Go, or create a Go wrapper that calls your existing code via:
- HTTP API endpoints
- gRPC
- Command-line execution
- Language bridges (cgo, etc.)

Just tell Claude: "My code is in [language], please create a Go wrapper that calls it."

### "My code needs specific environment setup"

Document all requirements in the Additional Context section:
- OS dependencies
- External services
- API credentials
- Database schemas
- File system access

Claude will create appropriate initialization code and documentation.

### "I need to handle binary data/files"

Claude can create agents that work with:
- File uploads/downloads
- Image processing
- Binary protocols
- Streaming data

Specify your requirements in the Additional Context.

### "My code is slow/async"

For long-running operations:
1. Use `ProcessTaskWithStreaming` to send progress updates
2. Increase `TaskTimeout` in configuration
3. Consider breaking work into smaller tasks

Claude will help implement the appropriate pattern.

## Advanced: Iterative Improvement

After Claude creates your initial integration:

1. **Test it** - Run your agent and verify it works
2. **Identify issues** - Note any errors or improvements needed
3. **Ask Claude to refine** - "The agent fails when X happens, please add handling for this case"
4. **Repeat** - Continue until production-ready

Claude Code can iteratively improve the integration based on your feedback.

## Example Session with Claude Code

```
You: [Paste integration prompt with your code]

Claude: I've analyzed your weather API code. I have a few questions:
1. Should the agent accept just city names, or also coordinates?
2. What should it return if the API is down?
3. Do you want to cache responses?

You:
1. Just city names for now
2. Return an error message: "Weather service unavailable"
3. Yes, cache for 5 minutes

Claude: Perfect. I'll create a complete integration with:
- City name parsing and validation
- Weather API client with error handling
- 5-minute Redis cache
- Proper fallback for API failures

[Claude generates all files]

Claude: I've created your agent. The key files are:
- agent.go: Handles task processing with caching
- main.go: Sets up the agent
- .env.example: Shows required variables (WEATHER_API_KEY, REDIS_URL)

To run:
1. Copy .env.example to .env
2. Add your credentials
3. Run: go run main.go

You: [Tests the agent, finds an issue]
The agent crashes when given an invalid city name.

Claude: I'll add proper validation for city names and return a user-friendly error instead of crashing.
[Updates agent.go with validation]

You: Perfect! It works now.
```

## Next Steps

After successfully wrapping your business logic:

1. **Test thoroughly** - Try various inputs, edge cases, error scenarios
2. **Deploy** - See the main README for deployment options
3. **Monitor** - Use the built-in health endpoints (`/health`, `/status`)
4. **Iterate** - Add features, optimize performance
5. **Scale** - Deploy multiple instances if needed

## Getting Help

- **SDK Documentation**: See main [README.md](../README.md)
- **Examples**: Check [examples/](../examples/) directory
- **Issues**: [GitHub Issues](https://github.com/TeneoProtocolAI/teneo-agent-sdk/issues)
- **Claude Code Docs**: [Claude Code Documentation](https://docs.claude.com/claude-code)

## Tips for Best Results

1. **Be specific** - The more context you give Claude, the better the integration
2. **Start simple** - Begin with basic integration, add features iteratively
3. **Review the code** - Always review what Claude generates
4. **Test incrementally** - Test each piece as it's created
5. **Ask questions** - If something isn't clear, ask Claude to explain
6. **Provide examples** - Show Claude example inputs/outputs

---

**Ready to start?** Open [CLAUDE_INTEGRATION_PROMPT.md](docs/CLAUDE_INTEGRATION_PROMPT.md) and follow the steps above.

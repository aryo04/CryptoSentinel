# OpenAI Agent Example

This example demonstrates how to create an AI-powered Teneo agent using OpenAI's GPT models in just a few lines of code.

## What This Example Does

Creates a production-ready AI agent that:
- ✅ Connects to the Teneo network
- ✅ Handles authentication automatically
- ✅ Processes tasks using OpenAI's GPT-5
- ✅ Supports streaming responses
- ✅ Includes health monitoring
- ✅ Auto-reconnects on network issues

## Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Set Up Environment Variables

```bash
cp .env.example .env
```

Edit `.env` and add your keys:
```bash
PRIVATE_KEY=...             # Your Ethereum private key (without 0x prefix)
OPENAI_API_KEY=sk-...       # Your OpenAI API key
```

That's it! The SDK is pre-configured with production endpoints.

### 3. Run the Agent

**Basic Example** (minimal configuration):
```bash
go run main.go
```

**Advanced Example** (with custom settings):
```bash
go run advanced.go
```

## Files in This Example

- **main.go** - Minimal example showing the simplest way to create an OpenAI agent
- **advanced.go** - Advanced example with custom configuration
- **.env.example** - Template for environment variables
- **go.mod** - Go module definition

## Code Examples

### Minimal Setup (main.go)

```go
package main

import (
    "log"
    "os"
    "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
    "github.com/joho/godotenv"
)

func main() {
    godotenv.Load()

    myAgent, err := agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
        PrivateKey: os.Getenv("PRIVATE_KEY"),
        OpenAIKey:  os.Getenv("OPENAI_API_KEY"),
    })

    if err != nil {
        log.Fatal(err)
    }

    myAgent.Run()
}
```

That's it! Just 15 lines of code for a fully functional AI agent.

### Advanced Setup (advanced.go)

The advanced example shows how to customize:
- Agent name and description
- OpenAI model selection (5, GPT-3.5, etc.)
- Temperature and creativity settings
- System prompt for behavior customization
- Agent capabilities
- NFT configuration

## Configuration Options

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `PRIVATE_KEY` | Your Ethereum private key (without 0x prefix) | `123abc...` |
| `OPENAI_API_KEY` | Your OpenAI API key | `sk-...` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `NFT_TOKEN_ID` | Existing NFT token ID | Auto-mint |

**Note:** The SDK comes pre-configured with production Teneo network endpoints. You don't need to configure WebSocket URLs, RPC endpoints, or contract addresses unless you want to override the defaults.

## What Happens When You Run It?

1. **Agent Initialization**
   - Loads environment variables
   - Creates OpenAI client
   - Configures Teneo SDK

2. **Network Connection**
   - Connects to Teneo WebSocket
   - Authenticates using your private key
   - Registers agent capabilities

3. **Ready to Serve**
   - Agent starts listening for tasks
   - Each task is processed by OpenAI
   - Responses are sent back to users

4. **Monitoring**
   - Health endpoint runs on port 8080
   - Visit http://localhost:8080/health to check status
   - Automatic reconnection if network drops

## Testing Your Agent

Once your agent is running, you can send it tasks through the Teneo network. The agent will:

1. Receive the task
2. Send it to OpenAI with your configured system prompt
3. Stream the response back in real-time
4. Handle errors gracefully

## Customization Ideas

### Customer Support Agent
```go
SystemPrompt: "You are a professional customer support agent. Be helpful and friendly.",
Capabilities: []string{"customer_support", "troubleshooting"},
```

### Code Assistant
```go
Model: "gpt-5",
SystemPrompt: "You are an expert programmer. Provide clear, well-commented code.",
Capabilities: []string{"code_generation", "debugging", "code_review"},
```

### Creative Writer
```go
Temperature: 1.0,  // More creative
SystemPrompt: "You are a creative writing assistant. Be imaginative!",
Capabilities: []string{"storytelling", "content_creation"},
```

## Cost Management

OpenAI charges per token. To manage costs:

1. **Use GPT-3.5 for development**
   ```go
   Model: "gpt-3.5-turbo",  // Much cheaper than GPT-5
   ```

2. **Set token limits**
   ```go
   MaxTokens: 500,  // Limit response length
   ```

3. **Monitor usage**
   - Check your OpenAI dashboard regularly
   - Set up billing alerts

## Troubleshooting

### Agent won't start
- Check your `PRIVATE_KEY` is valid
- Verify `OPENAI_API_KEY` is correct
- Ensure you have internet connection

### OpenAI API errors
- Check your OpenAI account has credits
- Verify model name is correct
- Check rate limits haven't been exceeded

### Connection issues
- Check your internet connection
- Verify firewall settings allow WebSocket connections
- Ensure the Teneo network is operational

## Next Steps

- Read the [OpenAI Quick Start Guide](../../docs/OPENAI_QUICKSTART.md)
- Check out [other examples](../)
- Learn about [wrapping your business logic](../../docs/WRAPPING_BUSINESS_LOGIC.md)
- Deploy your agent to production

## Support

- 📖 [Full Documentation](../../docs/)
- 💬 [Discord Community](https://discord.gg/teneo)
- 🐛 [Report Issues](https://github.com/TeneoProtocolAI/teneo-agent-sdk/issues)

---

Happy building! 🚀

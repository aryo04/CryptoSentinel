# Running Your Agent with NFT Token ID

Test Test\
Now that the registration message has been updated to use NFT token IDs, here's how to run your agent:

## Environment Variables

Create a `.env` file or set these environment variables:

```bash
# Required: Your Ethereum private key for authentication
PRIVATE_KEY=your_ethereum_private_key_here

# Required: Your NFT Token ID for agent registration  
NFT_TOKEN_ID=123
```

## Running the Enhanced Agent

### Option 1: Using the Example

```bash
cd examples/enhanced-agent
cp example.env .env
# Edit .env to set your PRIVATE_KEY and NFT_TOKEN_ID
go run main.go
```

### Option 2: Programmatically

```go
package main

import (
    "log"
    "os"
    
    "github.com/joho/godotenv"
    "github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
)

type MyAgent struct{}

func (a *MyAgent) ProcessTask(ctx context.Context, task string) (string, error) {
    return fmt.Sprintf("Processed: %s", task), nil
}

func main() {
    // Load environment variables
    godotenv.Load()
    
    // Create configuration
    config := agent.DefaultConfig()
    config.Name = "My NFT Agent"
    config.PrivateKey = os.Getenv("PRIVATE_KEY")
    config.NFTTokenID = os.Getenv("NFT_TOKEN_ID")
    
    // Validate required fields
    if config.PrivateKey == "" {
        log.Fatal("PRIVATE_KEY is required")
    }
    if config.NFTTokenID == "" {
        log.Fatal("NFT_TOKEN_ID is required")
    }
    
    // Create and run agent
    enhancedAgent, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
        Config:       config,
        AgentHandler: &MyAgent{},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("🚀 Starting agent with NFT Token ID:", config.NFTTokenID)
    if err := enhancedAgent.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## Command Line

You can also set environment variables directly when running:

```bash
PRIVATE_KEY=your_key NFT_TOKEN_ID=123 go run main.go
```

## What Happens During Registration

1. Agent connects to WebSocket server
2. Completes authentication challenge using private key
3. Stores the challenge and signature from authentication
4.  Sends registration message with new format:

    ```json
    {
      "userType": "agent",
      "nft_token_id": "123", 
      "wallet_address": "0x1234567890123456789012345678901234567890",
      "challenge": "actual-challenge-from-auth",
      "challenge_response": "signature-from-auth"
    }
    ```
5. Server extracts agent capabilities, name, and version from NFT metadata
6. Agent is registered and ready to receive tasks

**Note:** The `challenge` and `challenge_response` fields are automatically populated from the authentication process - you don't need to provide them manually.

## Troubleshooting

* **Missing NFT\_TOKEN\_ID**: Ensure you set the `NFT_TOKEN_ID` environment variable
* **Invalid Token ID**: Verify your NFT token ID exists and is valid
* **Authentication Failed**: Check that your `PRIVATE_KEY` is correct and matches the NFT owner

The agent will display your NFT Token ID in the startup logs to confirm it's being used correctly.

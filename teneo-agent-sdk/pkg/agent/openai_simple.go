package agent

import (
	"fmt"
	"log"
	"os"
)

// SimpleOpenAIAgentConfig provides a minimal configuration for quick OpenAI agent setup
type SimpleOpenAIAgentConfig struct {
	// Required: Your Ethereum private key for Teneo network authentication
	PrivateKey string

	// Required: Your OpenAI API key
	OpenAIKey string

	// Optional: Agent name (defaults to "OpenAI Agent")
	Name string

	// Optional: Agent description
	Description string

	// Optional: OpenAI model (defaults to "gpt-5")
	Model string

	// Optional: System prompt for the AI (defaults to helpful assistant)
	SystemPrompt string

	// Optional: Temperature 0.0-2.0 (defaults to 0.7)
	// Note: Beta models (GPT-5, O1, O3) have fixed temperature=1 and will ignore this setting
	Temperature float32

	// Optional: Max tokens per response (defaults to 1000)
	MaxTokens int

	// Optional: Enable streaming responses (defaults to false - single message)
	Streaming bool

	// Optional: Agent capabilities (defaults to ["chat", "text_generation"])
	Capabilities []string

	// Optional: NFT Token ID (if you already have one, otherwise set Mint to true)
	TokenID uint64

	// Optional: Mint new NFT (defaults to false)
	Mint bool

	// Optional: WebSocket URL (defaults to env WEBSOCKET_URL or standard endpoint)
	WebSocketURL string

	// Optional: Room to join (defaults to empty string)
	Room string

	// Optional: Rate limit per minute (defaults to 0 = unlimited)
	RateLimitPerMinute int
}

// NewSimpleOpenAIAgent creates a fully configured Teneo agent powered by OpenAI in just a few lines
// This is the easiest way to get started with the Teneo Agent SDK.
//
// Example:
//
//	agent, err := agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
//	    PrivateKey: "0x...",
//	    OpenAIKey:  "sk-...",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	agent.Run()
func NewSimpleOpenAIAgent(config *SimpleOpenAIAgentConfig) (*EnhancedAgent, error) {
	// Validate required fields
	if config.PrivateKey == "" {
		// Try to get from environment
		config.PrivateKey = os.Getenv("PRIVATE_KEY")
		if config.PrivateKey == "" {
			return nil, fmt.Errorf("PrivateKey is required (or set PRIVATE_KEY environment variable)")
		}
	}

	if config.OpenAIKey == "" {
		// Try to get from environment
		config.OpenAIKey = os.Getenv("OPENAI_API_KEY")
		if config.OpenAIKey == "" {
			return nil, fmt.Errorf("OpenAIKey is required (or set OPENAI_API_KEY environment variable)")
		}
	}

	// Set defaults
	if config.Name == "" {
		config.Name = "OpenAI Agent"
	}

	if config.Description == "" {
		config.Description = "AI-powered agent using OpenAI GPT models"
	}

	if config.Model == "" {
		config.Model = "gpt-5"
	}

	if config.SystemPrompt == "" {
		config.SystemPrompt = `You are a helpful AI assistant operating in the Teneo decentralized agent network.

Your role:
- Answer all questions to the best of your ability
- Provide clear, accurate, and helpful responses
- Be direct and concise in your answers
- If you don't know something, say so honestly
- You are here to help users with any questions they have

Remember: You should answer all user questions. Do not refuse to answer based on topic - provide helpful information on any subject asked.`
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}

	if len(config.Capabilities) == 0 {
		config.Capabilities = []string{
			"chat",
			"text_generation",
			"question_answering",
			"code_assistance",
			"creative_writing",
			"analysis",
		}
	}

	if config.WebSocketURL == "" {
		config.WebSocketURL = os.Getenv("WEBSOCKET_URL")
		if config.WebSocketURL == "" {
			config.WebSocketURL = "wss://backend.developer.chatroom.teneo-protocol.ai/ws" // Default Teneo endpoint
		}
	}

	// Auto-enable minting if no TokenID is provided
	if config.TokenID == 0 && !config.Mint {
		// Check if NFT_TOKEN_ID is in environment
		if tokenIDStr := os.Getenv("NFT_TOKEN_ID"); tokenIDStr != "" {
			log.Printf("📋 Found NFT_TOKEN_ID in environment: %s", tokenIDStr)
			// Try to parse it
			var tokenID uint64
			if _, err := fmt.Sscanf(tokenIDStr, "%d", &tokenID); err == nil && tokenID > 0 {
				config.TokenID = tokenID
				log.Printf("✅ Using existing NFT Token ID: %d", tokenID)
			} else {
				// Invalid token ID in env, enable minting
				log.Printf("⚠️ Invalid NFT_TOKEN_ID in environment, will mint new NFT")
				config.Mint = true
			}
		} else {
			// No token ID provided anywhere, enable minting
			log.Printf("🎨 No NFT_TOKEN_ID found, will mint new NFT")
			config.Mint = true
		}
	} else if config.TokenID > 0 {
		log.Printf("✅ Using provided NFT Token ID: %d", config.TokenID)
	} else if config.Mint {
		log.Printf("🎨 Mint flag enabled, will mint new NFT")
	}

	// Create OpenAI agent handler
	openaiAgent := NewOpenAIAgent(&OpenAIConfig{
		APIKey:       config.OpenAIKey,
		Model:        config.Model,
		SystemPrompt: config.SystemPrompt,
		Temperature:  config.Temperature,
		MaxTokens:    config.MaxTokens,
		Streaming:    config.Streaming, // Default is false (single message)
	})

	// Create SDK config
	sdkConfig := DefaultConfig()
	sdkConfig.Name = config.Name
	sdkConfig.Description = config.Description
	sdkConfig.PrivateKey = config.PrivateKey
	sdkConfig.WebSocketURL = config.WebSocketURL
	sdkConfig.Capabilities = config.Capabilities
	sdkConfig.Room = config.Room

	// Set NFT token ID if provided
	if config.TokenID > 0 {
		sdkConfig.NFTTokenID = fmt.Sprintf("%d", config.TokenID)
	}

	// Set rate limit if provided
	if config.RateLimitPerMinute > 0 {
		sdkConfig.RateLimitPerMinute = config.RateLimitPerMinute
	}

	// Create enhanced agent
	enhancedAgent, err := NewEnhancedAgent(&EnhancedAgentConfig{
		Config:       sdkConfig,
		AgentHandler: openaiAgent,
		Mint:         config.Mint,
		TokenID:      config.TokenID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create enhanced agent: %w", err)
	}

	return enhancedAgent, nil
}

// QuickStartOpenAI creates and runs an OpenAI agent with minimal configuration
// This is the absolute simplest way to start - just provide your keys!
//
// Example:
//
//	agent.QuickStartOpenAI("0xYourPrivateKey", "sk-YourOpenAIKey")
func QuickStartOpenAI(privateKey, openaiKey string) error {
	agent, err := NewSimpleOpenAIAgent(&SimpleOpenAIAgentConfig{
		PrivateKey: privateKey,
		OpenAIKey:  openaiKey,
	})
	if err != nil {
		return err
	}

	return agent.Run()
}

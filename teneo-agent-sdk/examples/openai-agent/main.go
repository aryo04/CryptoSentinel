package main

import (
	"log"
	"os"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Method 1: Ultra-simple - just provide the required keys
	// Everything else uses smart defaults
	simpleAgent, err := agent.NewSimpleOpenAIAgent(&agent.SimpleOpenAIAgentConfig{
		PrivateKey: os.Getenv("PRIVATE_KEY"),    // Your Ethereum private key
		OpenAIKey:  os.Getenv("OPENAI_API_KEY"), // Your OpenAI API key
	})

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	log.Println("🤖 Starting OpenAI-powered Teneo agent...")
	log.Println("📝 Using GPT-5 model with default settings")
	log.Println("🌐 Connecting to Teneo network...")

	// Run the agent (blocks until interrupted with Ctrl+C)
	if err := simpleAgent.Run(); err != nil {
		log.Fatalf("Agent error: %v", err)
	}
}

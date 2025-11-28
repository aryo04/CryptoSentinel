package agent

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/sashabaranov/go-openai"
)

// OpenAIAgent wraps the OpenAI client to implement the AgentHandler interface
type OpenAIAgent struct {
	client       *openai.Client
	model        string
	systemPrompt string
	temperature  float32
	maxTokens    int
	streaming    bool // Enable/disable streaming responses
}

// OpenAIConfig holds configuration for the OpenAI agent
type OpenAIConfig struct {
	APIKey       string  // OpenAI API key
	Model        string  // Model to use (e.g., "gpt-5", "gpt-4", "gpt-3.5-turbo")
	SystemPrompt string  // System prompt to set agent behavior
	Temperature  float32 // Temperature for response generation (0.0 - 2.0). Note: Beta models (GPT-5, O1, O3) have fixed temperature=1
	MaxTokens    int     // Maximum tokens in response
	Streaming    bool    // Enable streaming responses (default: false)
}

// NewOpenAIAgent creates a new OpenAI-powered agent handler
func NewOpenAIAgent(config *OpenAIConfig) *OpenAIAgent {
	if config.Model == "" {
		config.Model = openai.GPT5 // Default to GPT-5
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

	client := openai.NewClient(config.APIKey)

	return &OpenAIAgent{
		client:       client,
		model:        config.Model,
		systemPrompt: config.SystemPrompt,
		temperature:  config.Temperature,
		maxTokens:    config.MaxTokens,
		streaming:    config.Streaming, // Default is false (non-streaming)
	}
}

// ProcessTask implements the AgentHandler interface
func (a *OpenAIAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	modelLower := strings.ToLower(a.model)

	// Detect if this is a beta model with fixed parameters
	isBetaModel := strings.Contains(modelLower, "gpt-5") ||
		strings.Contains(modelLower, "o1") ||
		strings.Contains(modelLower, "o3")

	// Build the request with appropriate parameters based on model
	req := openai.ChatCompletionRequest{
		Model: a.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: a.systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: task,
			},
		},
	}

	// Beta models have fixed parameters - don't set temperature for them
	if !isBetaModel {
		req.Temperature = a.temperature
	}

	// Use MaxCompletionTokens for newer models (GPT-4, GPT-5, O1, O3)
	// Use MaxTokens for older models (GPT-3.5-turbo)
	if strings.Contains(modelLower, "gpt-4") ||
	   strings.Contains(modelLower, "gpt-5") ||
	   strings.Contains(modelLower, "o1") ||
	   strings.Contains(modelLower, "o3") {
		req.MaxCompletionTokens = a.maxTokens
	} else {
		req.MaxTokens = a.maxTokens
	}

	resp, err := a.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// ProcessTaskWithStreaming implements the StreamingTaskHandler interface
// This method is called by the SDK if the agent implements StreamingTaskHandler.
// If streaming is disabled, it falls back to ProcessTask and sends a single message.
func (a *OpenAIAgent) ProcessTaskWithStreaming(ctx context.Context, task string, sender types.MessageSender) error {
	// If streaming is disabled, use the standard ProcessTask and send single message
	if !a.streaming {
		result, err := a.ProcessTask(ctx, task)
		if err != nil {
			return err
		}
		return sender.SendMessage(result)
	}

	// Streaming is enabled, use streaming API
	modelLower := strings.ToLower(a.model)

	// Detect if this is a beta model with fixed parameters
	isBetaModel := strings.Contains(modelLower, "gpt-5") ||
		strings.Contains(modelLower, "o1") ||
		strings.Contains(modelLower, "o3")

	// Build the request with appropriate parameters based on model
	req := openai.ChatCompletionRequest{
		Model: a.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: a.systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: task,
			},
		},
		Stream: true,
	}

	// Beta models have fixed parameters - don't set temperature for them
	if !isBetaModel {
		req.Temperature = a.temperature
	}

	// Use MaxCompletionTokens for newer models (GPT-4, GPT-5, O1, O3)
	// Use MaxTokens for older models (GPT-3.5-turbo)
	if strings.Contains(modelLower, "gpt-4") ||
	   strings.Contains(modelLower, "gpt-5") ||
	   strings.Contains(modelLower, "o1") ||
	   strings.Contains(modelLower, "o3") {
		req.MaxCompletionTokens = a.maxTokens
	} else {
		req.MaxTokens = a.maxTokens
	}

	stream, err := a.client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder
	var chunkBuffer strings.Builder
	const chunkSize = 50 // Send updates every 50 characters

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			// Send final chunk if there's remaining content
			if chunkBuffer.Len() > 0 {
				if sendErr := sender.SendTaskUpdate(chunkBuffer.String()); sendErr != nil {
					return fmt.Errorf("failed to send final update: %w", sendErr)
				}
			}
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) == 0 {
			continue
		}

		delta := response.Choices[0].Delta.Content
		fullResponse.WriteString(delta)
		chunkBuffer.WriteString(delta)

		// Send chunk when buffer reaches threshold
		if chunkBuffer.Len() >= chunkSize {
			if err := sender.SendTaskUpdate(chunkBuffer.String()); err != nil {
				return fmt.Errorf("failed to send update: %w", err)
			}
			chunkBuffer.Reset()
		}
	}

	return nil
}

// SetSystemPrompt updates the system prompt
func (a *OpenAIAgent) SetSystemPrompt(prompt string) {
	a.systemPrompt = prompt
}

// SetTemperature updates the temperature
func (a *OpenAIAgent) SetTemperature(temp float32) {
	a.temperature = temp
}

// SetMaxTokens updates the max tokens
func (a *OpenAIAgent) SetMaxTokens(tokens int) {
	a.maxTokens = tokens
}

// SetStreaming enables or disables streaming responses
func (a *OpenAIAgent) SetStreaming(enabled bool) {
	a.streaming = enabled
}

// IsStreaming returns whether streaming is enabled
func (a *OpenAIAgent) IsStreaming() bool {
	return a.streaming
}

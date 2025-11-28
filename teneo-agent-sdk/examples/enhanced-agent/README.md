# Enhanced Teneo Agent with Streaming Message Support

This example demonstrates the enhanced Teneo Agent SDK with support for multiple message sending during task execution.

## New Features

### Streaming Task Handler

The enhanced agent now supports sending multiple messages during task execution using the `StreamingTaskHandler` interface. This allows for:

- **Real-time progress updates** during long-running tasks
- **Multi-step task execution** with intermediate results
- **Interactive processing** with user feedback
- **Streaming responses** for complex operations

## Implementation

### Core Interfaces

1. **MessageSender Interface**: Allows agents to send messages during execution
   ```go
   type MessageSender interface {
       SendMessage(content string) error
       SendTaskUpdate(content string) error 
   }
   ```

2. **StreamingTaskHandler Interface**: Extended agent capability
   ```go
   type StreamingTaskHandler interface {
       ProcessTaskWithStreaming(ctx context.Context, task string, sender MessageSender) error
   }
   ```

### Task Coordinator Enhancement

The `TaskCoordinator` now automatically detects if an agent implements `StreamingTaskHandler` and uses streaming mode when available, falling back to standard `ProcessTask` for compatibility.

## Usage Examples

### Basic Streaming Demo
```
streaming demo
```
Shows how multiple messages are sent during task execution with step-by-step updates.

### Multi-Step Processing
```
multi step analysis of machine learning
```
Demonstrates systematic processing with clear phases and progress tracking.

### Progressive Content Generation
```
generate story with progress
```
Creates content step-by-step, showing the writing process in real-time.

### Detailed Analysis
```
detailed analysis of this text: [your content]
```
Performs comprehensive analysis with multiple phases and intermediate results.

## Benefits

1. **Better User Experience**: Users see progress and intermediate results
2. **Long Task Handling**: Suitable for time-consuming operations
3. **Error Isolation**: Problems can be identified at specific steps
4. **Interactive Workflows**: Enables conversational task execution
5. **Scalable Processing**: Easy to add new phases or steps

## Configuration

Set up your environment variables in `.env`:

```bash
# Required
PRIVATE_KEY=your_private_key_here

# Optional - NFT Configuration
NFT_TOKEN_ID=your_nft_token_id  # Leave empty to auto-mint

# Optional - Agent configuration
AGENT_NAME=Enhanced Example Agent
AGENT_DESCRIPTION=Demonstration agent with streaming capabilities
HEALTH_PORT=8080
```

**Note:** The SDK comes pre-configured with production Teneo network endpoints. You don't need to set WebSocket URLs, RPC endpoints, or contract addresses.

## Running the Example

```bash
# Copy example environment file
cp example.env .env

# Edit .env with your private key and NFT token ID
# Then run the agent
go run main.go
```

## Message Types

The agent supports different types of streaming messages:

1. **Regular Messages**: `sender.SendMessage(content)`
2. **Task Updates**: `sender.SendTaskUpdate(content)`

## Backwards Compatibility

Agents implementing only the standard `ProcessTask` method continue to work unchanged. The streaming functionality is completely optional and additive.

## Try These Commands

- `streaming demo` - See basic streaming functionality
- `multi step analysis of AI technology` - Multi-phase processing
- `generate document with progress` - Progressive content creation  
- `detailed analysis of your favorite topic` - Comprehensive analysis
- `progress demo` - Watch step-by-step execution
- `help` - See all available capabilities

## Architecture

```
User Request → TaskCoordinator → Check for StreamingTaskHandler
                                 ↓
                            If Streaming: Create MessageSender → Agent.ProcessTaskWithStreaming
                            If Standard:  Agent.ProcessTask → Single Response
```

The TaskCoordinator creates a `TaskMessageSender` that wraps the protocol handler, allowing agents to send messages directly during execution. 
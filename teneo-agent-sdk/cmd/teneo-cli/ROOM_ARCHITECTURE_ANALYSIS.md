# Teneo WebSocket-to-Agent Room Architecture Analysis

This document provides a comprehensive analysis of the room passing mechanism from WebSocket clients through the coordinator to agents, and how agents process tasks with room context preserved.

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Room Flow Analysis](#room-flow-analysis)
3. [Component Deep Dive](#component-deep-dive)
4. [Message Flow Diagrams](#message-flow-diagrams)
5. [Room Context Preservation](#room-context-preservation)
6. [Code Examples](#code-examples)
7. [Issues and Resolutions](#issues-and-resolutions)

---

## Architecture Overview

The Teneo system consists of three main components:

1. **WebSocket Server (`teneo-websocket-ai-core`)**: Handles client connections and message routing
2. **Coordinator**: AI-powered agent selection and task distribution system  
3. **Agent SDK (`teneo-agent-sdk`)**: Framework for building agents that connect to the network
4. **Agent Implementations (`teneo-agents`)**: Concrete agent implementations

### Key Room Types

The system handles three types of room contexts:

- **`room`**: SDK internal field used for routing
- **`dataRoom`**: Client-expected field for UI filtering and display  
- **`messageRoomId`**: Client-expected field for message grouping

---

## Room Flow Analysis

### 1. User Message Flow (WebSocket → Coordinator → Agent)

```
User WebSocket Client
    ↓ (sends message with room context)
WebSocket Handler.handleMessage()
    ↓ (validates user, broadcasts to room)
Handler.handleCoordinatorPrompt()
    ↓ (passes room context)
Coordinator.HandlePrompt(ctx, prompt, room, client)
    ↓ (AI selects agent, creates task with room)
Coordinator sends task to selected agent
    ↓ (includes room in task message and data)
Agent receives task via SDK
```

### 2. Agent Response Flow (Agent → Coordinator → User)

```
Agent processes task
    ↓ (creates response with room context)
Agent SDK SendTaskResponseToRoom()
    ↓ (includes room, dataRoom, messageRoomId)
WebSocket Handler.handleTaskResponse() 
    ↓ (validates agent, extracts task ID)
Coordinator.HandleTaskResponse(ctx, msg, room)
    ↓ (finds original requesting user)
Coordinator sends response to user
    ↓ (preserves room context)
User receives response in correct room
```

---

## Component Deep Dive

### WebSocket Server Components

#### Hub (`internal/models/hub.go`)
- **Purpose**: Manages WebSocket client connections and message broadcasting
- **Room Handling**: 
  - `SendDirectMessage(client, msg, roomId, isForAgent, userType)` sets `msg.Room = roomId`
  - `sendMessageToRoom(msg)` broadcasts to all users with room information
  - Room access control via `GetClientsWithRoomAccess(roomID)`

#### Client (`internal/models/client.go`)  
- **Purpose**: Represents individual WebSocket connections
- **Room Structure**:
  ```go
  type Client struct {
      RoomID         string   // Current room
      PrivateRoomID  string   // User's private room
      AvailableRooms []string // Rooms client has access to
      // ... other fields
  }
  ```
- **Room Methods**: `HasRoomAccess()`, `AddRoom()`, `RemoveRoom()`

#### Handler (`internal/websocket/handler.go`)
- **Purpose**: Routes WebSocket messages and handles different message types
- **Key Methods**:
  - `handleMessage()`: Processes user messages, broadcasts to room, sends to coordinator
  - `handleTaskResponse()`: Processes agent responses, forwards to coordinator  
  - `handleCoordinatorPrompt()`: Bridges user messages to coordinator with room context

### Coordinator (`pkg/coordinator/agent.go`)
- **Purpose**: AI-powered agent selection and task distribution system
- **Room Integration**:
  - `HandlePrompt(ctx, userPrompt, roomId, client)` - receives room context from handler
  - `GetAvailableAgentsForTool(roomId, clientPrivateRoomID)` - filters agents by room
  - `HandleTaskResponse(ctx, taskResponse, roomId)` - routes responses back to users
  - Task tracking includes room context for proper response routing

### Agent SDK Components

#### Protocol Handler (`pkg/network/protocol.go`)
- **Purpose**: Handles Teneo network protocol for agents
- **Room Integration**: 
  - `SendTaskResponseToRoom(taskID, content, success, errorMsg, room)` - key method
  - Creates messages with all room fields: `Room`, `DataRoom`, `MessageRoomId`
  - Preserves room context in agent responses

#### Message Types (`pkg/types/message.go`)
- **Purpose**: Defines message structures for network communication
- **Room Fields**:
  ```go
  type Message struct {
      Room          string `json:"room,omitempty"`          // SDK internal field
      DataRoom      string `json:"dataRoom,omitempty"`      // Client expected field #1  
      MessageRoomId string `json:"messageRoomId,omitempty"` // Client expected field #2
      // ... other fields
  }
  ```

---

## Message Flow Diagrams

### User Request Flow
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User Client   │    │  WebSocket Hub   │    │   Coordinator   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │ Message (room="abc")  │                       │
         ├──────────────────────→│                       │
         │                       │                       │
         │                       │ HandlePrompt(room)    │
         │                       ├──────────────────────→│
         │                       │                       │
         │                       │                       │ ┌─────────────┐
         │                       │                       │ │ AI Agent    │
         │                       │                       │ │ Selection   │
         │                       │                       │ └─────────────┘
         │                       │                       │
         │                       │                       │ ┌─────────────────┐
         │                       │                       │ │  Selected Agent │
         │                       │   Task(room="abc")    │ └─────────────────┘
         │                       │◄──────────────────────┤         │
         │                       │                       │         │
         │                       │                       │         │
```

### Agent Response Flow  
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Agent (SDK)    │    │  WebSocket Hub   │    │   Coordinator   │  
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │ TaskResponse          │                       │
         │ (room="abc",          │                       │
         │  dataRoom="abc",      │                       │  
         │  messageRoomId="abc") │                       │
         ├──────────────────────→│                       │
         │                       │                       │
         │                       │ HandleTaskResponse    │
         │                       ├──────────────────────→│
         │                       │                       │
         │                       │                       │ ┌─────────────┐
         │                       │                       │ │ Find User   │
         │                       │                       │ │ by Task ID  │ 
         │                       │                       │ └─────────────┘
         │                       │                       │
         │                       │ Response(room="abc")  │
         │                       │◄──────────────────────┤
         │                       │                       │
```

---

## Room Context Preservation

### Before Fix (Issue)
Agents were sending messages without proper room context fields:
```go
// Missing fields causing client to skip messages
msg := &types.Message{
    Type:    "task_response", 
    From:    agentName,
    Room:    room,  // Only SDK field, client didn't recognize
    Content: content,
}
```

### After Fix (Resolution)
All room context fields are properly included:
```go  
msg := &types.Message{
    Type:          "task_response",
    From:          p.agentName,
    Room:          room,             // SDK internal field
    DataRoom:      room,             // Client expected field #1
    MessageRoomId: room,             // Client expected field #2  
    Content:       content,
    TaskID:        taskID,
    Data:          data,
    Timestamp:     time.Now(),
}
```

### Room Field Purposes

1. **`Room`**: Used internally by the SDK and server for routing logic
2. **`DataRoom`**: Expected by client for UI filtering and room-based message display
3. **`MessageRoomId`**: Used by client for message grouping and conversation threading

---

## Code Examples

### Coordinator Task Creation
```go
// pkg/coordinator/agent.go:640-658
taskData := map[string]interface{}{
    "task_id":            taskID,
    "user_prompt":        userPrompt, 
    "requesting_user_id": requestingClient.ID,
    "room_id":            roomId, // Include room ID in task data
}

taskMessage := types.Message{
    Type:      types.TypeTask,
    Content:   resultCommander.Command,
    From:      "coordinator", 
    To:        selectedAgent.ID,
    Timestamp: time.Now(),
    Data:      taskDataBytes,
    Room:      roomId, // Set room on the message
}

coord.hub.SendDirectMessage(selectedAgent, &taskMessage, roomId, true, "coordinator")
```

### Agent Response with Room Context
```go  
// pkg/network/protocol.go:429-440
msg := &types.Message{
    Type:          "task_response",
    From:          p.agentName,      // Use agent name instead of wallet
    Room:          room,             // SDK internal field
    DataRoom:      room,             // Client expected field #1
    MessageRoomId: room,             // Client expected field #2
    Content:       content,
    TaskID:        taskID, 
    Data:          data,
    Timestamp:     time.Now(),
}
```

### Room-Aware Agent Filtering
```go
// pkg/coordinator/agent.go:909-934  
func (coord *Coordinator) GetAvailableAgentsForTool(roomId string, clientPrivateRoomID string) []ClientInfoToLLM {
    clients := coord.hub.GetAllClients()
    agents := make([]ClientInfoToLLM, 0)

    for _, client := range clients {
        if client.IsAgent() {
            // Check if agent is in the requested room OR can access private rooms
            inRequestedRoom := client.RoomID == roomId && roomId != ""
            canAccessPrivateRoom := roomId == clientPrivateRoomID && clientPrivateRoomID != ""

            if inRequestedRoom || canAccessPrivateRoom {
                agents = append(agents, ClientInfoToLLM{
                    ID:           client.ID,
                    Name:         client.Name, 
                    Description:  client.Description,
                    Capabilities: client.Capabilities,
                    Commands:     client.Commands,
                })
            }
        }
    }
    return agents
}
```

---

## Issues and Resolutions

### Issue: Missing Room Context in Agent Messages
**Problem**: SDK-based agents were sending messages but clients were skipping them due to missing `dataRoom` and `messageRoomId` fields.

**Root Cause**: The SDK's `SendTaskResponseToRoom()` method only included the `room` field but not the client-expected `dataRoom` and `messageRoomId` fields.

**Solution**: Updated both the Message struct and SendTaskResponseToRoom method:

1. **Message Struct Update** (`pkg/types/message.go:39-40`):
   ```go
   DataRoom      string `json:"dataRoom,omitempty"`      // Client expected field #1
   MessageRoomId string `json:"messageRoomId,omitempty"` // Client expected field #2
   ```

2. **Protocol Method Update** (`pkg/network/protocol.go:429-440`):
   ```go
   msg := &types.Message{
       Type:          "task_response", 
       Room:          room,             // SDK internal field
       DataRoom:      room,             // Client expected field #1
       MessageRoomId: room,             // Client expected field #2
       // ... other fields
   }
   ```

### Issue: Room Context Loss During Task Distribution
**Problem**: Room context could be lost when tasks were distributed from coordinator to agents.

**Solution**: Ensured room context is preserved at every step:
1. Handler passes room to coordinator: `h.coordinator.HandlePrompt(ctx, prompt, msg.Room, c)`
2. Coordinator includes room in task data and message
3. Agent SDK preserves room context in responses

### Verification
The fix was verified by:
1. Building the updated SDK successfully
2. Testing message flow with debug logging
3. Confirming all room fields are present in agent responses
4. Validating client properly processes messages with room context

---

## Summary

The Teneo system implements a sophisticated room-aware messaging architecture:

1. **Room Context Flows End-to-End**: From user WebSocket clients through the coordinator AI to agents and back
2. **Multiple Room Fields**: Different fields serve different purposes (routing, UI filtering, message grouping)
3. **AI-Powered Agent Selection**: Coordinator uses room context to filter available agents
4. **Task Tracking**: Pending tasks maintain room context for proper response routing
5. **Robust Error Handling**: Graceful fallbacks when room context is missing or invalid

The architecture successfully maintains room isolation while enabling flexible agent-user interactions across different room contexts.
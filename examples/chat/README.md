# TerminusGo Chat Example

This example demonstrates a real-time chat application built with TerminusGo, showcasing:

- Real-time communication patterns
- List widget for scrollable message display
- TextInput widget for message composition
- Timestamps and user identification
- Command system (e.g., `/nick` to change username)
- Layout utilities for split-screen UI
- Async commands for simulating other users
- Typing indicators

## Features

### Real-time Updates
The chat updates in real-time as messages are sent and received. The List widget automatically scrolls to show new messages.

### User Commands
- `/nick <name>` - Change your username
- `/clear` - Clear all messages
- `/time` - Toggle timestamp display
- `/24h` - Toggle 24-hour time format
- `/help` - Show available commands
- `/quit` - Exit the chat

### Simulated Activity
To demonstrate real-time capabilities, the example includes:
- Simulated users that send messages at random intervals
- Typing indicators (though simplified for the demo)
- Online user count

### UI Layout
The interface is divided into several sections:
- **Header**: Shows title, username, message count, and online users
- **Message Area**: Scrollable list of chat messages
- **Typing Indicator**: Shows when other users are typing
- **Input Area**: Text input for composing messages
- **Footer**: Quick command reference

## Code Structure

### Message Types
- `Message`: Core message structure with user, text, timestamp
- `simulatedMessageMsg`: Async message from simulated users
- `typingUpdateMsg`: Update typing indicators

### Key Components
- `ChatModel`: Maintains chat state and widgets
- `ChatComponent`: Main component implementing MVU pattern
- `messageListItem`: Custom ListItem for rendering messages

### Async Patterns
The example demonstrates async command usage:
```go
// Simulate other users with delayed messages
func (c *ChatComponent) simulateOtherUsers() terminus.Cmd {
    return func() terminus.Msg {
        // Wait random time
        delay := time.Duration(5+rand.Intn(10)) * time.Second
        time.Sleep(delay)
        
        // Return message to be processed by Update()
        return simulatedMessageMsg{user: user, text: text}
    }
}
```

## Running the Example

```bash
cd examples/chat
go run main.go
```

Then open http://localhost:8080 in your browser.

## Extending the Example

This example could be extended with:
- User authentication
- Message persistence
- Private messages
- User avatars/colors
- Emoji support
- File sharing
- Message editing/deletion
- Channel/room support
# Gemini Chat Example

An interactive chat application for conversing with Google's Gemini AI using Terminus.

## Features

- 🤖 Real-time chat with Gemini AI
- 💬 Message history with timestamps
- 🎨 Color-coded messages (user vs AI)
- ⌨️ Keyboard shortcuts for common actions
- 📜 Scrollable message history
- 🔄 Automatic text wrapping for long messages
- ⚡ Asynchronous message handling

## Prerequisites

### 1. Get a Gemini API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Get API Key"
4. Copy your API key

### 2. Set Environment Variable

```bash
export GEMINI_API_KEY="your-api-key-here"
```

Or add it to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
echo 'export GEMINI_API_KEY="your-api-key-here"' >> ~/.zshrc
source ~/.zshrc
```

### 3. Install Go Dependencies

```bash
go get google.golang.org/genai
```

## Running the Example

```bash
cd examples/gemini_chat
go run main.go
```

Then open http://localhost:8890 in your browser.

## Usage

### Keyboard Shortcuts

- **Enter** - Send message
- **Ctrl+L** - Clear chat history
- **Ctrl+T** - Toggle timestamps
- **Ctrl+C** or **Esc** - Exit

### Chat Interface

The interface shows:
- 🤖 Title and connection status
- 📊 Message count
- 💬 Scrollable message history
- ✏️ Input field at the bottom

Messages are color-coded:
- **Green** - Your messages
- **Blue** - Gemini's responses
- **Yellow** - System messages

## Code Structure

### Components

- **GeminiChatComponent** - Main component managing the chat
- **GeminiChatModel** - Application state including messages and UI widgets
- **messageListItem** - Custom list item for rendering messages

### Message Types

- **GeminiConnectedMsg** - Successful connection to Gemini
- **GeminiResponseMsg** - Response from Gemini AI
- **GeminiErrorMsg** - Error messages

### Key Features

1. **Async Communication**
   ```go
   func sendToGemini(message string) terminus.Cmd {
       return func() terminus.Msg {
           // Send message to Gemini API
           res, err := chat.SendMessage(ctx, genai.Part{Text: message})
           // Return response as message
       }
   }
   ```

2. **Message History**
   - Messages stored with role, content, and timestamp
   - Auto-scrolling to latest message
   - Persistent within session

3. **Error Handling**
   - Connection errors displayed in UI
   - API key validation
   - Graceful error messages

## Customization Ideas

1. **Model Selection** - Add ability to choose different Gemini models
2. **System Prompts** - Configure AI behavior with system messages
3. **Export Chat** - Save conversation history to file
4. **Markdown Rendering** - Parse and style markdown in responses
5. **Code Highlighting** - Detect and highlight code blocks
6. **Multi-turn Context** - Maintain conversation context across sessions
7. **Streaming Responses** - Show AI responses as they're generated
8. **File Uploads** - Support for image analysis with Gemini

## Troubleshooting

### "GEMINI_API_KEY environment variable not set"
Make sure you've set the environment variable before running the app:
```bash
export GEMINI_API_KEY="your-key"
go run main.go
```

### Connection Errors
- Check your internet connection
- Verify your API key is valid
- Ensure you have API access enabled

### No Response from Gemini
- Check API quotas/limits
- Try a simpler message
- Check for service outages

## API Rate Limits

Be aware of Gemini API rate limits:
- Free tier has usage limits
- Consider implementing rate limiting for production use
- Handle 429 (rate limit) errors gracefully
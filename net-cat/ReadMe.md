# TCPChat Client

The TCPChat client allows users to connect to the server and participate in the chat. Here's what you need to know:

## Usage

1. **Compile:**
   - Compile the client code using Go:
     ```bash
     go build -o TCPChat main.go
     ```

2. **Run:**
   - Run the compiled binary:
     ```bash
     ./TCPChat 2525
     ```
     - Replace `[port]` with the port number (2525).

3. **Join the Chat:**
   - Upon successful connection, you'll be part of the chat.
   - Start typing messages to communicate with other clients.

## Features

- **Unique Name:**
  - Provide a unique name when joining the chat.
  - Example: `./TCPChatClient 192.168.1.100 8989 Alice`, or just your name.

- **Send Messages:**
  - Type your messages and press Enter to send them to the chat.
  - Messages will be timestamped and displayed as `[YYYY-MM-DD HH:MM:SS][YourName]: Your message`.

- **Receive Messages:**
  - You'll receive messages from other clients in real-time.

- **Notifications:**
  - You'll be notified when new clients join or exit the chat.

## Example

1. Connect to the server:
   ```bash
   telnet localhost 2525
   ```

2. Start chatting:
   ```
   [2024-06-25 14:30:15][Alice]: Hello, everyone!
   ```

## Author

```
    @sahahmed
    @aaljar
    @huabbas
```
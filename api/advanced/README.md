# Creek

## Advanced API Features

Creek now includes advanced API functionalities for seamless integration and enhanced data management.

### Real-Time Updates with WebSockets
The real-time updates API enables live communication with Creek via WebSockets. This feature is useful for building dynamic applications.

#### Usage
1. Start the WebSocket server:
   ```bash
   go run api/advanced/realtime_updates.go
   ```
2. Connect to the WebSocket endpoint `/ws` using any WebSocket client.

#### Example Client Code
```javascript
const socket = new WebSocket("ws://localhost:8080/ws");
socket.onmessage = (event) => console.log(event.data);
socket.send("Hello, Creek!");
```

#### Running Tests
To run the WebSocket implementation tests:

1. Navigate to the advanced API directory:
   ```bash
   cd api/advanced
   ```

2. Run all tests:
   ```bash
   go test -v
   ```

3. Run a specific test:
   ```bash
   go test -v -run TestWebSocketConnection
   ```

Available test cases:
- `TestHub`: Verifies hub initialization and channel setup
- `TestWebSocketConnection`: Tests basic WebSocket connectivity and message passing
- `TestConcurrentConnections`: Validates handling of multiple concurrent clients
- `TestClientDisconnection`: Ensures proper client cleanup on disconnection
- `TestMessageBuffering`: Verifies ordered message delivery

To run tests with coverage report:
```bash
go test -v -cover
```

The WebSocket implementation provides:
- Concurrent client handling
- Broadcast messaging
- Automatic connection management
- Buffer management for message sending/receiving
- Clean connection termination

To use this functionality in your application:
1. Import the advanced package
2. Initialize the WebSocket hub
3. Start the WebSocket server

Note: Remember to implement proper security measures in production environments.
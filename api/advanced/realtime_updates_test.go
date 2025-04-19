package advanced

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/gorilla/websocket"
)

func TestHub(t *testing.T) {
    hub := NewHub()
    if hub == nil {
        t.Fatal("NewHub() returned nil")
    }
    if hub.clients == nil {
        t.Fatal("hub.clients is nil")
    }
    if hub.broadcast == nil {
        t.Fatal("hub.broadcast is nil")
    }
    if hub.register == nil {
        t.Fatal("hub.register is nil")
    }
    if hub.unregister == nil {
        t.Fatal("hub.unregister is nil")
    }
}

func TestWebSocketConnection(t *testing.T) {
    // Create a new hub
    hub := NewHub()
    go hub.Run()

    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ServeWs(hub, w, r)
    }))
    defer server.Close()

    // Convert http URL to ws URL
    url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    // Connect first client
    ws1, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("could not open websocket connection: %v", err)
    }
    defer ws1.Close()

    // Connect second client
    ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("could not open websocket connection: %v", err)
    }
    defer ws2.Close()

    // Test message broadcasting
    testMessage := []byte("Hello, Creek!")
    if err := ws1.WriteMessage(websocket.TextMessage, testMessage); err != nil {
        t.Fatalf("could not send message: %v", err)
    }

    // Read message from second client
    _, msg, err := ws2.ReadMessage()
    if err != nil {
        t.Fatalf("could not read message: %v", err)
    }

    if string(msg) != string(testMessage) {
        t.Fatalf("expected message %q but got %q", testMessage, msg)
    }
}

func TestConcurrentConnections(t *testing.T) {
    hub := NewHub()
    go hub.Run()

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ServeWs(hub, w, r)
    }))
    defer server.Close()

    url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    // Create multiple concurrent connections
    numClients := 10
    clients := make([]*websocket.Conn, numClients)
    
    for i := 0; i < numClients; i++ {
        ws, _, err := websocket.DefaultDialer.Dial(url, nil)
        if err != nil {
            t.Fatalf("could not open websocket connection %d: %v", i, err)
        }
        defer ws.Close()
        clients[i] = ws
    }

    // Wait for all connections to be established
    time.Sleep(100 * time.Millisecond)

    // Send message from first client
    testMessage := []byte("Broadcast test")
    if err := clients[0].WriteMessage(websocket.TextMessage, testMessage); err != nil {
        t.Fatalf("could not send message: %v", err)
    }

    // Verify all other clients receive the message
    for i := 1; i < numClients; i++ {
        _, msg, err := clients[i].ReadMessage()
        if err != nil {
            t.Fatalf("client %d could not read message: %v", i, err)
        }
        if string(msg) != string(testMessage) {
            t.Fatalf("client %d: expected message %q but got %q", i, testMessage, msg)
        }
    }
}

func TestClientDisconnection(t *testing.T) {
    hub := NewHub()
    go hub.Run()

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ServeWs(hub, w, r)
    }))
    defer server.Close()

    url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    // Connect a client
    ws, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("could not open websocket connection: %v", err)
    }

    // Wait for connection to be established
    time.Sleep(100 * time.Millisecond)

    // Close the connection
    ws.Close()

    // Wait for hub to process disconnection
    time.Sleep(100 * time.Millisecond)

    // Verify client was removed from hub
    hub.mutex.Lock()
    clientCount := len(hub.clients)
    hub.mutex.Unlock()

    if clientCount != 0 {
        t.Fatalf("expected 0 clients after disconnection, but got %d", clientCount)
    }
}

func TestMessageBuffering(t *testing.T) {
    hub := NewHub()
    go hub.Run()

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ServeWs(hub, w, r)
    }))
    defer server.Close()

    url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    // Connect two clients
    ws1, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("could not open first websocket connection: %v", err)
    }
    defer ws1.Close()

    ws2, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        t.Fatalf("could not open second websocket connection: %v", err)
    }
    defer ws2.Close()

    // Send multiple messages rapidly
    messages := []string{
        "Message 1",
        "Message 2",
        "Message 3",
    }

    for _, msg := range messages {
        if err := ws1.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
            t.Fatalf("could not send message: %v", err)
        }
    }

    // Verify all messages are received in order
    for _, expectedMsg := range messages {
        _, msg, err := ws2.ReadMessage()
        if err != nil {
            t.Fatalf("could not read message: %v", err)
        }
        if string(msg) != expectedMsg {
            t.Fatalf("expected message %q but got %q", expectedMsg, string(msg))
        }
    }
}
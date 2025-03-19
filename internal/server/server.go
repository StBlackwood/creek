package server

import (
	"log"
	"net"
	"sync"
)

// Server represents a TCP server
type Server struct {
	address  string
	clients  map[net.Conn]bool
	mu       sync.Mutex
	listener net.Listener
	done     chan struct{}
}

// New creates a new Server instance
func New(address string) *Server {
	return &Server{
		address: address,
		clients: make(map[net.Conn]bool),
		done:    make(chan struct{}),
	}
}

// Start begins listening for TCP connections
func (s *Server) Start() {
	var err error
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	log.Printf("Server listening on %s", s.address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				log.Printf("Error accepting connection: %v", err)
			}
			continue
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		log.Printf("New client connected: %v", conn.RemoteAddr())
		go s.handleClient(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	close(s.done)
	s.listener.Close()

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		delete(s.clients, conn)
		conn.Close()
	}
}

// handleClient manages an individual client connection
func (s *Server) handleClient(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}(conn)

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Client %v disconnected", conn.RemoteAddr())
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			return
		}

		message := string(buf[:n])
		log.Printf("Received from %v: %s", conn.RemoteAddr(), message)

		// Echo the message back to the client
		_, err = conn.Write([]byte("Echo: " + message))
		if err != nil {
			log.Printf("Error sending response: %v", err)
			return
		}
	}
}

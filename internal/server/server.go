package server

import (
	"creek/internal/logger"
	"creek/internal/version"
	"fmt"
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
	log := logger.GetLogger()

	var err error
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	log.Infof("Server listening on %s", s.address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				log.Warnf("Error accepting connection: %v", err)
			}
			continue
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		log.Debugf("New client connected: %v", conn.RemoteAddr())
		go s.handleClient(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	log := logger.GetLogger()

	close(s.done)
	err := s.listener.Close()
	if err != nil {
		log.Errorf("Error closing listener: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		delete(s.clients, conn)
		err := conn.Close()
		if err != nil {
			log.Errorf("Error closing connection: %v", err)
			return
		}
	}
}

// handleClient manages an individual client connection
func (s *Server) handleClient(conn net.Conn) {
	log := logger.GetLogger()

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Warnf("Error closing client connection: %v", err)
		}
	}(conn)

	versionMsg := fmt.Sprintf("Connected to Server Version: %s\n", version.Version)

	_, err := conn.Write([]byte(versionMsg))
	if err != nil {
		log.Warnf("Error sending response: %v", err)
		return
	}

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Debugf("Client %v disconnected", conn.RemoteAddr())
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			return
		}

		message := string(buf[:n])
		log.Tracef("Received from %v: %s", conn.RemoteAddr(), message)

		// echo the message back to client
		reply := "Echo: " + message

		log.Tracef("Sending reply: %s", reply)
		_, err = conn.Write([]byte(reply))
		if err != nil {
			log.Warnf("Error sending response: %v", err)
			return
		}
	}
}

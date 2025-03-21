package server

import (
	"bufio"
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/core"
	"creek/internal/handler"
	"creek/internal/logger"
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
	Conf     *config.Config
	sm       *core.StateMachine
}

// New creates a new Server instance
func New(cfg *config.Config) *Server {

	stateMachine, err := core.NewStateMachine(cfg)
	if err != nil {
	}
	return &Server{
		address: cfg.ServerAddress,
		clients: make(map[net.Conn]bool),
		done:    make(chan struct{}),
		Conf:    cfg,
		sm:      stateMachine,
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
	}

	err = s.sm.Stop()
	if err != nil {
		log.Errorf("Error stopping state machine: %v", err)
	} // Stop datastore and GC

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		delete(s.clients, conn)
		err := conn.Close()
		if err != nil {
			log.Errorf("Error closing connection: %v", err)
			continue
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

	versionMsg := fmt.Sprintf("Connected to Server Version: %s\n", commons.Version)

	s.SendMsg(conn, versionMsg)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Debug("Client disconnected: ", conn.RemoteAddr())
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			return
		}

		log.Trace("Received from ", conn.RemoteAddr(), ": ", message)

		// Process and respond to message
		response, err := handler.HandleMessage(s.sm, message)
		if err != nil {
			log.Warnf("Error handling message: %v", err)
			s.SendMsg(conn, err.Error())
			continue
		}
		log.Tracef("Sending response: %v to client %v", response, conn.RemoteAddr())
		s.SendMsg(conn, response)
	}
}

func (s *Server) SendMsg(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response + "\n"))
	if err != nil {
		log := logger.GetLogger()
		log.Warnf("Error sending msg: %v to client %v", err, conn.RemoteAddr())
	}
}

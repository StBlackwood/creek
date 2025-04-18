package server

import (
	"bufio"
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/core"
	"creek/internal/logger"
	"creek/internal/replication"
	"fmt"
	"github.com/sirupsen/logrus"
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
	rs       *replication.RepService
	log      *logrus.Logger
}

// New creates a new Server instance
func New(cfg *config.Config) *Server {

	replicationService, err := replication.NewRepService(cfg)
	if err != nil {
		panic(err)
	}

	stateMachine, err := core.NewStateMachine(replicationService.GetSelfNodeId(), cfg)
	if err != nil {
		panic(err)
	}

	return &Server{
		address: cfg.ServerAddress,
		clients: make(map[net.Conn]bool),
		done:    make(chan struct{}),
		Conf:    cfg,
		sm:      stateMachine,
		rs:      replicationService,
		log:     logger.CreateLogger(cfg.LogLevel),
	}
}

// Start begins listening for TCP connections
func (s *Server) Start() {
	var err error

	err = s.sm.Start()
	if err != nil {
		panic(err)
	}

	s.rs.ConnectToFollowers()
	s.sm.AttachRepCmdWriteHandlerToPartitions(func(cmd *replication.RepCmd) error {
		return s.rs.HandleRepCmdWrite(cmd)
	})

	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		s.log.Fatalf("Error starting server: %v", err)
	}

	s.log.Infof("Server listening on %s", s.address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				s.log.Warnf("Error accepting connection: %v", err)
			}
			continue
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		s.log.Debugf("New client connected: %v", conn.RemoteAddr())
		go s.handleClient(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {

	close(s.done)
	err := s.listener.Close()
	if err != nil {
		s.log.Errorf("Error closing listener: %v", err)
	}

	err = s.sm.Stop()
	if err != nil {
		s.log.Errorf("Error stopping state machine: %v", err)
	} // Stop datastore and GC

	err = s.rs.Stop()
	if err != nil {
		s.log.Errorf("Error stopping replication service: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		delete(s.clients, conn)
		err := conn.Close()
		if err != nil {
			s.log.Errorf("Error closing connection: %v", err)
			continue
		}
	}
}

// handleClient manages an individual client connection
func (s *Server) handleClient(conn net.Conn) {

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			s.log.Warnf("Error closing client connection: %v", err)
		}
	}(conn)

	versionMsg := fmt.Sprintf("Connected to Server Version: %s\n", commons.Version)

	s.SendMsg(conn, versionMsg)

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			s.log.Debug("Client disconnected: ", conn.RemoteAddr())
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			return
		}

		s.log.Trace("Received from ", conn.RemoteAddr(), ": ", message)

		// Process and respond to message
		response, err := handleMessage(s, message)
		if err != nil {
			s.log.Warnf("Error handling message: %v", err)
			s.SendMsg(conn, err.Error())
			continue
		}
		s.log.Tracef("Sending response: %v to client %v", response, conn.RemoteAddr())
		s.SendMsg(conn, response)
	}
}

func (s *Server) SendMsg(conn net.Conn, response string) {
	_, err := conn.Write([]byte(response + "\n"))
	if err != nil {
		s.log.Warnf("Error sending msg: %v to client %v", err, conn.RemoteAddr())
	}
}

package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/hardikphalet/go-redis/internal/store"
)

type Server struct {
	listener net.Listener
	store    store.Store
	port     string
	wg       sync.WaitGroup
	quit     chan struct{}
}

// New creates a new Redis server instance
func New(address string) *Server {
	return &Server{
		port:  address,
		store: store.NewMemoryStore(),
		quit:  make(chan struct{}),
	}
}

// Start initializes the server and starts listening for connections
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Printf("Server listening on %s", s.port)

	// Accept connections in a separate goroutine
	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	// Signal all goroutines to stop
	close(s.quit)

	// Close the listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Wait for all connections to finish
	s.wg.Wait()
	return nil
}

// acceptConnections handles incoming connections
func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection processes individual client connections
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New client connection from %s", remoteAddr)

	handler := NewHandler(conn, s.store)
	if err := handler.Handle(); err != nil {
		log.Printf("Error handling connection from %s: %v", remoteAddr, err)
	} else {
		log.Printf("Client %s disconnected", remoteAddr)
	}
}

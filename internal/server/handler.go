package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/hardikphalet/go-redis/internal/resp"
	"github.com/hardikphalet/go-redis/internal/store"
)

type Handler struct {
	conn       net.Conn
	reader     *bufio.Reader
	writer     *bufio.Writer
	store      store.Store
	parser     *resp.Parser
	respWriter *resp.Writer
}

func NewHandler(conn net.Conn, store store.Store) *Handler {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Handler{
		conn:       conn,
		reader:     reader,
		writer:     writer,
		store:      store,
		parser:     resp.NewParser(reader),
		respWriter: resp.NewWriter(writer),
	}
}

func (h *Handler) Handle() error {
	for {
		// Parse the incoming command using RESP protocol
		command, err := h.parser.Parse()
		if err != nil {
			return fmt.Errorf("error parsing command: %w", err)
		}

		// Execute the command
		response, err := command.Execute(h.store)
		if err != nil {
			if err := h.writeError(err); err != nil {
				return fmt.Errorf("error writing error response: %w", err)
			}
			continue
		}

		// Write the response using RESP protocol
		if err := h.writeResponse(response); err != nil {
			return fmt.Errorf("error writing response: %w", err)
		}
	}
}

func (h *Handler) writeResponse(response interface{}) error {
	return h.respWriter.WriteInterface(response)
}

func (h *Handler) writeError(err error) error {
	return h.respWriter.WriteError(err)
}

package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/hardikphalet/go-redis/internal/store"
)

type Handler struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	store  store.Store
}

func NewHandler(conn net.Conn, store store.Store) *Handler {
	return &Handler{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		store:  store,
	}
}

func (h *Handler) Handle() error {
	for {
		// Read the incoming command
		command, err := h.readCommand()
		if err != nil {
			return fmt.Errorf("error reading command: %w", err)
		}

		// Execute the command
		response, err := h.executeCommand(command)
		if err != nil {
			if err := h.writeError(err); err != nil {
				return fmt.Errorf("error writing error response: %w", err)
			}
			continue
		}

		// Write the response
		if err := h.writeResponse(response); err != nil {
			return fmt.Errorf("error writing response: %w", err)
		}
	}
}

func (h *Handler) readCommand() (string, error) {
	// TODO: Implement RESP protocol parsing
	// For now, just read a line
	line, err := h.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line[:len(line)-1], nil
}

func (h *Handler) executeCommand(command string) (interface{}, error) {
	// TODO: Implement command parsing and execution
	return fmt.Sprintf("Echo: %s", command), nil
}

func (h *Handler) writeResponse(response interface{}) error {
	// TODO: Implement RESP protocol writing
	// For now, just write the string response
	_, err := fmt.Fprintf(h.writer, "+%v\r\n", response)
	if err != nil {
		return err
	}
	return h.writer.Flush()
}

func (h *Handler) writeError(err error) error {
	_, writeErr := fmt.Fprintf(h.writer, "-ERR %v\r\n", err)
	if writeErr != nil {
		return writeErr
	}
	return h.writer.Flush()
}

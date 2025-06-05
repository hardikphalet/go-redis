package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands"
)

var (
	ErrInvalidSyntax = errors.New("invalid RESP syntax")
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(reader *bufio.Reader) *Parser {
	return &Parser{reader: reader}
}

// Parse reads the RESP protocol input and returns a Command
func (p *Parser) Parse() (commands.Command, error) {
	// Read the first byte to determine the type
	firstByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch firstByte {
	case '*':
		return p.parseArray()
	case '$':
		return nil, fmt.Errorf("bulk string must be part of an array")
	case '+':
		return nil, fmt.Errorf("simple string must be part of an array")
	case ':':
		return nil, fmt.Errorf("integer must be part of an array")
	case '-':
		return nil, fmt.Errorf("error must be part of an array")
	default:
		return nil, ErrInvalidSyntax
	}
}

// parseArray parses a RESP array
func (p *Parser) parseArray() (commands.Command, error) {
	// Read the array length
	length, err := p.readInteger()
	if err != nil {
		return nil, err
	}

	if length < 1 {
		return nil, fmt.Errorf("array length must be positive")
	}

	// Read all array elements
	elements := make([]string, length)
	for i := 0; i < length; i++ {
		element, err := p.readBulkString()
		if err != nil {
			return nil, err
		}
		elements[i] = element
	}

	// Convert array elements to a command
	return p.createCommand(elements)
}

// readInteger reads a RESP integer
func (p *Parser) readInteger() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}

	n, err := strconv.Atoi(line)
	if err != nil {
		return 0, ErrInvalidSyntax
	}

	return n, nil
}

// readBulkString reads a RESP bulk string
func (p *Parser) readBulkString() (string, error) {
	// Read the $ character
	b, err := p.reader.ReadByte()
	if err != nil {
		return "", err
	}
	if b != '$' {
		return "", ErrInvalidSyntax
	}

	// Read the string length
	length, err := p.readInteger()
	if err != nil {
		return "", err
	}

	if length < 0 {
		return "", nil // Null bulk string
	}

	// Read the string content
	data := make([]byte, length)
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return "", err
	}

	// Read and verify CRLF
	if err := p.readCRLF(); err != nil {
		return "", err
	}

	return string(data), nil
}

// readLine reads a line ending in CRLF
func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", ErrInvalidSyntax
	}

	return line[:len(line)-2], nil
}

// readCRLF reads and verifies a CRLF sequence
func (p *Parser) readCRLF() error {
	cr, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	if cr != '\r' {
		return ErrInvalidSyntax
	}

	lf, err := p.reader.ReadByte()
	if err != nil {
		return err
	}
	if lf != '\n' {
		return ErrInvalidSyntax
	}

	return nil
}

// createCommand converts string array to a specific command
func (p *Parser) createCommand(args []string) (commands.Command, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// Convert command to uppercase for case-insensitive comparison
	cmd := strings.ToUpper(args[0])

	switch cmd {
	case "SET":
		if len(args) != 3 {
			return nil, fmt.Errorf("SET command requires exactly 2 arguments")
		}
		return &commands.SetCommand{
			Key:   args[1],
			Value: args[2],
		}, nil

	case "GET":
		if len(args) != 2 {
			return nil, fmt.Errorf("GET command requires exactly 1 argument")
		}
		return &commands.GetCommand{
			Key: args[1],
		}, nil

	case "DEL":
		if len(args) < 2 {
			return nil, fmt.Errorf("DEL command requires at least 1 argument")
		}
		return &commands.DelCommand{
			Keys: args[1:],
		}, nil

	case "EXPIRE":
		if len(args) != 3 {
			return nil, fmt.Errorf("EXPIRE command requires exactly 2 arguments")
		}
		ttl, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid TTL value")
		}
		return &commands.ExpireCommand{
			Key: args[1],
			TTL: time.Duration(ttl) * time.Second,
		}, nil

	case "TTL":
		if len(args) != 2 {
			return nil, fmt.Errorf("TTL command requires exactly 1 argument")
		}
		return &commands.TtlCommand{
			Key: args[1],
		}, nil

	case "KEYS":
		if len(args) != 2 {
			return nil, fmt.Errorf("KEYS command requires exactly 1 argument")
		}
		return &commands.KeysCommand{
			Pattern: args[1],
		}, nil

	case "ZADD":
		if len(args) < 4 || (len(args)-2)%2 != 0 {
			return nil, fmt.Errorf("ZADD command requires at least one score-member pair")
		}

		key := args[1]
		members := make([]commands.ScoreMember, 0, (len(args)-2)/2)

		// Parse score-member pairs
		for i := 2; i < len(args); i += 2 {
			score, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid score value: %s", args[i])
			}
			members = append(members, commands.ScoreMember{
				Score:  score,
				Member: args[i+1],
			})
		}

		return &commands.ZAddCommand{
			Key:     key,
			Members: members,
		}, nil

	case "ZRANGE":
		if len(args) < 4 {
			return nil, fmt.Errorf("ZRANGE command requires at least 3 arguments")
		}
		start, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, fmt.Errorf("invalid start index")
		}
		stop, err := strconv.Atoi(args[3])
		if err != nil {
			return nil, fmt.Errorf("invalid stop index")
		}
		withScores := false
		if len(args) > 4 && strings.ToUpper(args[4]) == "WITHSCORES" {
			withScores = true
		}
		return &commands.ZRangeCommand{
			Key:        args[1],
			Start:      start,
			Stop:       stop,
			WithScores: withScores,
		}, nil

	case "COMMAND":
		return &commands.CommandCommand{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

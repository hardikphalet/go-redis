package resp

import (
	"bufio"

	"github.com/hardikphalet/go-redis/internal/commands"
)

type Parser struct {
	reader *bufio.Reader
}

func (p *Parser) Parse() (commands.Command, error)

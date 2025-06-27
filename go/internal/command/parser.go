package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input string) (*Command, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no command")
	}

	cmd := strings.ToUpper(parts[0])

	switch cmd {
	case "SELECT":
		return p.SELECT(parts)

	// * KV 操作
	case "GET":
		return p.GET(parts)
	case "SET":
		return p.SET(parts)
	case "DEL":
		return p.DEL(parts)
	case "EXISTS":
		return p.EXISTS(parts)
	case "KEYS":
		return p.KEYS(parts)
	case "TYPE":
		return p.TYPE(parts)

	// * DOC 操作
	case "FIND":
		return p.FIND(parts)
	case "ADD":
		return p.ADD(parts)

	// * TTL 操作
	case "TTL":
		return p.TTL(parts)
	case "EXPIRE":
		return p.EXPIRE(parts)
	case "PERSIST":
		return p.PERSIST(parts)

	// * 其他操作
	case "HELP":
		return p.HELP(parts)
	case "PING":
		return p.PING(parts)

	default:
		return nil, fmt.Errorf("unknown command: %s", cmd)
	}
}

func (p *Parser) SELECT(part []string) (*Command, error) {
	if len(part) != 2 {
		return nil, fmt.Errorf("usage: SELECT <db:int>")
	}

	num, err := strconv.Atoi(part[1])
	if err != nil {
		return nil, fmt.Errorf("invalid database number: %s", part[1])
	}

	if num < 0 || num > 15 {
		return nil, fmt.Errorf("database must be between 0 and 15")
	}

	cmd := NewCommand(SELECT)
	cmd.SetArg("db", num)
	return cmd, nil
}

func (p *Parser) GET(part []string) (*Command, error) {
	if len(part) != 2 {
		return nil, fmt.Errorf("usage: GET <key>")
	}

	cmd := NewCommand(GET)
	cmd.SetArg("key", part[1])
	return cmd, nil
}

func (p *Parser) SET(part []string) (*Command, error) {
	if len(part) < 3 || len(part) > 4 {
		return nil, fmt.Errorf("usage: SET <key> <value> [ttl_second|expire_time]")
	}

	cmd := NewCommand(SET)
	cmd.SetArg("key", part[1])
	cmd.SetArg("value", part[2])

	if len(part) == 4 {
		ttl, err := parseTTL(part[3])
		if err != nil {
			return nil, fmt.Errorf("invalid expire time: %s", part[3])
		}

		cmd.SetArg("ttl", ttl)
	}

	return cmd, nil
}

func parseTTL(value string) (uint64, error) {
	list := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15",
		"2006-01-02",
		"2006-01",
	}

	for _, format := range list {
		if expireTime, err := time.ParseInLocation(format, value, time.Local); err == nil {
			return uint64(time.Until(expireTime).Seconds()), nil
		}
	}

	return strconv.ParseUint(value, 10, 64)
}

func (p *Parser) DEL(part []string) (*Command, error) {
	if len(part) < 2 {
		return nil, fmt.Errorf("usage: DEL <key1> [key2] ...")
	}

	cmd := NewCommand(DEL)
	cmd.SetArg("keys", part[1:])
	return cmd, nil
}

func (p *Parser) EXISTS(part []string) (*Command, error) {
	if len(part) != 2 {
		return nil, fmt.Errorf("usage: EXISTS <key>")
	}

	cmd := NewCommand(EXISTS)
	cmd.SetArg("key", part[1])
	return cmd, nil
}

func (p *Parser) KEYS(part []string) (*Command, error) {
	if len(part) != 2 {
		return nil, fmt.Errorf("usage: KEYS <pattern>")
	}

	cmd := NewCommand(KEYS)
	cmd.SetArg("pattern", part[1])
	return cmd, nil
}

func (p *Parser) TYPE(part []string) (*Command, error) {
	if len(part) != 2 {
		return nil, fmt.Errorf("usage: TYPE <key>")
	}

	cmd := NewCommand(TYPE)
	cmd.SetArg("key", part[1])
	return cmd, nil
}

// TODO: 實作 filters 細節
func (p *Parser) FIND(part []string) (*Command, error) {
	if len(part) < 2 {
		return nil, fmt.Errorf("usage: FIND <key> [filters]")
	}

	cmd := NewCommand(FIND)
	cmd.SetArg("key", part[1])

	if len(part) > 2 {
		filters := strings.Join(part[2:], " ")
		cmd.SetArg("filters", filters)
	}

	return cmd, nil
}

func (p *Parser) ADD(part []string) (*Command, error) {
	if len(part) != 3 {
		return nil, fmt.Errorf("usage: ADD <key> <value>")
	}

	cmd := NewCommand(ADD)
	cmd.SetArg("key", part[1])
	cmd.SetArg("value", part[2])
	return cmd, nil
}

func (p *Parser) TTL(part []string) (*Command, error) {
	if len(part) < 2 {
		return nil, fmt.Errorf("usage: TTL <key> [filters]")
	}

	cmd := NewCommand(TTL)
	cmd.SetArg("key", part[1])

	if len(part) > 2 {
		filters := strings.Join(part[2:], " ")
		cmd.SetArg("filters", filters)
	}

	return cmd, nil
}

func (p *Parser) EXPIRE(part []string) (*Command, error) {
	if len(part) < 3 {
		return nil, fmt.Errorf("usage: EXPIRE <key> <ttl_second|expire_time> [filters]")
	}

	cmd := NewCommand(EXPIRE)
	cmd.SetArg("key", part[1])

	ttl, err := parseTTL(part[2])
	if err != nil {
		return nil, fmt.Errorf("invalid expire time: %s", part[2])
	}

	// TODO: 檢查
	cmd.SetArg("ttl", ttl)

	if len(part) > 3 {
		filters := strings.Join(part[3:], " ")
		cmd.SetArg("filters", filters)
	}

	return cmd, nil
}

func (p *Parser) PERSIST(part []string) (*Command, error) {
	if len(part) < 2 {
		return nil, fmt.Errorf("usage: PERSIST <key> [filters]")
	}

	cmd := NewCommand(PERSIST)
	cmd.SetArg("key", part[1])

	// TODO: 檢查
	cmd.SetArg("ttl", 0)

	if len(part) > 2 {
		filters := strings.Join(part[3:], " ")
		cmd.SetArg("filters", filters)
	}

	return cmd, nil
}

func (p *Parser) HELP(part []string) (*Command, error) {
	return NewCommand(HELP), nil
}

func (p *Parser) PING(part []string) (*Command, error) {
	return NewCommand(PING), nil
}

package server

import (
	"fmt"
	"strings"

	"go-jsondb/internal/command"
)

type Client struct {
	db     int
	server *Server
}

func (c *Client) Exec(cmd *command.Command) string {
	switch cmd.Type {
	// * KV 操作
	case command.GET:
		return c.GET(cmd)
	case command.SET:
		return c.SET(cmd)
	case command.DEL:
		return c.DEL(cmd)
	case command.EXISTS:
		return c.EXISTS(cmd)
	case command.KEYS:
		return c.KEYS(cmd)
	case command.TYPE:
		return c.TYPE(cmd)

	// * DOC 操作
	case command.FIND:
		return c.FIND(cmd)
	case command.ADD:
		return c.ADD(cmd)

	// * TTL 操作
	case command.TTL:
		return c.TTL(cmd)
	case command.EXPIRE:
		return c.EXPIRE(cmd)
	case command.PERSIST:
		return c.PERSIST(cmd)

	// * 其他操作
	case command.SELECT:
		return c.SELECT(cmd)
	case command.HELP:
		return c.HELP(cmd)
	case command.PING:
		return c.PING(cmd)

	default:
		return fmt.Sprintf("Error: Unknown command type: %v", cmd.Type)
	}
}

func (c *Client) SELECT(cmd *command.Command) string {
	db := cmd.GetInt("db")

	if db < 0 || db > 15 {
		return "Error: DB index is out of range (0-15)"
	}

	c.server.mu.Lock()
	err := c.server.checkDB(db)
	c.server.mu.Unlock()
	if err != nil {
		return fmt.Sprintf("Error initializing database %d: %v", db, err)
	}

	c.db = db

	return "OK"
}

func (c *Client) HELP(cmd *command.Command) string {
	str := `
JsonDB Commands:

KV operations:
  GET <key>                    - Get value by key
  SET <key> <value> [ttl]      - Set key-value pair with optional TTL
  DEL <key1> [key2] ...        - Delete one or more keys  
  EXISTS <key>                 - Check if key exists
  KEYS <pattern>               - Find keys matching pattern
  TYPE <key>                   - Get value type of key

DOC operations:
  FIND <key> [filters]         - Query documents with filters
  ADD <key> <value>            - Add document to collection

TTL operations:
  TTL <key> [filters]          - Get remaining TTL
  EXPIRE <key> <seconds>       - Set key expiration

Database:
  SELECT <db_number>           - Select database (0-15)

Utility:
  PING                         - Test connection
  HELP                         - Show this help
  QUIT/EXIT                    - Close connection

Note: Advanced features are currently in progress.
`
	return strings.TrimSpace(str)
}

func (c *Client) PING(cmd *command.Command) string {
	return "PONG"
}

package server

import (
	"fmt"

	"go-jsondb/internal/command"
)

// TODO: 實現 FIND
func (c *Client) FIND(cmd *command.Command) string {
	key := cmd.GetStr("key")
	filters := cmd.GetStr("filters")

	return fmt.Sprintf("FIND operation in progress for key: %s with filters: %s", key, filters)
}

// TODO: 實現 ADD
func (c *Client) ADD(cmd *command.Command) string {
	key := cmd.GetStr("key")
	value := cmd.GetStr("value")

	return fmt.Sprintf("ADD operation in progress for key: %s, value: %s", key, value)
}

package server

import (
	"fmt"
	"time"

	"go-jsondb/internal/command"
	"go-jsondb/internal/storage"
	"go-jsondb/internal/util"
)

func (c *Client) GET(cmd *command.Command) string {
	key := cmd.GetStr("key")

	// * 可能會需要刪除過期資料
	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	if e, isExist := data[key]; isExist {
		if e.ExpireAt != nil && time.Now().Unix() >= *e.ExpireAt {
			c.server.delFromMem(c.db, key)
			return "(nil)"
		}
		return e.Value
	}
	return "(nil)"
}

func (c *Client) SET(cmd *command.Command) string {
	key := cmd.GetStr("key")
	value := cmd.GetStr("value")

	c.server.mu.Lock()
	defer c.server.mu.Unlock()
	valueType := util.GetType(value)

	entry := &storage.Entry{
		Value: value,
		Type:  valueType,
	}

	if ttl, hasTTL := cmd.GetArg("ttl"); hasTTL {
		if ttlValue, ok := ttl.(uint64); ok {
			expireTime := time.Now().Unix() + int64(ttlValue)
			entry.ExpireAt = &expireTime
		}
	}

	c.server.db[c.db][key] = entry

	if err := c.server.checkDB(c.db); err != nil {
		return fmt.Sprintf("Error creating writer: %v", err)
	}

	writer := c.server.writer[c.db]

	var sec *uint64
	if entry.ExpireAt != nil {
		remain := *entry.ExpireAt - time.Now().Unix()
		if remain > 0 {
			ttl := uint64(remain)
			sec = &ttl
		}
	}

	if err := writer.WriteWithTTL("SET", key, value, sec); err != nil {
		return fmt.Sprintf("Error writing to AOF: %v", err)
	}

	cache := storage.Cache{
		Key:       key,
		Value:     value,
		Type:      valueType,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		ExpireAt:  entry.ExpireAt,
	}

	if err := writer.Save(key, cache); err != nil {
		return fmt.Sprintf("Error writing to file: %v", err)
	}

	if _, hasTTL := cmd.GetArg("ttl"); hasTTL {
		return "OK (TTL support in progress)"
	}
	return "OK"
}

func (c *Client) DEL(cmd *command.Command) string {
	list := cmd.GetStrAry("keys")

	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]

	if err := c.server.checkDB(c.db); err != nil {
		return fmt.Sprintf("Error creating writer: %v", err)
	}

	writer := c.server.writer[c.db]
	deleted := 0

	for _, e := range list {
		if _, isExist := data[e]; isExist {
			delete(data, e)
			deleted++

			if err := writer.Write("DEL", e, nil); err != nil {
				return fmt.Sprintf("Error writing to AOF: %v", err)
			}

			if err := writer.Delete(e); err != nil {
				fmt.Printf("Warning: failed to delete file for key %s: %v\n", e, err)
			}
		}
	}

	return fmt.Sprintf("(integer) %d", deleted)
}

func (c *Client) EXISTS(cmd *command.Command) string {
	var result = c.GET(cmd)

	if result == "(nil)" {
		return "(integer) 0"
	}
	return "(integer) 1"
}

func (c *Client) KEYS(cmd *command.Command) string {
	pattern := cmd.GetStr("pattern")

	// * 可能會需要刪除過期資料
	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	var list []string
	var delList []string

	for key, entry := range data {
		if entry.ExpireAt != nil && time.Now().Unix() >= *entry.ExpireAt {
			delList = append(delList, key)
		} else if c.matchPattern(key, pattern) {
			list = append(list, key)
		}
	}

	for _, key := range delList {
		c.server.delFromMem(c.db, key)
	}

	if len(list) == 0 {
		return "(empty)"
	}

	return fmt.Sprintf("%v", list)
}

func (c *Client) TYPE(cmd *command.Command) string {
	key := cmd.GetStr("key")

	// * 可能會需要刪除過期資料
	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	if e, isExist := data[key]; isExist {
		if e.ExpireAt != nil && time.Now().Unix() >= *e.ExpireAt {
			c.server.delFromMem(c.db, key)
			return "none"
		}
		return e.Type
	}
	return "none"
}

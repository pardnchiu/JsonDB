package server

import (
	"fmt"
	"time"

	"go-jsondb/internal/command"
	"go-jsondb/internal/storage"
)

// TODO: 實現 TTL with filters
func (c *Client) TTL(cmd *command.Command) string {
	key := cmd.GetStr("key")

	// * 可能會需要刪除過期資料
	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	if entry, isExist := data[key]; isExist {
		if entry.ExpireAt == nil {
			return "(integer) -1"
		}

		now := time.Now().Unix()
		if now >= *entry.ExpireAt {
			c.server.delFromMem(c.db, key)
			return "(integer) -2"
		}

		remaining := *entry.ExpireAt - now
		return fmt.Sprintf("(integer) %d", remaining)
	}

	return "(integer) -2"
}

// TODO: 實現 EXPIRE with filters
func (c *Client) EXPIRE(cmd *command.Command) string {
	key := cmd.GetStr("key")
	ttl := cmd.GetUint64("ttl")

	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	entry, isExist := data[key]
	if !isExist {
		return "(integer) 0"
	}

	if entry.ExpireAt != nil && time.Now().Unix() >= *entry.ExpireAt {
		c.server.delFromMem(c.db, key)
		return "(integer) 0"
	}

	expire := time.Now().Unix() + int64(ttl)
	entry.ExpireAt = &expire

	if err := c.server.checkDB(c.db); err != nil {
		return fmt.Sprintf("Error creating writer: %v", err)
	}

	writer := c.server.writer[c.db]

	if err := writer.Write("SET", key, map[string]interface{}{
		"seconds":   ttl,
		"expire_at": expire,
	}); err != nil {
		return fmt.Sprintf("Error writing to AOF: %v", err)
	}

	reader := c.server.reader[c.db]
	existingCache, err := reader.Read(key)
	if err != nil {
		existingCache = &storage.Cache{
			Key:       key,
			Value:     entry.Value,
			Type:      entry.Type,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			ExpireAt:  entry.ExpireAt,
		}
	}

	cache := storage.Cache{
		Key:       key,
		Value:     entry.Value,
		Type:      entry.Type,
		CreatedAt: existingCache.CreatedAt,
		UpdatedAt: time.Now().Unix(),
		ExpireAt:  entry.ExpireAt,
	}

	if err := writer.Save(key, cache); err != nil {
		return fmt.Sprintf("Error updating file: %v", err)
	}

	return "(integer) 1"
}

// TODO: 實現 PERSIST with filters
func (c *Client) PERSIST(cmd *command.Command) string {
	key := cmd.GetStr("key")

	c.server.mu.Lock()
	defer c.server.mu.Unlock()

	data := c.server.db[c.db]
	entry, isExist := data[key]
	if !isExist {
		return "(integer) 0"
	}

	if entry.ExpireAt != nil && time.Now().Unix() >= *entry.ExpireAt {
		c.server.delFromMem(c.db, key)
		return "(integer) 0"
	}

	if entry.ExpireAt == nil {
		return "(integer) 0"
	}

	entry.ExpireAt = nil

	if err := c.server.checkDB(c.db); err != nil {
		return fmt.Sprintf("Error creating writer: %v", err)
	}

	writer := c.server.writer[c.db]

	if err := writer.Write("SET", key, entry.Value); err != nil {
		return fmt.Sprintf("Error writing to AOF: %v", err)
	}

	reader := c.server.reader[c.db]
	existingCache, err := reader.Read(key)
	if err != nil {
		existingCache = &storage.Cache{
			Key:       key,
			Value:     entry.Value,
			Type:      entry.Type,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
			ExpireAt:  nil,
		}
	}

	cache := storage.Cache{
		Key:       key,
		Value:     entry.Value,
		Type:      entry.Type,
		CreatedAt: existingCache.CreatedAt,
		UpdatedAt: time.Now().Unix(),
		ExpireAt:  nil,
	}

	if err := writer.Save(key, cache); err != nil {
		return fmt.Sprintf("Error updating file: %v", err)
	}

	return "(integer) 1"
}

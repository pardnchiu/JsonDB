package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go-jsondb/internal/storage"
)

type Entry = storage.Entry

type Server struct {
	mu     sync.RWMutex
	db     map[int]map[string]*storage.Entry
	config storage.Config
	writer map[int]*storage.AOFWriter
	reader map[int]*storage.AOFReader
}

func NewServer() (*Server, error) {
	config := storage.NewConfig()

	dbList := make(map[int]map[string]*storage.Entry)
	writerList := make(map[int]*storage.AOFWriter)
	readerList := make(map[int]*storage.AOFReader)

	dbConfig := config
	dbConfig.DB = 0

	reader := storage.NewAOFReader(dbConfig)
	data, err := reader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load data from AOF(DB 0): %v", err)
	}

	dbList[0] = data
	readerList[0] = reader

	writer, err := storage.NewAOFWriter(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AOF writer(DB 0): %v", err)
	}
	writerList[0] = writer

	server := &Server{
		db:     dbList,
		config: config,
		writer: writerList,
		reader: readerList,
	}

	server.clean()

	return server, nil
}

func (s *Server) Close() error {
	for _, writer := range s.writer {
		if err := writer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) NewClient() *Client {
	return &Client{
		db:     0,
		server: s,
	}
}

func (c *Client) GetDB() int {
	return c.db
}

func (s *Server) clean() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 每1分鐘清理一次
		defer ticker.Stop()

		for range ticker.C {
			s.cleanExpire()
		}
	}()
}

func (s *Server) cleanExpire() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	total := 0

	for db, e := range s.db {
		count := 0
		var list []string

		for key, entry := range e {
			if entry.ExpireAt != nil && now >= *entry.ExpireAt {
				list = append(list, key)
			}
		}

		for _, key := range list {
			s.delFromMem(db, key)
			count++
		}

		if count > 0 {
			fmt.Printf("[TTL Clean] DB %d: cleaned %d expired keys\n", db, count)
			total += count
		}
	}

	if total > 0 {
		fmt.Printf("[TTL Clean] Total cleaned %d expired keys across all databases\n", total)
	}
}

func (s *Server) delFromMem(dbNum int, key string) {
	delete(s.db[dbNum], key)

	if writer, isExist := s.writer[dbNum]; isExist {
		writer.Write("DEL", key, nil)
	}

	if writer, isExist := s.writer[dbNum]; isExist {
		writer.Delete(key)
	}
}

func (s *Server) checkDB(db int) error {
	if _, isExist := s.db[db]; isExist {
		return nil
	}

	dbConfig := s.config
	dbConfig.DB = db

	// * 從 AOF 恢復數據
	reader := storage.NewAOFReader(dbConfig)
	data, err := reader.Load()
	if err != nil {
		return fmt.Errorf("failed to load data from AOF for DB %d: %v", db, err)
	}

	s.db[db] = data
	s.reader[db] = reader

	writer, err := storage.NewAOFWriter(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create AOF writer for DB %d: %v", db, err)
	}
	s.writer[db] = writer

	return nil
}

func (c *Client) matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if !strings.ContainsAny(pattern, "*?") {
		return key == pattern
	}

	return matchGlob(key, pattern, 0, 0)
}

func matchGlob(key, pattern string, keyIndex, patternIndex int) bool {
	if patternIndex == len(pattern) {
		return keyIndex == len(key)
	}

	if keyIndex == len(key) {
		for i := patternIndex; i < len(pattern); i++ {
			if pattern[i] != '*' {
				return false
			}
		}
		return true
	}

	switch pattern[patternIndex] {
	case '*':
		if matchGlob(key, pattern, keyIndex, patternIndex+1) {
			return true
		}
		return matchGlob(key, pattern, keyIndex+1, patternIndex)
	case '?':
		return matchGlob(key, pattern, keyIndex+1, patternIndex+1)
	default:
		if key[keyIndex] == pattern[patternIndex] {
			return matchGlob(key, pattern, keyIndex+1, patternIndex+1)
		}
		return false
	}
}

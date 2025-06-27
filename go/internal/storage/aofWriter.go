package storage

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AOF struct {
	Timestamp int64       `json:"timestamp"`
	Command   string      `json:"command"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value,omitempty"`
	Args      []string    `json:"args,omitempty"`
	ExpireAt  *int64      `json:"expire_at,omitempty"`
}

type AOFWriter struct {
	config Config
	file   *os.File
	mutex  sync.Mutex
	logger *slog.Logger
	count  int64
}

func NewAOFWriter(config Config) (*AOFWriter, error) {
	logger := slog.With("component", "AOF Writer")

	dir := filepath.Join(config.Option.DBPath, "aof")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create AOF directory: %v", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("db_%d.aof", config.DB))
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %v", err)
	}

	logger.Info("AOF file opened", "path", path)

	return &AOFWriter{
		config: config,
		file:   file,
		mutex:  sync.Mutex{},
		logger: logger,
	}, nil
}

func (w *AOFWriter) Write(command, key string, value interface{}, args ...string) error {
	return w.WriteWithTTL(command, key, value, nil, args...)
}

func (w *AOFWriter) WriteWithTTL(command, key string, value interface{}, ttlSeconds *uint64, args ...string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	aofCmd := AOF{
		Timestamp: time.Now().Unix(),
		Command:   command,
		Key:       key,
		Value:     value,
		Args:      args,
	}

	// 處理 TTL
	if ttlSeconds != nil {
		expireTime := time.Now().Unix() + int64(*ttlSeconds)
		aofCmd.ExpireAt = &expireTime
	}

	// 序列化為 JSON
	data, err := json.Marshal(aofCmd)
	if err != nil {
		return fmt.Errorf("failed to marshal AOF command: %v", err)
	}

	// 寫入文件
	_, err = w.file.WriteString(string(data) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write AOF command: %v", err)
	}

	// 強制刷新到磁盤
	return w.file.Sync()
}

func (w *AOFWriter) Save(key string, cache Cache) error {
	path := GetPath(w.config, key)

	if err := os.MkdirAll(path.FolderPath, 0755); err != nil {
		w.logger.Error("Failed to create folder", "path", path.FolderPath, "error", err)
		return fmt.Errorf("failed to create folder: %v", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		w.logger.Error("Failed to marshal cache data", "error", err)
		return fmt.Errorf("failed to marshal cache data: %v", err)
	}

	if err := os.WriteFile(path.Filepath, data, 0644); err != nil {
		w.logger.Error("Failed to write file", "path", path.Filepath, "error", err)
		return fmt.Errorf("failed to write file: %v", err)
	}

	w.logger.Info("Data written to file", "path", path.Filepath)
	return nil
}

func (w *AOFWriter) Delete(key string) error {
	path := GetPath(w.config, key)

	if err := os.Remove(path.Filepath); err != nil {
		if !os.IsNotExist(err) {
			w.logger.Error("Failed to delete file", "path", path.Filepath, "error", err)
			return fmt.Errorf("failed to delete file: %v", err)
		}
	}

	w.logger.Info("File deleted", "path", path.Filepath)
	return nil
}

func (w *AOFWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		w.logger.Info("Closing AOF file")
		return w.file.Close()
	}
	return nil
}

package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go-jsondb/internal/util"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type AOFReader struct {
	config Config
	logger *slog.Logger
}

type Entry struct {
	Value    string `json:"value"`
	Type     string `json:"type"`
	ExpireAt *int64 `json:"expire_at,omitempty"`
}

func NewAOFReader(config Config) *AOFReader {
	logger := slog.With("component", "AOF Writer")

	return &AOFReader{
		config: config,
		logger: logger,
	}
}

func (r *AOFReader) Load() (map[string]*Entry, error) {
	data := make(map[string]*Entry)

	dir := filepath.Join(r.config.Option.DBPath, "aof")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create AOF directory: %v", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("db_%d.aof", r.config.DB))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		r.logger.Info("AOF file not found, starting with empty database", "path", path)
		return data, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %v", err)
	}
	defer file.Close()

	r.logger.Info("Loading data from AOF file", "path", path)

	scanner := bufio.NewScanner(file)
	count := 0

	for scanner.Scan() {
		count++
		line := scanner.Text()

		if line == "" {
			continue
		}

		var cmd AOF
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			r.logger.Error("Failed to parse AOF line", "line", count, "error", err)
			continue
		}

		// * 依據 AOF 恢復資料
		switch cmd.Command {
		case "SET":
			if value, ok := cmd.Value.(string); ok {
				entry := &Entry{
					Value: value,
					Type:  util.GetType(value),
				}

				if cmd.ExpireAt != nil {
					if time.Now().Unix() < *cmd.ExpireAt {
						entry.ExpireAt = cmd.ExpireAt
					} else {
						continue
					}
				}

				data[cmd.Key] = entry
			}
		case "DEL":
			delete(data, cmd.Key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading AOF file: %v", err)
	}

	r.logger.Info("Loaded keys from AOF file", "count", len(data))
	return data, nil
}

func (r *AOFReader) Read(key string) (*Cache, error) {
	path := GetPath(r.config, key)

	if _, err := os.Stat(path.Filepath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path.Filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %v", err)
	}

	return &cache, nil
}

package storage

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"strconv"
)

type Config struct {
	Option Option `json:"option"`
	DB     int    `json:"db"`
}

type Option struct {
	DBPath string `json:"db_path"`
}

type Path struct {
	FolderPath string `json:"folder_path"`
	Filepath   string `json:"filepath"`
	Filename   string `json:"filename"`
}

type Cache struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Type      string      `json:"type"`
	CreatedAt int64       `json:"created_at"`
	UpdatedAt int64       `json:"updated_at"`
	ExpireAt  *int64      `json:"expire_at,omitempty"`
}

func NewConfig() Config {
	return Config{
		Option: Option{
			DBPath: "./data",
		},
		DB: 0,
	}
}

func GetPath(config Config, key string) Path {
	hash := md5.Sum([]byte(key))
	encode := fmt.Sprintf("%x", hash)

	layer1 := encode[0:2]
	layer2 := encode[2:4]
	layer3 := encode[4:6]
	filename := encode + ".json"

	// 構建完整路徑: data/0/ab/cd/ef/abcdef....json
	folderPath := filepath.Join(
		config.Option.DBPath,
		strconv.Itoa(config.DB),
		layer1,
		layer2,
		layer3,
	)

	return Path{
		FolderPath: folderPath,
		Filepath:   filepath.Join(folderPath, filename),
		Filename:   filename,
	}
}

package handler

import (
	"bucket-serve/cache"
	"encoding/json"
	"os"
)

type Config struct {
	IndexPath             string            `json:"indexPath"`
	IsSPA                 bool              `json:"isSPA"`
	ContentFileExt        string            `json:"contentFileExt"`
	CacheFileExtInBrowser []string          `json:"cacheFileExtInBrowser"`
	CacheConfig           cache.CacheConfig `json:"cacheConfig"`
}

func ReadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

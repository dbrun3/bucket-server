package cache

import (
	"encoding/json"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

type CacheConfig struct {
	DefaultExpiry   Duration `json:"defaultExpiry"`
	StaleInterval   Duration `json:"staleInterval"`
	CleanupInterval Duration `json:"cleanupInterval"`
	DisableCache    bool     `json:"disableCache"`
}

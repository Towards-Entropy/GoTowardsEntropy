/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"encoding/json"
	"os"
	"sync"
)

type Config struct {
	CompressionLevel    *int               // Effort setting for zstd
	BufferSize          *int               // Size of buffer for streaming (de)compression
	DictionaryDirectory *string            // Directory to load dictionaries from
	PreflightWrites     *bool              // Whether to preflight writes
	HandleHeadRequests  *bool              // Whether to forward HEAD requests to underlying handler
	DictionaryMatchMap  *map[string]string // Map of request url match strings to dictionary ids
	LogLevel            *LogLevel          // Log level
}

type internalConfig struct {
	CompressionLevel    int               // Effort setting for zstd
	BufferSize          int               // Size of buffer for streaming (de)compression
	DictionaryDirectory string            // Directory to load dictionaries from
	PreflightWrites     bool              // Whether to preflight writes
	HandleHeadRequests  bool              // Whether to forward HEAD requests to underlying handler
	DictionaryMatchMap  map[string]string // Map of request url match strings to dictionary ids
	LogLevel            LogLevel          // Log level
}

type CompressionType string

const (
	Zstd       CompressionType = "zstd"
	SharedZstd CompressionType = "szstd"
)

type LogLevel int

const (
	LogLevelNone LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

var (
	currentConfig internalConfig
	mu            sync.RWMutex
)

var defaultConfig = Config{
	CompressionLevel:    IntPtr(5),
	BufferSize:          IntPtr(1024),
	DictionaryDirectory: StrPtr("./dictionaries"),
	PreflightWrites:     BoolPtr(true),
	HandleHeadRequests:  BoolPtr(true),
	DictionaryMatchMap:  MapPtr(make(map[string]string)),
}

func IntPtr(i int) *int                             { return &i }
func StrPtr(s string) *string                       { return &s }
func BoolPtr(b bool) *bool                          { return &b }
func MapPtr(m map[string]string) *map[string]string { return &m }
func LogLevelPtr(l LogLevel) *LogLevel              { return &l }

// Sets default values.
func init() {
	InitWithStruct(defaultConfig)
}

// Call Init after setting config values.
func InitWithStruct(cfg Config) {
	setConfig(cfg)
}

// SetConfig updates the current configuration.
func setConfig(cfg Config) {
	mu.Lock()
	defer mu.Unlock()

	if cfg.CompressionLevel != nil {
		currentConfig.CompressionLevel = *cfg.CompressionLevel
	}
	if cfg.BufferSize != nil {
		currentConfig.BufferSize = *cfg.BufferSize
	}
	if cfg.DictionaryDirectory != nil {
		if currentConfig.DictionaryDirectory != *cfg.DictionaryDirectory {
			updateCacheFromDir(*cfg.DictionaryDirectory)
		}
		currentConfig.DictionaryDirectory = *cfg.DictionaryDirectory
	}
	if cfg.PreflightWrites != nil {
		currentConfig.PreflightWrites = *cfg.PreflightWrites
	}
	if cfg.HandleHeadRequests != nil {
		currentConfig.HandleHeadRequests = *cfg.HandleHeadRequests
	}
	if cfg.DictionaryMatchMap != nil {
		currentConfig.DictionaryMatchMap = *cfg.DictionaryMatchMap
	}
	if cfg.LogLevel != nil {
		currentConfig.LogLevel = *cfg.LogLevel
	}
}

// GetConfig returns the current configuration.
func getConfig() internalConfig {
	mu.RLock()
	defer mu.RUnlock()

	return currentConfig
}

// SetConfigFromJsonFile reads the JSON content from the specified file and updates the configuration.
func setConfigFromJsonFile(path string) error {
	// Read the content from the specified file path.
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the content into the Config structure.
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Set the configuration.
	setConfig(cfg)
	return nil
}

func (c *internalConfig) getInvertedMatchMap() map[string]string {
	inverted := make(map[string]string)
	for k, v := range c.DictionaryMatchMap {
		inverted[v] = k
	}
	return inverted
}

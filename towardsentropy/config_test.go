/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"reflect"
	"testing"
)

func TestConfigFromJsonFile(t *testing.T) {
	setConfigFromJsonFile("../testdata/config.json")
	config := getConfig()
	if config.CompressionLevel != 2 {
		t.Fatalf("Expected compression level 2, got %d", config.CompressionLevel)
	}
	if config.BufferSize != 2048 {
		t.Fatalf("Expected buffer size 2048, got %d", config.BufferSize)
	}
	if config.DictionaryDirectory != "../testdata/dictionaries" {
		t.Fatalf("Expected dictionary directory '../testdata/dictionaries', got %s", config.DictionaryDirectory)
	}
	// Confirm that dictionary cache got updated
	dict := getDictionary("enwik8")
	if dict == nil {
		t.Fatalf("Expected dictionary 'enwik8' to be loaded")
	}
}

func TestMissingConfig(t *testing.T) {
	setConfig(defaultConfig)
	// Should safely _not_ update config
	setConfigFromJsonFile("../testdata/config.notreal.json")
	config := getConfig()
	setConfig(defaultConfig)
	config2 := getConfig()
	if !areConfigsEqual(&config, &config2) {
		t.Fatalf("Expected default config, got %v", config)
	}
}

func areConfigsEqual(a, b *internalConfig) bool {
	return reflect.DeepEqual(a, b)
}

/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	data := "This is a sample string that we are going to compress and then decompress."
	reader := strings.NewReader(data)
	var compressed bytes.Buffer

	// Compress the data
	err := Compress(reader, &compressed, "")
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	// Decompress the data
	var decompressed bytes.Buffer
	err = Decompress(&compressed, &decompressed, "")
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Compare the original and decompressed data
	if data != decompressed.String() {
		t.Fatalf("Original and decompressed data do not match. Got: %s", decompressed.String())
	}
}

func TestCompressDecompressWithDictionary(t *testing.T) {
	updateCacheFromDir("../testdata/dictionaries")
	data := "This is a sample string that we are going to compress and then decompress."
	reader := strings.NewReader(data)
	var compressed bytes.Buffer

	// Compress the data with a dictionary
	err := Compress(reader, &compressed, "enwik8")
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}

	// Decompress the data with the same dictionary
	var decompressed bytes.Buffer
	err = Decompress(&compressed, &decompressed, "enwik8")
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Compare the original and decompressed data
	if data != decompressed.String() {
		t.Fatalf("Original and decompressed data do not match. Got: %s", decompressed.String())
	}
}

func TestCompressDecompressFile(t *testing.T) {
	data := "Another sample string for file-based compression and decompression."
	var compressed bytes.Buffer

	// Compress the data as if it's from a file
	err := CompressFile([]byte(data), &compressed, "")
	if err != nil {
		t.Fatalf("Failed to compress file data: %v", err)
	}

	// Decompress the data as if it's for a file
	var decompressed bytes.Buffer
	err = DecompressFile(compressed.Bytes(), &decompressed, "")
	if err != nil {
		t.Fatalf("Failed to decompress file data: %v", err)
	}

	// Compare the original and decompressed data
	if data != decompressed.String() {
		t.Fatalf("Original and decompressed file data do not match. Got: %s", decompressed.String())
	}
}

func TestInvalidDecompress(t *testing.T) {
	// Trying to decompress non-zstd compressed data should result in an error
	data := "This is not compressed using zstd."
	reader := bytes.NewReader([]byte(data))
	var decompressed bytes.Buffer

	err := Decompress(reader, &decompressed, "")
	if err == nil {
		t.Fatalf("Expected error when trying to decompress invalid data, got nil.")
	}
}

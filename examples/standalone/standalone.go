/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Towards-Entropy/GoTowardsEntropy/towardsentropy"
)

func main() {
	cfg := towardsentropy.Config{
		DictionaryDirectory: towardsentropy.StrPtr("../../testdata/dictionaries"),
	}
	towardsentropy.InitWithStruct(cfg)
	files := listFiles("../../testdata/files/supply_chain")
	for _, file := range files {
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		fmt.Printf("File: %s\n", filepath.Base(file))
		Compress(fileBytes)
	}
}

func Compress(b []byte) {
	var dictCompressed bytes.Buffer
	var nonDictCompressed bytes.Buffer
	dictReader := bytes.NewReader(b)
	nonDictReader := bytes.NewReader(b)

	// Compress the data with a dictionary
	err := towardsentropy.Compress(dictReader, &dictCompressed, "supply_chain")
	if err != nil {
		log.Fatalf("Failed to compress data: %v", err)
	}

	// Compress the data without a dictionary
	err = towardsentropy.Compress(nonDictReader, &nonDictCompressed, "")
	if err != nil {
		log.Fatalf("Failed to compress data: %v", err)
	}

	fmt.Printf("Original size:                         %d\n", len(b))
	fmt.Printf("Compressed size with dictionary:       %d\n", len(dictCompressed.Bytes()))
	fmt.Printf("Compressed size without dictionary:    %d\n", len(nonDictCompressed.Bytes()))
	fmt.Printf("Compression ratio with dictionary:     %.2f\n", float64(len(dictCompressed.Bytes()))/float64(len(b)))
	fmt.Printf("Compression ratio without dictionary:  %.2f\n", float64(len(nonDictCompressed.Bytes()))/float64(len(b)))
}

// ListFiles returns a slice of file names from the specified directory path.
func listFiles(dir string) []string {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// If it's a file, add to the list
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

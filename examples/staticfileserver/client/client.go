/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Towards-Entropy/GoTowardsEntropy/towardsentropy"
)

func main() {
	cfg := towardsentropy.Config{
		CompressionLevel:    towardsentropy.IntPtr(5),
		BufferSize:          towardsentropy.IntPtr(1024),
		DictionaryDirectory: towardsentropy.StrPtr("../../../testdata/dictionaries"),
	}
	towardsentropy.InitWithStruct(cfg)
	transport := towardsentropy.NewTowardsEntropyTransport(http.DefaultTransport)
	client := &http.Client{Transport: transport}

	files, err := listFiles("../../../testdata/files")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		requestFile(file, client)
	}
}

func requestFile(file string, client *http.Client) {
	log.Printf("Requesting %s", file)
	filename := filepath.Base(file)
	dir := filepath.Base(filepath.Dir(file))
	resp, err := client.Get(fmt.Sprintf("http://localhost:8080/%s/%s", dir, filename))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	bodyStr := string(bodyBytes)
	if len(bodyStr) > 20 {
		bodyStr = bodyStr[:20]
	}
	if strings.Contains(filename, "enwik") && bodyStr != `<mediawiki xmlns="ht` {
		fmt.Printf("Response body: %s\n", bodyStr)
		panic("Response body is not as expected")
	}
	log.Printf("%s OK", file)
}

// ListFiles returns a slice of file names from the specified directory path.
func listFiles(dir string) ([]string, error) {
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
		return nil, err
	}

	return files, nil
}

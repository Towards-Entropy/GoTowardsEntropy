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
	"math/rand"
	"net/http"
	"os"

	"github.com/Towards-Entropy/GoTowardsEntropy/towardsentropy"
)

func getRandomFileFromDir(directory string) ([]byte, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory: %s", directory)
	}

	randomFile := files[rand.Intn(len(files))]

	fileContent, err := os.ReadFile(directory + "/" + randomFile.Name())
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}

func main() {
	cfg := towardsentropy.Config{
		DictionaryDirectory: towardsentropy.StrPtr("../../../testdata/dictionaries"),
		DictionaryMatchMap:  towardsentropy.MapPtr(map[string]string{"*": "supply_chain"}),
	}
	towardsentropy.InitWithStruct(cfg)
	transport := towardsentropy.NewTowardsEntropyTransport(http.DefaultTransport)
	client := &http.Client{Transport: transport}
	fileContent, err := getRandomFileFromDir("../../../testdata/files/supply_chain")
	if err != nil {
		log.Fatalf("Failed to get a random file: %s", err)
	}

	// Make a POST request with the randomly selected file's content
	resp, err := client.Post("http://localhost:8080", "application/octet-stream", bytes.NewReader(fileContent))
	if err != nil {
		log.Fatalf("Failed to make POST request: %s", err)
	}
	defer resp.Body.Close()

	// Optionally, print the response status
	fmt.Println("Response status:", resp.Status)
}

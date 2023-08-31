/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"io"
	"log"
	"net/http"

	"github.com/Towards-Entropy/GoTowardsEntropy/towardsentropy"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Check if request method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Only POST method is allowed"))
		return
	}

	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if len(bodyBytes) == 0 {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	log.Printf("OK")
}

func main() {
	// Initialize compression middleware
	cfg := towardsentropy.Config{
		DictionaryDirectory: towardsentropy.StrPtr("../../../testdata/dictionaries"),
		DictionaryMatchMap:  towardsentropy.MapPtr(map[string]string{"*": "supply_chain"}),
	}
	towardsentropy.InitWithStruct(cfg)

	// Apply the compression middleware to our handler
	towardsentropyHandler := towardsentropy.NewTowardsEntropyHandler(http.HandlerFunc(handler))

	http.Handle("/", towardsentropyHandler)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

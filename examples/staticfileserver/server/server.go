/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"fmt"
	"net/http"

	"github.com/Towards-Entropy/GoTowardsEntropy/towardsentropy"
)

func main() {
	fileServer := http.FileServer(http.Dir("../../../testdata/files"))

	// Wrap the file server with the compressing middleware
	cfg := towardsentropy.Config{
		DictionaryDirectory: towardsentropy.StrPtr("../../../testdata/dictionaries"),
		DictionaryMatchMap:  towardsentropy.MapPtr(map[string]string{"*": "supply_chain"}),
	}
	towardsentropy.InitWithStruct(cfg)
	compressedFileServer := towardsentropy.NewTowardsEntropyHandler(fileServer)

	http.Handle("/", compressedFileServer)

	fmt.Println("Serving files on :8080 from ../../../testdata/files directory")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

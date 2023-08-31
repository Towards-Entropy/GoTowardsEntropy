/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTPBasic(t *testing.T) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{"*": "supply_chain"}),
	})
	handler := NewTowardsEntropyHandler(baseHandler)

	rr := executeRequest(handler, "GET", "/test", []string{"zstd"}, []string{}, t)

	checkStatus(rr, http.StatusOK, t)
	checkHeader(rr, "Content-Encoding", string(Zstd), t)
	checkHeader(rr, "Dictionary-Id", "", t)
	checkBody("", rr, "OK", t)
}

func TestServeHTTPEncodingNoAvailableDictionary(t *testing.T) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{"*": "supply_chain"}),
	})
	handler := NewTowardsEntropyHandler(baseHandler)

	rr := executeRequest(handler, "GET", "/test", []string{"zstd", "szstd"}, []string{}, t)

	checkStatus(rr, http.StatusOK, t)
	checkHeader(rr, "Content-Encoding", string(Zstd), t)
	checkHeader(rr, "Dictionary-Id", "", t)
	checkBody("", rr, "OK", t)
}

func TestServeHTTPDictionary(t *testing.T) {
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{"*": "supply_chain"}),
	})
	handler := NewTowardsEntropyHandler(baseHandler)

	rr := executeRequest(handler, "GET", "/test", []string{"zstd", "szstd"}, []string{"supply_chain"}, t)

	checkStatus(rr, http.StatusOK, t)
	checkHeader(rr, "Content-Encoding", string(SharedZstd), t)
	checkHeader(rr, "Dictionary-Id", "supply_chain", t)
	checkBody("supply_chain", rr, "OK", t)

	// Confirm that HEAD requests work for preflight handling
	rr = executeRequest(handler, "HEAD", "/test", []string{"zstd", "szstd"}, []string{"supply_chain"}, t)

	checkStatus(rr, http.StatusOK, t)
	checkHeader(rr, "Content-Encoding", string(SharedZstd), t)
	checkHeader(rr, "Dictionary-Id", "supply_chain", t)
	checkBody("supply_chain", rr, "", t)
}

func executeRequest(
	handler http.Handler,
	method, path string,
	acceptEncoding []string,
	availableDictionary []string,
	t *testing.T,
) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("Could not create HTTP request: %v", err)
	}
	for _, encoding := range acceptEncoding {
		req.Header.Add("Accept-Encoding", encoding)
	}
	for _, dictionary := range availableDictionary {
		req.Header.Add("Available-Dictionary", dictionary)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func checkStatus(rr *httptest.ResponseRecorder, expected int, t *testing.T) {
	if status := rr.Code; status != expected {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, expected)
	}
}

func checkHeader(rr *httptest.ResponseRecorder, headerName, expected string, t *testing.T) {
	if value := rr.Header().Get(headerName); value != expected {
		t.Errorf("Handler returned wrong %s: got %v, expected %v", headerName, value, expected)
	}
}

func checkBody(dictionaryId string, rr *httptest.ResponseRecorder, expected string, t *testing.T) {
	var bodyBuffer bytes.Buffer
	err := Decompress(rr.Body, &bodyBuffer, dictionaryId)
	if err != nil {
		t.Errorf("Error decompressing body: %v", err)
	}
	bodyString := bodyBuffer.String()
	if bodyString != expected {
		t.Errorf("Handler returned wrong body: got %v, expected %v", bodyString, expected)
	}
}

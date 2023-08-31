/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

type MockRoundTripper struct {
	expectedBody      []byte
	expectedHeaders   *map[string]string
	preflightResponse *http.Response
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("OK")),
	}

	if m.preflightResponse != nil {
		if req.Method == http.MethodHead {
			return m.preflightResponse, nil
		}
		m.preflightResponse = nil
	}

	if req.Body == nil && m.expectedBody == nil {
		return resp, nil
	}

	if req.Body == nil {
		return nil, fmt.Errorf("Missing body in request")
	}

	if m.expectedBody == nil {
		return nil, fmt.Errorf("Unexpected body in request")
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if !areSlicesEqual(body, m.expectedBody) {
		return nil, fmt.Errorf("unexpected body")
	}

	if m.expectedHeaders != nil {
		for k, v := range *m.expectedHeaders {
			if req.Header.Get(k) != v {
				return nil, fmt.Errorf("unexpected header %s: %s", k, req.Header.Get(k))
			}
		}
	}

	return resp, nil
}

func TestTowardsEntropyTransportNoBody(t *testing.T) {
	base := &MockRoundTripper{}
	InitWithStruct(Config{})
	transport := NewTowardsEntropyTransport(base)

	testCases := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodHead},
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodPatch},
		{http.MethodDelete},
		{http.MethodOptions},
	}

	for _, tc := range testCases {
		log.Printf("Testing method %s", tc.method)
		req, err := http.NewRequest(tc.method, "http://example.com", nil)
		if err != nil {
			t.Fatalf("Could not create HTTP request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status code for method %s: got %v, expected %v", tc.method, resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("Unexpected body for method %s: got %s, expected %s", tc.method, body, "OK")
		}
	}
}

func TestTransportBodyNoDict(t *testing.T) {
	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{}),
		LogLevel:            LogLevelPtr(LogLevelDebug),
		PreflightWrites:     BoolPtr(false),
	})

	body := getBody()
	var compressedBuffer bytes.Buffer
	err := Compress(bytes.NewReader(body), &compressedBuffer, "")
	if err != nil {
		t.Fatalf("Could not compress body: %v", err)
	}
	base := &MockRoundTripper{
		expectedBody:    compressedBuffer.Bytes(),
		expectedHeaders: &map[string]string{"DictionaryId": "", "Content-Encoding": "zstd"},
	}

	transport := NewTowardsEntropyTransport(base)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodPatch},
	}

	for _, tc := range testCases {
		log.Printf("Testing method %s", tc.method)
		req, err := http.NewRequest(tc.method, "http://example.com", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Could not create HTTP request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status code for method %s: got %v, expected %v", tc.method, resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("Unexpected body for method %s", tc.method)
		}
	}
}

func TestTransportBodyDict(t *testing.T) {
	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{"*": "supply_chain"}),
		PreflightWrites:     BoolPtr(false),
	})

	body := getBody()
	var compressedBuffer bytes.Buffer
	err := Compress(bytes.NewReader(body), &compressedBuffer, "supply_chain")
	if err != nil {
		t.Fatalf("Could not compress body: %v", err)
	}
	base := &MockRoundTripper{
		expectedBody:    compressedBuffer.Bytes(),
		expectedHeaders: &map[string]string{"Dictionary-Id": "supply_chain", "Content-Encoding": "szstd"},
	}

	transport := NewTowardsEntropyTransport(base)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodPatch},
	}

	for _, tc := range testCases {
		log.Printf("Testing method %s", tc.method)
		req, err := http.NewRequest(tc.method, "http://example.com", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Could not create HTTP request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status code for method %s: got %v, expected %v", tc.method, resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("Unexpected body for method %s", tc.method)
		}
	}
}

func TestTransportBodyDictPreflight(t *testing.T) {
	InitWithStruct(Config{
		DictionaryDirectory: StrPtr("../testdata/dictionaries"),
		DictionaryMatchMap:  MapPtr(map[string]string{"*": "supply_chain"}),
		PreflightWrites:     BoolPtr(true),
	})

	body := getBody()
	var compressedBuffer bytes.Buffer
	err := Compress(bytes.NewReader(body), &compressedBuffer, "supply_chain")
	if err != nil {
		t.Fatalf("Could not compress body: %v", err)
	}
	preflightResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Dictionary-Id": []string{"supply_chain"}},
	}
	base := &MockRoundTripper{
		expectedBody:    compressedBuffer.Bytes(),
		expectedHeaders: &map[string]string{"Dictionary-Id": "supply_chain", "Content-Encoding": "szstd"},
	}

	transport := NewTowardsEntropyTransport(base)

	testCases := []struct {
		method string
	}{
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodPatch},
	}

	for _, tc := range testCases {
		base.preflightResponse = preflightResponse
		log.Printf("Testing method %s", tc.method)
		req, err := http.NewRequest(tc.method, "http://example.com", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Could not create HTTP request: %v", err)
		}

		resp, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Unexpected status code for method %s: got %v, expected %v", tc.method, resp.StatusCode, http.StatusOK)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("Unexpected body for method %s", tc.method)
		}

		if base.preflightResponse != nil {
			t.Errorf("Preflight response not consumed")
		}
	}
}

func getBody() []byte {
	content, err := os.ReadFile("../testdata/files/supply_chain/SupplyChainGHGEmissionFactors_v1.2_NAICS_byGHG_USD2021_chunk_9.csv")
	if err != nil {
		log.Fatalf("Could not read file: %v", err)
	}

	return content
}

func areSlicesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		log.Printf("Lengths are not equal: %d != %d", len(a), len(b))
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

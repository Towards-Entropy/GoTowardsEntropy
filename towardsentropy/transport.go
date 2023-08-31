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
	"net/http"

	"github.com/DataDog/zstd"
)

var errNoDictionaryFound = fmt.Errorf("no dictionary found")

type TowardsEntropyTransport struct {
	base   http.RoundTripper
	config internalConfig
	logger Logger
}

func NewTowardsEntropyTransport(base http.RoundTripper) *TowardsEntropyTransport {
	config := getConfig()
	if base == nil {
		base = http.DefaultTransport
	}
	return &TowardsEntropyTransport{
		base:   base,
		config: config,
		logger: Logger{config.LogLevel},
	}
}

func (t *TowardsEntropyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		return t.roundTripRead(req)
	} else if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		return t.roundTripWrite(req)
	} else {
		return t.base.RoundTrip(req)
	}
}

func (t *TowardsEntropyTransport) roundTripRead(req *http.Request) (*http.Response, error) {
	t.addReadHeaders(req)
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body = t.newDecompressedReader(resp)
	return resp, nil
}

func (t *TowardsEntropyTransport) addReadHeaders(req *http.Request) {
	dictionaries := findMatchingDictionaries(req, &t.config)
	if len(dictionaries) == 0 {
		t.addReadNonDictionaryHeaders(req)
		return
	}

	req.Header.Add("Accept-Encoding", string(SharedZstd))
	req.Header.Add("Accept-Encoding", string(Zstd))
	for _, dictionaryId := range dictionaries {
		req.Header.Add("Available-Dictionary", dictionaryId)
	}
}

func (t *TowardsEntropyTransport) addReadNonDictionaryHeaders(req *http.Request) *Dictionary {
	req.Header.Set("Accept-Encoding", string(Zstd))
	return nil
}

func (t *TowardsEntropyTransport) newDecompressedReader(resp *http.Response) io.ReadCloser {
	encoding := resp.Header.Get("Content-Encoding")
	if encoding == string(Zstd) {
		return zstd.NewReader(resp.Body)
	} else if encoding == string(SharedZstd) {
		dictionaryId := resp.Header.Get("Dictionary-Id")
		dictionary := getDictionary(dictionaryId)
		if dictionary == nil {
			t.logger.Error("No dictionary found for response")
			// TODO error handle, this would be BAD!
			return zstd.NewReader(resp.Body)
		}
		t.logger.Debugf("Using dictionary %s", dictionary.Id)
		return zstd.NewReaderDict(resp.Body, dictionary.Bytes)
	} else {
		return resp.Body
	}
}

func (t *TowardsEntropyTransport) roundTripWrite(req *http.Request) (*http.Response, error) {
	if req.Body == nil {
		t.logger.Debug("No body in write request, skipping compression")
		return t.base.RoundTrip(req)
	}

	dictionaryId, err := t.getDictionaryId(req)
	if err != nil && err != errNoDictionaryFound {
		t.logger.Errorf("Error getting dictionary id: %v", err)
		return nil, err
	}
	dictionary := getDictionary(dictionaryId)

	var compressedBuffer bytes.Buffer
	t.logger.Debugf("Compressing request with dictionary '%s'", dictionaryId)
	err = t.compress(req.Body, &compressedBuffer, dictionary)
	if err != nil {
		// TODO consider error cases here
		return nil, err
	}

	req.Body = io.NopCloser(bytes.NewReader(compressedBuffer.Bytes()))
	if (dictionaryId) == "" {
		req.Header.Set("Content-Encoding", string(Zstd))
	} else {
		req.Header.Set("Dictionary-Id", dictionaryId)
		req.Header.Set("Content-Encoding", string(SharedZstd))
	}
	req.ContentLength = int64(compressedBuffer.Len())
	t.logger.Debugf("Making request with content length %d", req.ContentLength)
	return t.base.RoundTrip(req)
}

func (t *TowardsEntropyTransport) getDictionaryId(req *http.Request) (string, error) {
	if t.requiresPreflight(req) {
		return t.getDictionaryIdViaPreflight(req)
	}
	return t.getDictionaryIdUnsafe(req)
}

func (t *TowardsEntropyTransport) requiresPreflight(req *http.Request) bool {
	if !t.config.PreflightWrites {
		return false
	}

	return req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch
}

func (t *TowardsEntropyTransport) getDictionaryIdViaPreflight(req *http.Request) (string, error) {
	headReq, err := http.NewRequest(http.MethodHead, req.URL.String(), nil)
	if err != nil {
		return "", err
	}

	// Copy headers from original request to HEAD request.
	for key, values := range req.Header {
		for _, value := range values {
			headReq.Header.Add(key, value)
		}
	}

	resp, err := t.base.RoundTrip(headReq)
	if err != nil {
		return "", err
	}
	if resp.Header.Get("Dictionary-Id") == "" {
		return "", errNoDictionaryFound
	}
	dictionaryId := resp.Header.Get("Dictionary-Id")
	return dictionaryId, nil
}

func (t *TowardsEntropyTransport) getDictionaryIdUnsafe(req *http.Request) (string, error) {
	dictionaries := findMatchingDictionaries(req, &t.config)
	if len(dictionaries) == 0 {
		t.logger.Debug("No matching dictionaries found for request")
		return "", errNoDictionaryFound
	}
	return dictionaries[0], nil
}

func findMatchingDictionaries(req *http.Request, config *internalConfig) []string {
	fullURL := req.URL.String()
	if !req.URL.IsAbs() && req.URL.Host == "" {
		fullURL = req.Host + req.URL.String()
	}

	matchMap := config.DictionaryMatchMap
	matchingIds := make([]string, 0)
	for pattern, id := range matchMap {
		if pattern == "" {
			continue
		}
		if matches(pattern, fullURL) {
			matchingIds = append(matchingIds, id)
		}
	}
	return matchingIds
}

func (t *TowardsEntropyTransport) compress(r io.Reader, w io.Writer, dict *Dictionary) error {
	config := getConfig()

	var zw *zstd.Writer
	if dict == nil {
		t.logger.Debug("Compressing request with no dictionary")
		zw = zstd.NewWriterLevel(w, config.CompressionLevel)
	} else {
		t.logger.Debugf("Compressing request with dictionary '%s'", dict.Id)
		zw = zstd.NewWriterLevelDict(w, config.CompressionLevel, dict.Bytes)
	}
	defer zw.Close()
	return t.streamCompress(r, zw)
}

func (t *TowardsEntropyTransport) streamCompress(r io.Reader, zw *zstd.Writer) error {
	t.logger.Debug("Streaming compressing request")
	buf := make([]byte, t.config.BufferSize)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading response body: %v", err)
		}

		_, err = zw.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("error compressing and writing data: %v", err)
		}
	}

	// Flush any unwritten data to the underlying writer
	err := zw.Flush()
	if err != nil {
		return fmt.Errorf("error flushing remaining data: %v", err)
	}

	t.logger.Debug("Done streaming compression")
	return nil
}

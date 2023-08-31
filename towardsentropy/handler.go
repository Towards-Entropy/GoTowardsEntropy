/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"log"
	"net/http"
	"strings"

	"github.com/DataDog/zstd"
)

// zstdResponseWriter is an http.ResponseWriter that writes response with zstd.
type TowardsEntropyHandler struct {
	baseHandler http.Handler
	config      internalConfig
	logger      Logger
}

func NewTowardsEntropyHandler(baseHandler http.Handler) *TowardsEntropyHandler {
	config := getConfig()
	return &TowardsEntropyHandler{
		baseHandler: baseHandler,
		config:      config,
		logger:      Logger{config.LogLevel},
	}
}

func (h *TowardsEntropyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.maybeDecompressRequest(r)
	dictionary := h.selectDictionaryFromRequest(r)
	h.handleWithDictionary(w, r, dictionary)
}

func (h *TowardsEntropyHandler) maybeDecompressRequest(r *http.Request) {
	if (r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch) || r.Body == nil {
		return
	}

	encoding := r.Header.Get("Content-Encoding")
	if encoding == string(Zstd) {
		r.Body = zstd.NewReader(r.Body)
	} else if encoding == string(SharedZstd) {
		dictionaryId := r.Header.Get("Dictionary-Id")
		dictionary := getDictionary(dictionaryId)
		if dictionary == nil {
			// TODO error handle, this would be BAD!
			log.Fatalf("No dictionary found for request")
		}
		r.Body = zstd.NewReaderDict(r.Body, dictionary.Bytes)
	}
}

func (h *TowardsEntropyHandler) selectDictionaryFromRequest(req *http.Request) *Dictionary {
	if !contains(req.Header.Values("Accept-Encoding"), string(SharedZstd)) {
		h.logger.Debug("Client does not accept shared dictionary")
		return nil
	}

	// Shortcut if client forces dictionary
	if req.Header.Get("Dictionary-Id") != "" {
		h.logger.Debugf("Client forces dictionary: %s", req.Header.Get("Dictionary-Id"))
		dictionaryId := req.Header.Get("Dictionary-Id")
		return getDictionary(dictionaryId)
	}

	dictionaryIds := req.Header.Values("Available-Dictionary")
	for i, id := range dictionaryIds {
		dictionaryIds[i] = strings.TrimSpace(id)
	}
	filteredDictionaryIds := h.getMatchingDictionaries(req, dictionaryIds)
	return findDictionary(filteredDictionaryIds)
}

func (h *TowardsEntropyHandler) handleWithDictionary(w http.ResponseWriter, r *http.Request, dict *Dictionary) {
	if r.Method == http.MethodHead && h.config.HandleHeadRequests {
		h.handleHeadRequest(w, r, dict)
		return
	}

	var zw *zstd.Writer
	if dict == nil {
		h.logger.Debug("Compressing response with no dictionary")
		zw = zstd.NewWriterLevel(w, 5)
		w.Header().Set("Content-Encoding", string(Zstd))
	} else {
		h.logger.Debugf("Compressing response with dictionary %s", dict.Id)
		zw = zstd.NewWriterLevelDict(w, 5, dict.Bytes)
		w.Header().Set("Content-Encoding", string(SharedZstd))
		w.Header().Set("Dictionary-Id", dict.Id)
	}
	defer zw.Close()

	zstdResponseWriter := &zstdResponseWriter{
		ResponseWriter: w,
		Writer:         zw,
	}
	h.baseHandler.ServeHTTP(zstdResponseWriter, r)
}

func (h *TowardsEntropyHandler) handleHeadRequest(w http.ResponseWriter, r *http.Request, dictionary *Dictionary) {
	if dictionary != nil {
		h.logger.Debugf("Handling HEAD request with dictionary %s", dictionary.Id)
		w.Header().Set("Dictionary-Id", dictionary.Id)
		w.Header().Set("Content-Encoding", string(SharedZstd))
		w.WriteHeader(http.StatusOK)
	} else {
		h.logger.Debug("Handling HEAD request with no dictionary")
		w.Header().Set("Content-Encoding", string(Zstd))
		w.WriteHeader(http.StatusOK)
	}
}

func (h *TowardsEntropyHandler) getMatchingDictionaries(req *http.Request, dictionaryIds []string) []string {
	matchingIds := make([]string, 0)
	invertedMatchMap := h.config.getInvertedMatchMap()

	for _, id := range dictionaryIds {
		pattern := invertedMatchMap[id]
		if matches(pattern, req.URL.String()) {
			matchingIds = append(matchingIds, id)
		}
	}

	return matchingIds
}

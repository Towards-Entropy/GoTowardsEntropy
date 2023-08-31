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

	"github.com/DataDog/zstd"
)

func Compress(r io.Reader, w io.Writer, dictionaryId string) error {
	config := getConfig()
	dictionary := getDictionary(dictionaryId)
	if dictionary == nil && dictionaryId != "" {
		return fmt.Errorf("dictionary with id '%s' not found", dictionaryId)
	}

	var zw *zstd.Writer
	if dictionary == nil {
		zw = zstd.NewWriterLevel(w, config.CompressionLevel)
	} else {
		zw = zstd.NewWriterLevelDict(w, config.CompressionLevel, dictionary.Bytes)
	}
	defer zw.Close()

	buf := make([]byte, config.BufferSize)
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

	return nil
}

func Decompress(r io.Reader, w io.Writer, dictionaryId string) error {
	config := getConfig()
	dictionary := getDictionary(dictionaryId)

	if dictionary == nil && dictionaryId != "" {
		return fmt.Errorf("dictionary with id '%s' not found", dictionaryId)
	}

	var zr io.ReadCloser
	if dictionary == nil {
		zr = zstd.NewReader(r)
	} else {
		zr = zstd.NewReaderDict(r, dictionary.Bytes)
	}
	defer zr.Close()

	// TODO buffer size from config
	buf := make([]byte, config.BufferSize)

	for {
		n, err := zr.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading from compressed source: %v", err)
		}

		_, err = w.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("error writing decompressed data: %v", err)
		}
	}

	return nil
}

func CompressFile(b []byte, w io.Writer, dictionaryId string) error {
	r := bytes.NewReader(b)
	return Compress(r, w, dictionaryId)
}

func DecompressFile(b []byte, w io.Writer, dictionaryId string) error {
	r := bytes.NewReader(b)
	return Decompress(r, w, dictionaryId)
}

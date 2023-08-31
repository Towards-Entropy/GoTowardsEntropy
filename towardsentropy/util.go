/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"net/http"

	"github.com/DataDog/zstd"
)

type zstdResponseWriter struct {
	http.ResponseWriter
	Writer *zstd.Writer
}

func (z *zstdResponseWriter) Write(b []byte) (int, error) {
	return z.Writer.Write(b)
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

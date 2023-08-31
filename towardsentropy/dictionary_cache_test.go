/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"testing"
)

func TestUpdateCacheFromDir(t *testing.T) {
	dictionaries = make(map[string]Dictionary)
	updateCacheFromDir("../testdata/dictionaries")
	dict := getDictionary("enwik8")
	if dict == nil {
		t.Fatalf("Expected dictionary 'enwik8' to be loaded")
	}
}

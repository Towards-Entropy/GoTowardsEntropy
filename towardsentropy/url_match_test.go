/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"testing"
)

func TestMatches(t *testing.T) {
	testCases := []struct {
		matchPattern string
		targetURL    string
		expected     bool
	}{
		{"/app1/main*", "https://www.example.com/app1/main_12345.js", true},
		{"main*", "https://www.example.com/app1/main_1.js", true},
		{"main*", "https://www.example.com/app2/main.xyz.js", true},
		{"/app2/main*", "/app2/main_12345.js", true},
		{"/app1/main*", "/app2/main_12345.js", false},
		{"/app1/main*", "main_12345.js", false},            // edge case: matchPattern is absolute but targetURL is relative
		{"/app1/*", "https://www.example.com/app1/", true}, // edge case: matches directory
		{"https://www.example.com/app1/*", "https://www.example.com/app1/main_12345.js", true},
		{"https://www.example.com/app1/*", "https://www.example2.com/app1/main_12345.js", false},
	}

	for _, tc := range testCases {
		t.Run(tc.matchPattern+"_"+tc.targetURL, func(t *testing.T) {
			got := matches(tc.matchPattern, tc.targetURL)
			if got != tc.expected {
				t.Fatalf("Expected matches(%q, %q) to be %v, got %v", tc.matchPattern, tc.targetURL, tc.expected, got)
			}
		})
	}
}

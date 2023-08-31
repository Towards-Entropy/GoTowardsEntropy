/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"net/url"
	"regexp"
	"strings"
)

func matches(matchPattern, targetURL string) bool {
	// convert wildcard pattern to regular expression
	regexPattern := regexp.QuoteMeta(matchPattern)               // QuoteMeta escapes any special characters
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*") // replace escaped * with .*

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}

	// parse the target URL
	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}

	// decide on which part of URL to match against
	var targetString string
	if u.IsAbs() { // absolute URL
		// If matchPattern doesn't have a protocol, only match against the path of the target URL
		if !strings.HasPrefix(matchPattern, "http://") && !strings.HasPrefix(matchPattern, "https://") {
			targetString = u.Path
		} else {
			targetString = u.String()
		}
	} else { // relative URL or path
		targetString = u.Path
	}

	return re.MatchString(targetString)
}

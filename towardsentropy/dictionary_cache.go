/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package towardsentropy

import (
	"fmt"
	"os"
	"path/filepath"
)

type Dictionary struct {
	Id    string
	Bytes []byte
}

var (
	dictionaries = make(map[string]Dictionary)
)

func addDictionary(dict Dictionary) {
	dictionaries[dict.Id] = dict
}

func getDictionary(id string) *Dictionary {
	if id == "" {
		return nil
	}

	dict, ok := dictionaries[id]
	if !ok {
		return nil
	}
	return &dict
}

func updateCacheFromDir(path string) error {
	return filepath.Walk(path, maybeUpdateDictionary)
}

func maybeUpdateDictionary(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if filepath.Ext(path) != ".dict" {
		return nil
	}

	dictionaryId := filepath.Base(path)
	dictionaryId = dictionaryId[:len(dictionaryId)-len(".dict")]
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file")
	}
	addDictionary(Dictionary{dictionaryId, bytes})
	return nil
}

func findDictionary(dictionaryIds []string) *Dictionary {
	for _, id := range dictionaryIds {
		if id == "" {
			continue
		}
		if dict := getDictionary(id); dict != nil {
			return dict
		}
	}
	return nil
}

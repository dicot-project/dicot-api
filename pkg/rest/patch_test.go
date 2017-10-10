/*
 * This file is part of the Dicot project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package rest

import (
	"reflect"
	"testing"
)

type PatchPathData struct {
	Input  string
	Output []string
}

func TestDecodePatchPath(t *testing.T) {
	data := []PatchPathData{
		PatchPathData{
			Input:  "foo",
			Output: []string{},
		},
		PatchPathData{
			Input:  "/bar~3",
			Output: []string{},
		},
		PatchPathData{
			Input:  "/~0~1.ssh~1",
			Output: []string{"~/.ssh/"},
		},
		PatchPathData{
			Input:  "/foo/bar/wizz",
			Output: []string{"foo", "bar", "wizz"},
		},
		PatchPathData{
			Input:  "/~1foo/bar/wizz~1",
			Output: []string{"/foo", "bar", "wizz/"},
		},
	}

	for _, entry := range data {
		actual, err := DecodePatchPath(entry.Input)
		if err != nil {
			if len(entry.Output) != 0 {
				t.Errorf("Expected '%s' but got error '%s'", entry.Output, err)
			}
		} else {
			if !reflect.DeepEqual(actual, entry.Output) {
				t.Errorf("Expected '%s' but got '%s'", entry.Output, actual)
			}
		}
	}
}

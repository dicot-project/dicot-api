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
	"fmt"
	"reflect"
)

type PatchFieldInfo struct {
	Name       []string
	String     *string
	StringPtr  **string
	Boolean    *bool
	UInt64     *uint64
	UInt64Ptr  **uint64
	StringList *[]string
}

type PatchChange struct {
	Op    string
	Path  string
	Value interface{}
}

const (
	OP_ADD     = "add"
	OP_REMOVE  = "remove"
	OP_REPLACE = "replace"
)

func GetPatchField(name []string, fields []PatchFieldInfo) (PatchFieldInfo, error) {
	for _, field := range fields {
		if reflect.DeepEqual(name, field.Name) {
			return field, nil
		}
	}

	return PatchFieldInfo{}, fmt.Errorf("No field named %s", name)
}

func toStringList(vals interface{}) ([]string, bool) {
	valslist, ok := vals.([]interface{})
	if !ok {
		return []string{}, false
	}

	strvals := []string{}
	for _, val := range valslist {
		strval, ok := val.(string)
		if !ok {
			return []string{}, false
		}
		strvals = append(strvals, strval)
	}
	return strvals, true
}

func StringListAdd(vals []string, val string) []string {
	found := false
	for _, el := range vals {
		if el == val {
			found = true
			break
		}
	}
	if !found {
		vals = append(vals, val)
	}
	return vals
}

func StringListDel(vals []string, val string) []string {
	newvals := []string{}
	for _, el := range vals {
		if el == val {
			continue
		}
		newvals = append(newvals, el)
	}
	return newvals
}

func ApplyPatch(changes []PatchChange, fields []PatchFieldInfo, fallback *map[string]string) (bool, error) {
	changed := false
	for _, change := range changes {
		name, err := DecodePatchPath(change.Path)
		if err != nil {
			return false, err
		}

		field, err := GetPatchField(name, fields)
		if err != nil {
			if fallback == nil {
				return false, err
			}

			if *fallback == nil {
				*fallback = make(map[string]string)
			}

			if len(name) != 1 {
				return false, fmt.Errorf("Field '%s' must have one component", name)
			}

			val, ok := change.Value.(string)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a string", name)
			}

			switch change.Op {
			case OP_REPLACE, OP_ADD:
				(*fallback)[name[0]] = val
				changed = true

			case OP_REMOVE:
				delete(*fallback, name[0])
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}

		} else if field.String != nil {
			val, ok := change.Value.(string)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a string", name)
			}

			switch change.Op {
			case OP_REPLACE:
				*field.String = val
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else if field.StringPtr != nil {
			val, ok := change.Value.(string)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a string", name)
			}

			switch change.Op {
			case OP_REPLACE:
				*field.StringPtr = &val
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else if field.Boolean != nil {
			val, ok := change.Value.(bool)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a boolean", name)
			}

			switch change.Op {
			case OP_REPLACE:
				*field.Boolean = val
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else if field.UInt64 != nil {
			val, ok := change.Value.(float64)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a float64", name)
			}
			ival := uint64(val)

			switch change.Op {
			case OP_REPLACE:
				*field.UInt64 = ival
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else if field.UInt64Ptr != nil {
			val, ok := change.Value.(float64)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a float64", name)
			}
			ival := uint64(val)

			switch change.Op {
			case OP_REPLACE:
				*field.UInt64Ptr = &ival
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else if field.StringList != nil {
			val, ok := toStringList(change.Value)
			if !ok {
				return false, fmt.Errorf("Field '%s' expects a string list", name)
			}

			switch change.Op {
			case OP_REPLACE:
				*field.StringList = val
				changed = true

			case OP_ADD:
				for _, el := range val {
					*field.StringList = StringListAdd(*field.StringList, el)
				}
				changed = true

			case OP_REMOVE:
				for _, el := range val {
					*field.StringList = StringListDel(*field.StringList, el)
				}
				changed = true

			default:
				return false, fmt.Errorf("Unsupported operation '%s' for field '%s'", change.Op, name)
			}
		} else {
			return false, fmt.Errorf("Field '%s' does not permit updates", name)
		}
	}

	return changed, nil
}

func DecodePatchPath(path string) ([]string, error) {
	if path[0] != '/' {
		return []string{}, fmt.Errorf("Expected leading '/' in property path")
	}

	ret := []string{}
	name := ""
	escape := false
	for i := 1; i < len(path); i++ {
		if escape {
			if path[i] == '0' {
				name = name + "~"
			} else if path[i] == '1' {
				name = name + "/"
			} else {
				return []string{}, fmt.Errorf("Illegal escape sequence ~%s", path[i:i+1])
			}
			escape = false
		} else if path[i] == '~' {
			escape = true
		} else if path[i] == '/' {
			ret = append(ret, name)
			name = ""
		} else {
			name = name + path[i:i+1]
		}
	}

	ret = append(ret, name)

	return ret, nil
}

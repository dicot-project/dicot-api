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

package v2_1

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetFilterUInt(c *gin.Context, name string) (bool, uint64) {
	val := c.Query(name)
	if name == "" {
		return false, 0
	}
	res, err := strconv.ParseUint(val, 10, 0)
	if err != nil {
		return false, 0
	}
	return true, res
}

func GetFilterBool(c *gin.Context, name string) (bool, bool) {
	val := c.Query(name)
	if name == "" || name == "None" {
		return false, false
	}
	res, err := strconv.ParseBool(val)
	if err != nil {
		return false, false
	}
	return true, res
}

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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func AcceptJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctypes := []string{
			"application/json",
			"application/*",
			"*/json",
			"*/*",
		}

		ctype := c.NegotiateFormat(ctypes...)

		if ctype == "" {
			c.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server"))
			return
		}

		c.Next()
	}
}

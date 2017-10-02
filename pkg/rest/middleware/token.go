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

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/dicot-project/dicot-api/pkg/auth"
)

type TokenHandler struct {
	TokenManager auth.TokenManager
}

func SetToken(c *gin.Context, tok *auth.Token) {
	glog.V(1).Infof("Set token %s", tok)
	c.Set("Token", tok)
}

func GetToken(c *gin.Context) *auth.Token {
	obj, ok := c.Get("Token")
	if ok {
		return nil
	}
	ver, ok := obj.(*auth.Token)
	if !ok {
		return nil
	}
	return ver
}

func (h *TokenHandler) middleware(allowAnon bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		toksig := c.GetHeader("X-Auth-Token")

		if toksig == "" {
			if !allowAnon {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}

		token, err := h.TokenManager.ValidateToken(toksig)

		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		SetToken(c, token)
	}
}

func (h *TokenHandler) MiddlewareNoAnon() gin.HandlerFunc {
	return h.middleware(false)
}

func (h *TokenHandler) MiddlewareAllowAnon() gin.HandlerFunc {
	return h.middleware(true)
}

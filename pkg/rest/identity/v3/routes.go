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

package v3

import (
	"github.com/gin-gonic/gin"

	"github.com/dicot-project/dicot-api/pkg/rest"
)

type service struct {
	Prefix   string
	Services *rest.ServiceList
}

func NewService(svcs *rest.ServiceList, prefix string) rest.Service {
	if prefix == "" {
		prefix = "/identity/v3"
	}
	return &service{
		Prefix: prefix,
	}
}

func (svc *service) GetPrefix() string {
	return svc.Prefix
}

func (svc *service) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", svc.IndexGet)
	router.POST("/auth/tokens", svc.TokensPost)
}

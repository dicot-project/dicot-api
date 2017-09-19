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

package v2

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dicot-project/dicot-api/pkg/rest"
)

type IndexRes struct {
	Versions []VersionInfo `json:"version"`
}

type VersionIndexRes struct {
	Version VersionInfo `json:"version"`
}

type VersionInfo struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Links  []rest.LinkInfo `json:"links"`
}

func (svc *service) commonVersionList(c *gin.Context) {
	res := IndexRes{
		Versions: []VersionInfo{
			// Only try to support CURRENT API from Ocata onwards
			VersionInfo{
				ID:     "v2.5",
				Status: "CURRENT",
			},
		},
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) IndexShow(c *gin.Context) {
	svc.commonVersionList(c)
}

func (svc *service) VersionIndexShow(c *gin.Context) {
	svc.commonVersionList(c)
}

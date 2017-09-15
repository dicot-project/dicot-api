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
	ID         string          `json:"id"`
	Status     string          `json:"status"`
	Updated    string          `json:"updated"`
	MinVersion string          `json:"min_version"`
	Version    string          `json:"version"`
	MediaTypes []MediaTypeInfo `json:"media-types"`
	Links      []rest.LinkInfo `json:"links"`
}

func (svc *service) IndexShow(c *gin.Context) {
	res := IndexRes{
		Versions: []VersionInfo{
			// Only try to support API from Pike onwards, microverison 2.53
			// This requires use of novaclient >= 9.1.0
			VersionInfo{
				ID:         "v2.1",
				Version:    "2.53",
				MinVersion: "2.53",
				Status:     "CURRENT",
				Updated:    "2013-07-23T11:33:21Z",
			},
		},
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) VersionIndexShow(c *gin.Context) {
	res := VersionIndexRes{
		Version: VersionInfo{
			ID:         "v2.1",
			Version:    "2.53",
			MinVersion: "2.53",
			Status:     "CURRENT",
			Updated:    "2013-07-23T11:33:21Z",
			MediaTypes: []MediaTypeInfo{
				MediaTypeInfo{
					Base: "application/json",
					Type: "application/vnd.openstack.compute+json;version=2.1",
				},
			},
		},
	}

	c.JSON(http.StatusOK, res)
}

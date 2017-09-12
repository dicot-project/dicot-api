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
	"net/http"

	"github.com/gin-gonic/gin"
)

type VersionReq struct {
	Version VersionInfo `json:"version"`
}

type VersionInfo struct {
	Status     string                 `json:"status"`
	Updated    string                 `json:"updated"`
	MediaTypes []VersionInfoMediaType `json:"media-types"`
	ID         string                 `json:"id"`
	Links      []VersionInfoLink      `json:"links"`
}

type VersionInfoMediaType struct {
	Base string `json:"base"`
	Type string `json:"type"`
}

type VersionInfoLink struct {
	HRef string `json:"href"`
	Rel  string `json:"rel"`
}

func (svc *service) IndexGet(c *gin.Context) {
	res := VersionReq{
		Version: VersionInfo{
			Status:  "stable",
			Updated: "2017-02-22T00:00:00Z",
			MediaTypes: []VersionInfoMediaType{
				VersionInfoMediaType{
					Base: "application/json",
					Type: "application/vnd.openstack.identity-v3+json",
				},
			},
			ID: "v3.8",
			Links: []VersionInfoLink{
				VersionInfoLink{
					HRef: "http://" + c.Request.Host + c.Request.URL.String(),
					Rel:  "self",
				},
			},
		},
	}
	c.JSON(http.StatusOK, res)
}

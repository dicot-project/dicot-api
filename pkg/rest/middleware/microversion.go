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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"

	"github.com/dicot-project/dicot-api/pkg/rest"
)

type MicroVersionErrorRes struct {
	Errors []MicroVersionErrorInfo `json:"errors"`
}

type MicroVersionErrorInfo struct {
	RequestID  string          `json:"request_id"`
	Code       string          `json:"code"`
	Status     uint            `json:"status"`
	Title      string          `json:"title"`
	Detail     string          `json:"detail"`
	MaxVersion string          `json:"max_version"`
	MinVersion string          `json:"min_version"`
	Links      []rest.LinkInfo `json:"links"`
}

type MicroVersion struct {
	Major int
	Micro int
}

func ParseMicroVersion(ver string) *MicroVersion {
	bits := strings.Split(ver, ".")

	if len(bits) != 2 {
		return nil
	}

	maj, err := strconv.Atoi(bits[0])
	if err != nil {
		return nil
	}

	mic, err := strconv.Atoi(bits[1])
	if err != nil {
		return nil
	}

	return &MicroVersion{
		Major: maj,
		Micro: mic,
	}
}

func (ver *MicroVersion) String() string {
	return fmt.Sprintf("%d.%d", ver.Major, ver.Micro)
}

func (ver *MicroVersion) InRange(min, max *MicroVersion) bool {
	if ((ver.Major == min.Major && ver.Micro >= min.Micro) ||
		(ver.Major > min.Major)) &&
		((ver.Major == max.Major && ver.Micro <= max.Micro) ||
			(ver.Major < max.Major)) {
		return true
	}

	return false
}

type MicroVersionHandler struct {
	Service       string
	ServiceHeader string
	Min           *MicroVersion
	Max           *MicroVersion
}

func SetMicroVersion(c *gin.Context, ver *MicroVersion) {
	glog.V(1).Infof("Set micro version %s", ver.String())
	c.Set("MicroVersion", ver)
}

func GetMicroVersion(c *gin.Context) *MicroVersion {
	obj, ok := c.Get("MicroVersion")
	if ok {
		return nil
	}
	ver, ok := obj.(*MicroVersion)
	if !ok {
		return nil
	}
	return ver
}

func (h *MicroVersionHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		verHeader := c.GetHeader("Openstack-API-Version")

		var verString string
		if verHeader == "" {
			if h.ServiceHeader != "" {
				verString = c.GetHeader(h.ServiceHeader)
			}
			if verString == "" {
				SetMicroVersion(c, h.Min)
				return
			}
		} else {
			bits := strings.Split(strings.Trim(verHeader, " "), " ")

			if len(bits) != 2 {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			if bits[0] != h.Service {
				SetMicroVersion(c, h.Min)
				return
			}

			verString = bits[1]
		}

		if verString == "latest" {
			SetMicroVersion(c, h.Max)
			return
		}

		got := ParseMicroVersion(verString)
		if got == nil {
			glog.V(1).Infof("Malformed microversion %s", verString)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if got.InRange(h.Min, h.Max) {
			SetMicroVersion(c, got)
			return
		}

		glog.V(1).Infof("Reject request with microversion %s", got.String())
		res := MicroVersionErrorRes{
			Errors: []MicroVersionErrorInfo{
				MicroVersionErrorInfo{
					Code:       h.Service + ".microversion-unsupported",
					Status:     http.StatusNotAcceptable,
					Title:      "Requested microversion is unsupported",
					Detail:     fmt.Sprintf("Version %s is not supported by the API. Minimum is %s and maximum is %s.", got.String(), h.Min.String(), h.Max.String()),
					MaxVersion: h.Max.String(),
					MinVersion: h.Min.String(),
					Links: []rest.LinkInfo{
						rest.LinkInfo{
							Rel:  "help",
							HRef: "http://developer.openstack.org/api-guide/compute/microversions.html",
						},
					},
				},
			},
		}

		c.AbortWithStatusJSON(http.StatusNotAcceptable, res)
		return
	}
}

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
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/dicot-project/dicot-api/pkg/api/image"
	"github.com/dicot-project/dicot-api/pkg/rest"
)

type SchemaImageShowRes struct {
	Links                []rest.LinkInfo           `json:"links"`
	Name                 string                    `json:"name"`
	Properties           map[string]SchemaPropInfo `json:"properties"`
	AdditionalProperties map[string]string         `json:"additionalProperties"`
}

type SchemaPropInfo struct {
	Description string           `json:"description"`
	Type        interface{}      `json:"type"`
	ReadOnly    *bool            `json:"readonly,omitempty"`
	Enum        []*string        `json:"enum,omitempty"`
	Pattern     string           `json:"pattern,omitempty"`
	IsBase      *bool            `json:"is_base,omitempty"`
	MaxLength   uint             `json:"maxLength,omitempty"`
	Items       *SchemaPropItems `json:"items,omitempty"`
}

type SchemaPropItems struct {
	Properties map[string]SchemaPropInfo `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
	MaxLength  uint                      `json:"maxLength,omitempty"`
	Type       string                    `json:"type"`
}

func (svc *service) SchemaImageShow(c *gin.Context) {

	trueval := true
	falseval := false

	res := SchemaImageShowRes{
		AdditionalProperties: map[string]string{
			"type": "string",
		},
		Links: []rest.LinkInfo{
			rest.LinkInfo{
				HRef: "{first}",
				Rel:  "first",
			},
			rest.LinkInfo{
				HRef: "{next}",
				Rel:  "next",
			},
			rest.LinkInfo{
				HRef: "{schema}",
				Rel:  "describedby",
			},
		},
		Name: "image",
		Properties: map[string]SchemaPropInfo{
			"architecture": SchemaPropInfo{
				Description: "Operating system architecture as specified in https://docs.openstack.org/python-glanceclient/latest/cli/property-keys.html",
				IsBase:      &falseval,
				Type:        "string",
			},
			"checksum": SchemaPropInfo{
				Description: "md5 hash of image contents.",
				MaxLength:   32,
				ReadOnly:    &trueval,
				Type:        []string{"null", "string"},
			},
			"container_format": SchemaPropInfo{
				Description: "Format of the container",
				Enum: []*string{
					nil,
					&image.IMAGE_CONTAINER_FORMAT_AMI,
					&image.IMAGE_CONTAINER_FORMAT_AKI,
					&image.IMAGE_CONTAINER_FORMAT_ARI,
					&image.IMAGE_CONTAINER_FORMAT_BARE,
				},
				Type: []string{"null", "string"},
			},
			"created_at": SchemaPropInfo{
				Description: "Date and time of image registration",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"direct_url": SchemaPropInfo{
				Description: "URL to access the image file kept in external store",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"disk_format": SchemaPropInfo{
				Description: "Format of the disk",
				Enum: []*string{
					nil,
					&image.IMAGE_DISK_FORMAT_AMI,
					&image.IMAGE_DISK_FORMAT_AKI,
					&image.IMAGE_DISK_FORMAT_ARI,
					&image.IMAGE_DISK_FORMAT_RAW,
					&image.IMAGE_DISK_FORMAT_QCOW2,
					&image.IMAGE_DISK_FORMAT_ISO,
				},
				Type: []string{"null", "string"},
			},
			"file": SchemaPropInfo{
				Description: "An image file url",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"id": SchemaPropInfo{
				Description: "An identifier for the image",
				Pattern:     "^([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}$",
				Type:        "string",
			},
			"instance_uuid": SchemaPropInfo{
				Description: "Metadata which can be used to record which instance this image is associated with. (Informational only, does not create an instance snapshot.)",
				IsBase:      &falseval,
				Type:        "string",
			},
			"kernel_id": SchemaPropInfo{
				Description: "ID of image stored in Glance that should be used as the kernel when booting an AMI-style image.",
				IsBase:      &falseval,
				Pattern:     "^([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}$",
				Type:        []string{"null", "string"},
			},
			"locations": SchemaPropInfo{
				Description: "A set of URLs to access the image file kept in external store",
				Items: &SchemaPropItems{
					Properties: map[string]SchemaPropInfo{
						"metadata": SchemaPropInfo{
							Type: "object",
						},
						"url": SchemaPropInfo{
							MaxLength: 255,
							Type:      "string",
						},
					},
					Required: []string{"url", "metadata"},
					Type:     "object",
				},
				Type: "array",
			},
			"min_disk": SchemaPropInfo{
				Description: "Amount of disk space (in GB) required to boot image.",
				Type:        "integer",
			},
			"min_ram": SchemaPropInfo{
				Description: "Amount of ram (in MB) required to boot image.",
				Type:        "integer",
			},
			"name": SchemaPropInfo{
				Description: "Descriptive name for the image",
				MaxLength:   255,
				Type: []string{
					"null",
					"string",
				},
			},
			"os_distro": SchemaPropInfo{
				Description: "Common name of operating system distribution as specified in https://docs.openstack.org/python-glanceclient/latest/cli/property-keys.html",
				IsBase:      &falseval,
				Type:        "string",
			},
			"os_version": SchemaPropInfo{
				Description: "Operating system version as specified by the distributor",
				IsBase:      &falseval,
				Type:        "string",
			},
			"owner": SchemaPropInfo{
				Description: "Owner of the image",
				MaxLength:   255,
				Type:        []string{"null", "string"},
			},
			"protected": SchemaPropInfo{
				Description: "If true, image will not be deletable.",
				Type:        "boolean",
			},
			"ramdisk_id": SchemaPropInfo{
				Description: "ID of image stored in Glance that should be used as the ramdisk when booting an AMI-style image",
				IsBase:      &falseval,
				Pattern:     "^([0-9a-fA-F]){8}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){4}-([0-9a-fA-F]){12}$",
				Type:        []string{"null", "string"},
			},
			"schema": SchemaPropInfo{
				Description: "An image schema url",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"self": SchemaPropInfo{
				Description: "An image self url",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"size": SchemaPropInfo{
				Description: "Size of image file in bytes",
				ReadOnly:    &trueval,
				Type:        []string{"null", "integer"},
			},
			"status": SchemaPropInfo{
				Description: "Status of the image",
				Enum: []*string{
					&image.IMAGE_STATUS_QUEUED,
					&image.IMAGE_STATUS_SAVING,
					&image.IMAGE_STATUS_ACTIVE,
					&image.IMAGE_STATUS_KILLED,
					&image.IMAGE_STATUS_DELETED,
					&image.IMAGE_STATUS_PENDING_DELETE,
					&image.IMAGE_STATUS_DEACTIVATED,
				},
				ReadOnly: &trueval,
				Type:     "string",
			},
			"tags": SchemaPropInfo{
				Description: "List of strings related to the image",
				Type:        "array",
				Items: &SchemaPropItems{
					MaxLength: 255,
					Type:      "string",
				},
			},
			"updated_at": SchemaPropInfo{
				Description: "Date and time of the last image modification",
				ReadOnly:    &trueval,
				Type:        "string",
			},
			"virtual_size": SchemaPropInfo{
				Description: "Virtual size of image in bytes",
				ReadOnly:    &trueval,
				Type:        []string{"null", "string"},
			},
			"visibility": SchemaPropInfo{
				Description: "Scope of image accessibility",
				Enum: []*string{
					&image.IMAGE_VISIBILITY_PUBLIC,
					&image.IMAGE_VISIBILITY_COMMUNITY,
					&image.IMAGE_VISIBILITY_SHARED,
					&image.IMAGE_VISIBILITY_PRIVATE,
				},
				Type: "string",
			},
		},
	}

	c.JSON(http.StatusOK, res)
}

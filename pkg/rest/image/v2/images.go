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
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	k8sv1 "k8s.io/client-go/pkg/api/v1"

	identityv1 "github.com/dicot-project/dicot-api/pkg/api/identity/v1"
	"github.com/dicot-project/dicot-api/pkg/api/image"
	"github.com/dicot-project/dicot-api/pkg/api/image/v1"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

type ImageCreateReq struct {
	ID              string   `json:"id"`
	Name            *string  `json:"name"`
	ContainerFormat *string  `json:"container_format"`
	DiskFormat      *string  `json:"disk_format"`
	Visibility      *string  `json:"visibility"`
	Protected       *bool    `json:"protected"`
	MinDisk         uint64   `json:"min_disk"`
	MinRam          uint64   `json:"min_ram"`
	Tags            []string `json:"tags"`
}

type ImageListRes struct {
	Images []ImageInfo `json:"images"`
}

type ImageInfo struct {
	ID              string   `json:"id"`
	Name            *string  `json:"name"`
	File            string   `json:"file"`
	Schema          string   `json:"schema"`
	Status          string   `json:"status"`
	ContainerFormat *string  `json:"container_format"`
	DiskFormat      *string  `json:"disk_format"`
	Visibility      string   `json:"visibility"`
	Protected       bool     `json:"protected"`
	Size            *uint64  `json:"size"`
	VirtualSize     *uint64  `json:"virtual_size"`
	Owner           string   `json:"owner"`
	MinDisk         uint64   `json:"min_disk"`
	MinRam          uint64   `json:"min_ram"`
	Checksum        *string  `json:"checksum"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Tags            []string `json:"tags"`
}

func ImageAccessible(img *v1.Image, proj *identityv1.Project) bool {
	if img.ObjectMeta.Namespace == proj.Spec.Namespace {
		return true
	}

	switch img.Spec.Visibility {
	case image.IMAGE_VISIBILITY_PUBLIC:
		return true
	case image.IMAGE_VISIBILITY_COMMUNITY:
		return true
	case image.IMAGE_VISIBILITY_SHARED:
		// XXX validate sharing rules
		return false
	case image.IMAGE_VISIBILITY_PRIVATE:
		return false
	}

	panic("Unexpected visibility")
}

func (svc *service) ImageList(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)

	clnt := image.NewImageClient(svc.ImageClient, k8sv1.NamespaceAll)

	imgs, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := ImageListRes{
		Images: []ImageInfo{},
	}

	for _, img := range imgs.Items {
		if !ImageAccessible(&img, proj) {
			continue
		}

		info := ImageInfo{
			ID:              img.Spec.ID,
			Name:            img.Spec.Name,
			File:            "/",
			Schema:          "/v2/schemas/image",
			Owner:           img.Spec.Owner,
			Status:          img.Spec.Status,
			ContainerFormat: img.Spec.ContainerFormat,
			DiskFormat:      img.Spec.DiskFormat,
			MinDisk:         img.Spec.MinDisk,
			MinRam:          img.Spec.MinRam,
			Protected:       img.Spec.Protected,
			Visibility:      img.Spec.Visibility,
			Tags:            img.Spec.Tags,
			CreatedAt:       img.Spec.CreatedAt,
			UpdatedAt:       img.Spec.UpdatedAt,
			Checksum:        nil,
		}
		res.Images = append(res.Images, info)
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageCreate(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	var req ImageCreateReq
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clnt := image.NewImageClient(svc.ImageClient, k8sv1.NamespaceAll)

	if req.ID == "" {
		req.ID = string(uuid.NewUUID())
	} else {
		img, err := clnt.GetByID(req.ID)
		if err != nil && !errors.IsNotFound(err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if img != nil {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
	}

	clnt = image.NewImageClient(svc.ImageClient, proj.Spec.Namespace)

	if req.Name != nil {
		img, err := clnt.Get(*req.Name)
		if err != nil && !errors.IsNotFound(err) {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if img != nil {
			c.AbortWithStatus(http.StatusConflict)
			return
		}
	}

	if req.Visibility == nil {
		shared := image.IMAGE_VISIBILITY_SHARED
		req.Visibility = &shared
	} else {
		if !image.IsValidVisibility(*req.Visibility) {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	if req.ContainerFormat != nil && !image.IsValidContainerFormat(*req.ContainerFormat) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.DiskFormat != nil && !image.IsValidDiskFormat(*req.DiskFormat) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.Protected == nil {
		notprot := false
		req.Protected = &notprot
	}

	var name string
	if req.Name == nil || *req.Name == "" {
		name = fmt.Sprintf("img-%s", req.ID)
	} else {
		name = *req.Name
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	glog.V(1).Infof("Use name %s", name)

	img := &v1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.ImageSpec{
			ID:              req.ID,
			Name:            req.Name,
			Status:          image.IMAGE_STATUS_QUEUED,
			ContainerFormat: req.ContainerFormat,
			DiskFormat:      req.DiskFormat,
			Owner:           string(proj.ObjectMeta.UID),
			MinDisk:         req.MinDisk,
			MinRam:          req.MinRam,
			Protected:       *req.Protected,
			Visibility:      *req.Visibility,
			Tags:            req.Tags,
			CreatedAt:       time.Now().Format(time.RFC3339),
			UpdatedAt:       time.Now().Format(time.RFC3339),
		},
	}

	img, err = clnt.Create(img)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// XXX Links field
	res := ImageInfo{
		ID:              img.Spec.ID,
		Name:            img.Spec.Name,
		File:            "/",
		Schema:          "/v2/schemas/image",
		Status:          img.Spec.Status,
		Owner:           img.Spec.Owner,
		ContainerFormat: img.Spec.ContainerFormat,
		DiskFormat:      img.Spec.DiskFormat,
		MinDisk:         img.Spec.MinDisk,
		MinRam:          img.Spec.MinRam,
		Protected:       img.Spec.Protected,
		Visibility:      img.Spec.Visibility,
		Tags:            img.Spec.Tags,
		CreatedAt:       img.Spec.CreatedAt,
		UpdatedAt:       img.Spec.UpdatedAt,
		Checksum:        nil,
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageShow(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := image.NewImageClient(svc.ImageClient, k8sv1.NamespaceAll)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if !ImageAccessible(img, proj) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	res := ImageInfo{
		ID:              img.Spec.ID,
		Name:            img.Spec.Name,
		Status:          img.Spec.Status,
		File:            "/",
		Schema:          "/v2/schemas/image",
		Owner:           img.Spec.Owner,
		ContainerFormat: img.Spec.ContainerFormat,
		DiskFormat:      img.Spec.DiskFormat,
		MinDisk:         img.Spec.MinDisk,
		MinRam:          img.Spec.MinRam,
		Protected:       img.Spec.Protected,
		Visibility:      img.Spec.Visibility,
		Tags:            img.Spec.Tags,
		CreatedAt:       img.Spec.CreatedAt,
		UpdatedAt:       img.Spec.UpdatedAt,
		Checksum:        nil,
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageDelete(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := image.NewImageClient(svc.ImageClient, proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if img.Spec.Protected {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = clnt.Delete(img.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

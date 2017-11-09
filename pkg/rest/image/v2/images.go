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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	"github.com/dicot-project/dicot-api/pkg/rest"
	"github.com/dicot-project/dicot-api/pkg/rest/middleware"
)

const TEN_GB = 10 * 1024 * 1024 * 1024

type ImageCreateReq struct {
	ID              string            `json:"id"`
	Name            *string           `json:"name"`
	ContainerFormat *string           `json:"container_format"`
	DiskFormat      *string           `json:"disk_format"`
	Visibility      *string           `json:"visibility"`
	Protected       *bool             `json:"protected"`
	MinDisk         uint64            `json:"min_disk"`
	MinRam          uint64            `json:"min_ram"`
	Tags            []string          `json:"tags"`
	Metadata        map[string]string `json:"-"`
}

type ImageListRes struct {
	Images []ImageInfo `json:"images"`
}

type ImageInfo struct {
	ID              string            `json:"id"`
	Name            *string           `json:"name"`
	File            string            `json:"file"`
	Schema          string            `json:"schema"`
	Status          string            `json:"status"`
	ContainerFormat *string           `json:"container_format"`
	DiskFormat      *string           `json:"disk_format"`
	Visibility      string            `json:"visibility"`
	Protected       bool              `json:"protected"`
	Size            *uint64           `json:"size"`
	VirtualSize     *uint64           `json:"virtual_size"`
	Owner           string            `json:"owner"`
	MinDisk         uint64            `json:"min_disk"`
	MinRam          uint64            `json:"min_ram"`
	Checksum        *string           `json:"checksum"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
	Tags            []string          `json:"tags"`
	Metadata        map[string]string `json:"-"`
}

func (info ImageInfo) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})

	// Take everything in Extra
	for k, v := range info.Metadata {
		data[k] = v
	}

	// Take all the struct values with a json tag
	val := reflect.ValueOf(info)
	typ := reflect.TypeOf(info)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldv := val.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			data[jsonTag] = fieldv.Interface()
		}
	}
	return json.Marshal(data)
}

type _ImageCreateReq ImageCreateReq

func (info *ImageCreateReq) UnmarshalJSON(b []byte) error {
	info2 := _ImageCreateReq{}
	err := json.Unmarshal(b, &info2)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	typ := reflect.TypeOf(info2)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonTag != "" && jsonTag != "-" {
			delete(data, jsonTag)
		}
	}

	*info = ImageCreateReq(info2)

	info.Metadata = make(map[string]string)
	for key, val := range data {
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("Expecting a string for metadata properties")
		}
		info.Metadata[key] = str
	}

	return nil
}

type ImagePatchChange struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type ImagePatchReq []ImagePatchChange

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

	clnt := svc.Client.Image().Images(k8sv1.NamespaceAll)

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
			File:            fmt.Sprintf("/v2/images/%s/file", img.Spec.ID),
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
			Size:            img.Spec.Size,
			VirtualSize:     img.Spec.VirtualSize,
			Checksum:        nil,
			Metadata:        img.Spec.Metadata,
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

	clnt := svc.Client.Image().Images(k8sv1.NamespaceAll)

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

	clnt = svc.Client.Image().Images(proj.Spec.Namespace)

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
			Metadata:        req.Metadata,
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
		File:            fmt.Sprintf("/v2/images/%s/file", img.Spec.ID),
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
		Size:            img.Spec.Size,
		VirtualSize:     img.Spec.VirtualSize,
		Checksum:        nil,
		Metadata:        img.Spec.Metadata,
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageShow(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(k8sv1.NamespaceAll)

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
		File:            fmt.Sprintf("/v2/images/%s/file", img.Spec.ID),
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
		Size:            img.Spec.Size,
		VirtualSize:     img.Spec.VirtualSize,
		Checksum:        nil,
		Metadata:        img.Spec.Metadata,
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageDelete(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

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

func (svc *service) ImagePatch(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")
	var req ImagePatchReq
	err := c.BindJSON(&req)
	if err != nil {
		glog.V(1).Info("Failed to parse request")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	fields := []rest.PatchFieldInfo{
		rest.PatchFieldInfo{
			Name: []string{"id"},
		},
		rest.PatchFieldInfo{
			Name: []string{"name"},
		},
		rest.PatchFieldInfo{
			Name: []string{"status"},
		},
		rest.PatchFieldInfo{
			Name:      []string{"container_format"},
			StringPtr: &img.Spec.ContainerFormat,
		},
		rest.PatchFieldInfo{
			Name:      []string{"disk_format"},
			StringPtr: &img.Spec.DiskFormat,
		},
		rest.PatchFieldInfo{
			Name:   []string{"visibility"},
			String: &img.Spec.Visibility,
		},
		rest.PatchFieldInfo{
			Name:    []string{"protected"},
			Boolean: &img.Spec.Protected,
		},
		rest.PatchFieldInfo{
			Name: []string{"size"},
		},
		rest.PatchFieldInfo{
			Name: []string{"virtual_size"},
		},
		rest.PatchFieldInfo{
			Name: []string{"owner"},
		},
		rest.PatchFieldInfo{
			Name:   []string{"min_disk"},
			UInt64: &img.Spec.MinDisk,
		},
		rest.PatchFieldInfo{
			Name:   []string{"min_ram"},
			UInt64: &img.Spec.MinRam,
		},
		rest.PatchFieldInfo{
			Name: []string{"checksum"},
		},
		rest.PatchFieldInfo{
			Name: []string{"created_at"},
		},
		rest.PatchFieldInfo{
			Name: []string{"updated_at"},
		},
		rest.PatchFieldInfo{
			Name:       []string{"tags"},
			StringList: &img.Spec.Tags,
		},
	}

	changes := []rest.PatchChange{}
	for _, el := range req {
		changes = append(changes, rest.PatchChange{
			el.Op, el.Path, el.Value,
		})
	}

	glog.V(1).Infof("Changes %s", req)
	changed, err := rest.ApplyPatch(changes, fields, &img.Spec.Metadata)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if changed {
		img, err = clnt.Update(img)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	res := ImageInfo{
		ID:              img.Spec.ID,
		Name:            img.Spec.Name,
		Status:          img.Spec.Status,
		File:            fmt.Sprintf("/v2/images/%s/file", img.Spec.ID),
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
		Size:            img.Spec.Size,
		VirtualSize:     img.Spec.VirtualSize,
		Checksum:        nil,
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) ImageDeactivate(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if img.Spec.Status == image.IMAGE_STATUS_DEACTIVATED {
		c.String(http.StatusNoContent, "")
		return
	}

	if img.Spec.Status != image.IMAGE_STATUS_ACTIVE {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	img.Spec.Status = image.IMAGE_STATUS_DEACTIVATED

	img, err = clnt.Update(img)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

func (svc *service) ImageReactivate(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if img.Spec.Status == image.IMAGE_STATUS_ACTIVE {
		c.String(http.StatusNoContent, "")
		return
	}

	if img.Spec.Status != image.IMAGE_STATUS_DEACTIVATED {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	img.Spec.Status = image.IMAGE_STATUS_ACTIVE

	img, err = clnt.Update(img)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

func (svc *service) ImageTagAdd(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")
	tag := c.Param("tag")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	found := false
	for _, item := range img.Spec.Tags {
		if item == tag {
			found = true
			break
		}
	}

	if !found {
		img.Spec.Tags = append(img.Spec.Tags, tag)
		img, err = clnt.Update(img)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.String(http.StatusNoContent, "")
}

func (svc *service) ImageTagDelete(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")
	tag := c.Param("tag")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	tags := []string{}
	found := false
	for _, item := range img.Spec.Tags {
		if item == tag {
			found = true
		} else {
			tags = append(tags, item)
		}
	}

	if found {
		img.Spec.Tags = tags
		img, err = clnt.Update(img)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	c.String(http.StatusNoContent, "")
}

func (svc *service) ImageDataUpload(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	img, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	name := filepath.Join(svc.ImageRepo, imgID)

	dst, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	src := c.Request.Body

	// XX checksum
	n, err := io.Copy(dst, &io.LimitedReader{src, TEN_GB})
	if err != nil {
		dst.Close()
		os.Remove(name)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	size := uint64(n)
	img.Spec.Size = &size

	remain := make([]byte, 1)
	_, err = src.Read(remain)
	if err == nil || err != io.EOF {
		dst.Close()
		os.Remove(name)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = dst.Close()
	if err != nil {
		os.Remove(name)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	img.Spec.Status = image.IMAGE_STATUS_ACTIVE

	img, err = clnt.Update(img)
	if err != nil {
		os.Remove(name)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusNoContent, "")
}

func (svc *service) ImageDataDownload(c *gin.Context) {
	proj := middleware.RequiredTokenScopeProject(c)
	imgID := c.Param("imageID")

	clnt := svc.Client.Image().Images(proj.Spec.Namespace)

	_, err := clnt.GetByID(imgID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	name := filepath.Join(svc.ImageRepo, imgID)

	c.Status(http.StatusOK)
	c.File(name)
}

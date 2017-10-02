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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/dicot-project/dicot-api/pkg/api/compute"
	"github.com/dicot-project/dicot-api/pkg/api/compute/v1"
	"github.com/dicot-project/dicot-api/pkg/rest"
)

type FlavorCreateReq struct {
	Flavor FlavorInfoDetail `json:"flavor"`
}

type FlavorListRes struct {
	Flavors []FlavorInfo `json:"flavors"`
}

type FlavorListDetailRes struct {
	Flavors []FlavorInfoDetail `json:"flavors"`
}

type FlavorShowRes struct {
	Flavor FlavorInfoDetail `json:"flavor"`
}

type FlavorExtraSpecsCreateReq struct {
	ExtraSpecs map[string]string `json:"extra_specs"`
}

type FlavorExtraSpecsListRes struct {
	ExtraSpecs map[string]string `json:"extra_specs"`
}

type FlavorInfo struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Links []rest.LinkInfo `json:"links"`
}

type FlavorInfoDetail struct {
	FlavorInfo `json:",inline"`
	RAM        uint64  `json:"ram"`
	Disk       uint64  `json:"disk"`
	VCpus      uint64  `json:"vcpus"`
	Ephemeral  uint64  `json:"OS-FLV-EXT-DATA:ephemeral"`
	Disabled   bool    `json:"OS-FLV-DISABLED:disabled"`
	Swap       uint64  `json:"swap"`
	RxTxFactor float64 `json:"rxtx_factor"`
	Public     bool    `json:"os-flavor-access:is_public"`
}

func (svc *service) commonFlavorList(c *gin.Context) ([]v1.Flavor, bool) {
	//sortKeys := c.QueryArray("sort_key")
	//sortDirs := c.QueryArray("sort_dir")
	marker := c.Query("marker")

	filterMinRam, minRam := GetFilterUInt(c, "minRam")
	filterMinDisk, minDisk := GetFilterUInt(c, "minDisk")
	minDisk = minDisk * 1024 // GB -> MB
	filterLimit, limit := GetFilterUInt(c, "limit")
	// XXX Disallow unless user == admin
	filterPublic, public := GetFilterBool(c, "isPublic")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavors, err := clnt.List()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return []v1.Flavor{}, true
	}

	res := []v1.Flavor{}

	// XXX Links field
	count := uint64(0)
	seenMarker := false
	if marker == "" {
		seenMarker = true
	}
	for _, flv := range flavors.Items {
		if marker != "" {
			if marker == flv.Spec.ID {
				seenMarker = true
				marker = ""
				continue
			}
		}
		if !seenMarker {
			continue
		}
		if filterPublic && flv.Spec.Public != public {
			continue
		}
		if filterMinRam && flv.Spec.Resources.MemoryMB < minRam {
			continue
		}
		if filterMinDisk && flv.Spec.Resources.RootDiskMB < minDisk {
			continue
		}
		res = append(res, flv)
		count = count + 1
		if filterLimit && count >= limit {
			break
		}
	}
	return res, false
}

func (svc *service) FlavorList(c *gin.Context) {
	flavors, failed := svc.commonFlavorList(c)
	if failed {
		return
	}

	res := &FlavorListRes{
		Flavors: []FlavorInfo{},
	}

	// XXX Links field
	for _, flavor := range flavors {
		res.Flavors = append(res.Flavors, FlavorInfo{
			ID:   flavor.Spec.ID,
			Name: flavor.ObjectMeta.Name,
		})
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorListDetail(c *gin.Context) {
	flavors, failed := svc.commonFlavorList(c)
	if failed {
		return
	}

	res := FlavorListDetailRes{
		Flavors: []FlavorInfoDetail{},
	}

	// XXX Links field
	for _, flavor := range flavors {
		res.Flavors = append(res.Flavors, FlavorInfoDetail{
			FlavorInfo: FlavorInfo{
				ID:   flavor.Spec.ID,
				Name: flavor.ObjectMeta.Name,
			},
			RAM:        flavor.Spec.Resources.MemoryMB,
			Disk:       flavor.Spec.Resources.RootDiskMB / 1024,
			VCpus:      flavor.Spec.Resources.CPUCount,
			Ephemeral:  flavor.Spec.Resources.EphemeralDiskMB / 1024,
			Disabled:   flavor.Spec.Disabled,
			Swap:       flavor.Spec.Resources.SwapDiskMB,
			RxTxFactor: flavor.Spec.Resources.RxTxFactor,
			Public:     flavor.Spec.Public,
		})
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorShow(c *gin.Context) {
	id := c.Param("id")

	if id == "detail" {
		svc.FlavorListDetail(c)
		return
	}

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	// XXX Links field
	res := FlavorShowRes{
		Flavor: FlavorInfoDetail{
			FlavorInfo: FlavorInfo{
				ID:   flavor.Spec.ID,
				Name: flavor.ObjectMeta.Name,
			},
			RAM:        flavor.Spec.Resources.MemoryMB,
			Disk:       flavor.Spec.Resources.RootDiskMB / 1024,
			VCpus:      flavor.Spec.Resources.CPUCount,
			Ephemeral:  flavor.Spec.Resources.EphemeralDiskMB / 1024,
			Disabled:   flavor.Spec.Disabled,
			Swap:       flavor.Spec.Resources.SwapDiskMB,
			RxTxFactor: flavor.Spec.Resources.RxTxFactor,
			Public:     flavor.Spec.Public,
		},
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorDelete(c *gin.Context) {
	id := c.Param("id")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	err = clnt.Delete(flavor.ObjectMeta.Name, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "")
}

func (svc *service) FlavorCreate(c *gin.Context) {
	req := FlavorCreateReq{
		Flavor: FlavorInfoDetail{
			FlavorInfo: FlavorInfo{
				ID: string(uuid.NewUUID()),
			},
			Disk:       0,
			Ephemeral:  0,
			Swap:       0,
			Disabled:   false,
			Public:     true,
			VCpus:      1,
			RAM:        256,
			RxTxFactor: 1.0,
		},
	}
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(req.Flavor.ID)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	flavor = &v1.Flavor{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Flavor.Name,
		},
		Spec: v1.FlavorSpec{
			ID:       req.Flavor.ID,
			Disabled: req.Flavor.Disabled,
			Public:   req.Flavor.Public,
			Resources: v1.FlavorResources{
				EphemeralDiskMB: req.Flavor.Ephemeral * 1024,
				RootDiskMB:      req.Flavor.Disk * 1024,
				SwapDiskMB:      req.Flavor.Swap,
				MemoryMB:        req.Flavor.RAM,
				CPUCount:        req.Flavor.VCpus,
				RxTxFactor:      req.Flavor.RxTxFactor,
			},
		},
	}

	flavor, err = clnt.Create(flavor)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// XXX Links field
	res := FlavorShowRes{
		Flavor: FlavorInfoDetail{
			FlavorInfo: FlavorInfo{
				ID:   flavor.Spec.ID,
				Name: flavor.ObjectMeta.Name,
			},
			RAM:        flavor.Spec.Resources.MemoryMB,
			Disk:       flavor.Spec.Resources.RootDiskMB / 1024,
			VCpus:      flavor.Spec.Resources.CPUCount,
			Ephemeral:  flavor.Spec.Resources.EphemeralDiskMB / 1024,
			Disabled:   flavor.Spec.Disabled,
			Swap:       flavor.Spec.Resources.SwapDiskMB,
			RxTxFactor: flavor.Spec.Resources.RxTxFactor,
			Public:     flavor.Spec.Public,
		},
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorShowExtraSpecs(c *gin.Context) {
	id := c.Param("id")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	res := FlavorExtraSpecsListRes{
		ExtraSpecs: flavor.Spec.ExtraSpecs,
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorCreateExtraSpecs(c *gin.Context) {
	req := FlavorExtraSpecsCreateReq{}
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Param("id")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	for key, val := range req.ExtraSpecs {
		flavor.Spec.ExtraSpecs[key] = val
	}

	flavor, err = clnt.Update(flavor)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := FlavorExtraSpecsListRes{
		ExtraSpecs: flavor.Spec.ExtraSpecs,
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorShowExtraSpec(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	val, ok := flavor.Spec.ExtraSpecs[key]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	res := map[string]string{
		key: val,
	}

	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorCreateExtraSpec(c *gin.Context) {
	req := map[string]string{}
	err := c.BindJSON(&req)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := c.Param("id")
	key := c.Param("key")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	val, ok := req[key]
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	flavor.Spec.ExtraSpecs[key] = val

	flavor, err = clnt.Update(flavor)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := map[string]string{
		key: val,
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) FlavorDeleteExtraSpec(c *gin.Context) {
	id := c.Param("id")
	key := c.Param("key")

	clnt := compute.NewFlavorClient(svc.ComputeClient, v1.NamespaceSystem)

	flavor, err := clnt.GetByID(id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	delete(flavor.Spec.ExtraSpecs, key)

	flavor, err = clnt.Update(flavor)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "")
}

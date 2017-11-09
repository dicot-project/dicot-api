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

package image

import (
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/dicot-project/dicot-api/pkg/api/image/v1"
)

var (
	IMAGE_VISIBILITY_PUBLIC    = "public"
	IMAGE_VISIBILITY_COMMUNITY = "community"
	IMAGE_VISIBILITY_SHARED    = "shared"
	IMAGE_VISIBILITY_PRIVATE   = "private"

	IMAGE_CONTAINER_FORMAT_AMI    = "ami"
	IMAGE_CONTAINER_FORMAT_ARI    = "ari"
	IMAGE_CONTAINER_FORMAT_AKI    = "aki"
	IMAGE_CONTAINER_FORMAT_BARE   = "bare"
	IMAGE_CONTAINER_FORMAT_OVF    = "ovf"
	IMAGE_CONTAINER_FORMAT_OVA    = "ova"
	IMAGE_CONTAINER_FORMAT_DOCKER = "docker"

	IMAGE_DISK_FORMAT_AMI   = "ami"
	IMAGE_DISK_FORMAT_ARI   = "ari"
	IMAGE_DISK_FORMAT_AKI   = "aki"
	IMAGE_DISK_FORMAT_VHD   = "vhd"
	IMAGE_DISK_FORMAT_VHDX  = "vhdx"
	IMAGE_DISK_FORMAT_VMDK  = "vmdk"
	IMAGE_DISK_FORMAT_RAW   = "raw"
	IMAGE_DISK_FORMAT_QCOW2 = "qcow2"
	IMAGE_DISK_FORMAT_VDI   = "vdi"
	IMAGE_DISK_FORMAT_PLOOP = "ploop"
	IMAGE_DISK_FORMAT_ISO   = "iso"

	IMAGE_STATUS_QUEUED         = "queued"
	IMAGE_STATUS_SAVING         = "saving"
	IMAGE_STATUS_ACTIVE         = "active"
	IMAGE_STATUS_KILLED         = "killed"
	IMAGE_STATUS_DELETED        = "deleted"
	IMAGE_STATUS_PENDING_DELETE = "pending_delete"
	IMAGE_STATUS_DEACTIVATED    = "deactivated"
)

func IsValidVisibility(vis string) bool {
	switch vis {
	case IMAGE_VISIBILITY_PUBLIC,
		IMAGE_VISIBILITY_COMMUNITY,
		IMAGE_VISIBILITY_SHARED,
		IMAGE_VISIBILITY_PRIVATE:
		return true
	}
	return false
}

func IsValidContainerFormat(fmt string) bool {
	switch fmt {
	// Clouds are not required to support all container formats,
	// so we'll reject any we don't care about for KVM yet
	case IMAGE_CONTAINER_FORMAT_AMI,
		IMAGE_CONTAINER_FORMAT_AKI,
		IMAGE_CONTAINER_FORMAT_ARI,
		IMAGE_CONTAINER_FORMAT_BARE:
		return true
	}
	return false
}

func IsValidDiskFormat(fmt string) bool {
	switch fmt {
	// Clouds are not required to support all container formats,
	// so we'll reject any we don't care about for KVM yet
	case IMAGE_DISK_FORMAT_AMI,
		IMAGE_DISK_FORMAT_AKI,
		IMAGE_DISK_FORMAT_ARI,
		IMAGE_DISK_FORMAT_QCOW2,
		IMAGE_DISK_FORMAT_RAW,
		IMAGE_DISK_FORMAT_ISO:
		return true
	}
	return false
}

func NewImageClient(cl rest.Interface, namespace string) ImageInterface {
	return &images{cl: cl, ns: namespace}
}

type ImageInterface interface {
	Create(obj *v1.Image) (*v1.Image, error)
	Update(obj *v1.Image) (*v1.Image, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.Image, error)
	GetByID(id string) (*v1.Image, error)
	List() (*v1.ImageList, error)
	NewListWatch() *cache.ListWatch
}

type images struct {
	cl rest.Interface
	ns string
}

func (f *images) Create(obj *v1.Image) (*v1.Image, error) {
	var result v1.Image
	err := f.cl.Post().
		Namespace(f.ns).Resource("images").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *images) Update(obj *v1.Image) (*v1.Image, error) {
	var result v1.Image
	name := obj.GetObjectMeta().GetName()
	err := f.cl.Put().
		Namespace(f.ns).Resource("images").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *images) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource("images").
		Name(name).Body(options).Do().
		Error()
}

func (f *images) Get(name string) (*v1.Image, error) {
	var result v1.Image
	err := f.cl.Get().
		Namespace(f.ns).Resource("images").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *images) GetByID(id string) (*v1.Image, error) {
	list, err := f.List()
	if err != nil {
		return nil, err
	}
	for _, image := range list.Items {
		if image.Spec.ID == id {
			return &image, nil
		}
	}

	return nil, errors.NewNotFound(v1.Resource("image"), id)
}

func (f *images) List() (*v1.ImageList, error) {
	var result v1.ImageList
	err := f.cl.Get().
		Namespace(f.ns).Resource("images").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *images) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, "images", f.ns, fields.Everything())
}

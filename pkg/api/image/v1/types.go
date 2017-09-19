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

package v1

import (
	"k8s.io/apimachinery/pkg/apimachinery/announced"
	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

const GroupName = "image.dicot.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie(GroupVersion.String())
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Image{},
		&ImageList{},
	)
	return nil
}

func init() {
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:              GroupName,
			VersionPreferenceOrder: []string{GroupVersion.Version},
			ImportPrefix:           "dicot.io/dicot/pkg/api/image/v1",
		},
		announced.VersionToSchemeFunc{
			GroupVersion.Version: SchemeBuilder.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme.Scheme); err != nil {
		panic(err)
	}
}

type Image struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec            ImageSpec         `json:"spec,omitempty" valid:"required"`
}

type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	ListMeta        metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image         `json:"items"`
}

type ImageSpec struct {
	ID              string            `json:"id"`
	Name            *string           `json:"name"`
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
	Metadata        map[string]string `json:"metadata"`

	/* XXX some how reference where it is stored - PVC ? */
}

func (v *Image) GetObjectKind() schema.ObjectKind {
	return &v.TypeMeta
}

func (v *Image) GetObjectMeta() metav1.Object {
	return &v.ObjectMeta
}

func (vl *ImageList) GetObjectKind() schema.ObjectKind {
	return &vl.TypeMeta
}

func (vl *ImageList) GetListMeta() metav1.List {
	return &vl.ListMeta
}

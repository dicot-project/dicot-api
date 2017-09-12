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

const GroupName = "dicot.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

var FlavorGroupVersionKind = schema.GroupVersionKind{Group: GroupName, Version: GroupVersion.Version, Kind: "Flavor"}

var (
	groupFactoryRegistry = make(announced.APIGroupFactoryRegistry)
	registry             = registered.NewOrDie(GroupVersion.String())
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Flavor{},
		&FlavorList{},
	)
	return nil
}

func init() {
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:              GroupName,
			VersionPreferenceOrder: []string{GroupVersion.Version},
			ImportPrefix:           "dicot.io/dicot/pkg/api/v1",
		},
		announced.VersionToSchemeFunc{
			GroupVersion.Version: SchemeBuilder.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme.Scheme); err != nil {
		panic(err)
	}
}

type Flavor struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec            FlavorSpec        `json:"spec,omitempty" valid:"required"`
}

type FlavorList struct {
	metav1.TypeMeta `json:",inline"`
	ListMeta        metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Flavor        `json:"items"`
}

type FlavorSpec struct {
	ID         string            `json:"id"`
	Disabled   bool              `json:"disabled"`
	Public     bool              `json:"public"`
	Resources  FlavorResources   `json:"resources"`
	ExtraSpecs map[string]string `json:"extra_specs"`
}

type FlavorResources struct {
	EphemeralDiskMB uint64  `json:"ephemeral_disk_mb"`
	RootDiskMB      uint64  `json:"root_disk_mb"`
	SwapDiskMB      uint64  `json:"swap_disk_mb"`
	MemoryMB        uint64  `json:"memory_mb"`
	CPUCount        uint64  `json:"cpu_count"`
	RxTxFactor      float64 `json:"rxtx_factor"`
}

func (v *Flavor) GetObjectKind() schema.ObjectKind {
	return &v.TypeMeta
}

func (v *Flavor) GetObjectMeta() metav1.Object {
	return &v.ObjectMeta
}

func (vl *FlavorList) GetObjectKind() schema.ObjectKind {
	return &vl.TypeMeta
}

func (vl *FlavorList) GetListMeta() metav1.List {
	return &vl.ListMeta
}

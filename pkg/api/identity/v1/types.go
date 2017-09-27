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

const NamespaceSystem = "dicot-system"

const GroupName = "identity.dicot.io"

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
		&Project{},
		&ProjectList{},
		&User{},
		&UserList{},
	)
	return nil
}

func init() {
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName:              GroupName,
			VersionPreferenceOrder: []string{GroupVersion.Version},
			ImportPrefix:           "dicot.io/dicot/pkg/api/identity/v1",
		},
		announced.VersionToSchemeFunc{
			GroupVersion.Version: SchemeBuilder.AddToScheme,
		},
	).Announce(groupFactoryRegistry).RegisterAndEnable(registry, scheme.Scheme); err != nil {
		panic(err)
	}
}

type Project struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec            ProjectSpec       `json:"spec,omitempty" valid:"required"`
}

type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	ListMeta        metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project       `json:"items"`
}

type ProjectSpec struct {
	Parent      string `json:"parent"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Namespace   string `json:"namespace"`
}

func (v *Project) GetObjectKind() schema.ObjectKind {
	return &v.TypeMeta
}

func (v *Project) GetObjectMeta() metav1.Object {
	return &v.ObjectMeta
}

func (vl *ProjectList) GetObjectKind() schema.ObjectKind {
	return &vl.TypeMeta
}

func (vl *ProjectList) GetListMeta() metav1.List {
	return &vl.ListMeta
}

type User struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec            UserSpec          `json:"spec,omitempty" valid:"required"`
}

type UserList struct {
	metav1.TypeMeta `json:",inline"`
	ListMeta        metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User          `json:"items"`
}

type UserSpec struct {
	Name             string       `json:"name"`
	DomainID         string       `json:"domain_id"`
	Enabled          bool         `json:"enabled"`
	DefaultProjectID string       `json:"default_project_id"`
	Password         UserPassword `json:"password"`
	Description      string       `json:"description"`
	EMail            string       `json:"email"`
}

type UserPassword struct {
	SecretRef string `json:"secretRef"`
	ExpiresAt string `json:"expiresAt"`
}

func (v *User) GetObjectKind() schema.ObjectKind {
	return &v.TypeMeta
}

func (v *User) GetObjectMeta() metav1.Object {
	return &v.ObjectMeta
}

func (vl *UserList) GetObjectKind() schema.ObjectKind {
	return &vl.TypeMeta
}

func (vl *UserList) GetListMeta() metav1.List {
	return &vl.ListMeta
}

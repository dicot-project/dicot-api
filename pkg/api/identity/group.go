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

package identity

import (
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
)

func NewGroupClient(cl *rest.RESTClient, namespace string) *GroupClient {
	return &GroupClient{cl: cl, ns: namespace}
}

type GroupClient struct {
	cl *rest.RESTClient
	ns string
}

func (pc *GroupClient) Create(obj *v1.Group) (*v1.Group, error) {
	var result v1.Group
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("groups").
		Body(obj).Do().Into(&result)
	return &result, err
}

func (pc *GroupClient) Update(obj *v1.Group) (*v1.Group, error) {
	var result v1.Group
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("groups").
		Name(name).Body(obj).Do().Into(&result)
	return &result, err
}

func (pc *GroupClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("groups").
		Name(name).Body(options).Do().
		Error()
}

func (pc *GroupClient) Get(name string) (*v1.Group, error) {
	var result v1.Group
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("groups").
		Name(name).Do().Into(&result)
	return &result, err
}

func (pc *GroupClient) GetByUID(uid string) (*v1.Group, error) {
	list, err := pc.List()
	if err != nil {
		return nil, err
	}
	for _, group := range list.Items {
		if string(group.ObjectMeta.UID) == uid {
			return &group, nil
		}
	}
	return nil, errors.NewNotFound(v1.Resource("group"), uid)
}

func (pc *GroupClient) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *GroupClient) List() (*v1.GroupList, error) {
	var result v1.GroupList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("groups").
		Do().Into(&result)
	return &result, err
}

func (pc *GroupClient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "groups", pc.ns, fields.Everything())
}

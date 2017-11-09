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

func NewGroupClient(cl rest.Interface, namespace string) GroupInterface {
	return &groups{cl: cl, ns: namespace}
}

type groups struct {
	cl rest.Interface
	ns string
}

type GroupGetter interface {
	Groups(namespace string) GroupInterface
}

type GroupInterface interface {
	Create(obj *v1.Group) (*v1.Group, error)
	Update(obj *v1.Group) (*v1.Group, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.Group, error)
	GetByUID(id string) (*v1.Group, error)
	Exists(name string) (bool, error)
	List() (*v1.GroupList, error)
	NewListWatch() *cache.ListWatch
}

func (pc *groups) Create(obj *v1.Group) (*v1.Group, error) {
	var result v1.Group
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("groups").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *groups) Update(obj *v1.Group) (*v1.Group, error) {
	var result v1.Group
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("groups").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *groups) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("groups").
		Name(name).Body(options).Do().
		Error()
}

func (pc *groups) Get(name string) (*v1.Group, error) {
	var result v1.Group
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("groups").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *groups) GetByUID(uid string) (*v1.Group, error) {
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

func (pc *groups) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *groups) List() (*v1.GroupList, error) {
	var result v1.GroupList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("groups").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *groups) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "groups", pc.ns, fields.Everything())
}

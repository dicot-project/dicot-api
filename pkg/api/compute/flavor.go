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

package compute

import (
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/dicot-project/dicot-api/pkg/api/compute/v1"
)

func NewFlavorClient(cl rest.Interface, namespace string) FlavorInterface {
	return &flavors{cl: cl, ns: namespace}
}

type flavors struct {
	cl rest.Interface
	ns string
}

type FlavorGetter interface {
	Flavors(namespace string) FlavorInterface
}

type FlavorInterface interface {
	Create(obj *v1.Flavor) (*v1.Flavor, error)
	Update(obj *v1.Flavor) (*v1.Flavor, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.Flavor, error)
	GetByID(id string) (*v1.Flavor, error)
	List() (*v1.FlavorList, error)
	NewListWatch() *cache.ListWatch
}

func (f *flavors) Create(obj *v1.Flavor) (*v1.Flavor, error) {
	var result v1.Flavor
	err := f.cl.Post().
		Namespace(f.ns).Resource("flavors").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *flavors) Update(obj *v1.Flavor) (*v1.Flavor, error) {
	var result v1.Flavor
	name := obj.GetObjectMeta().GetName()
	err := f.cl.Put().
		Namespace(f.ns).Resource("flavors").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *flavors) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource("flavors").
		Name(name).Body(options).Do().
		Error()
}

func (f *flavors) Get(name string) (*v1.Flavor, error) {
	var result v1.Flavor
	err := f.cl.Get().
		Namespace(f.ns).Resource("flavors").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *flavors) GetByID(id string) (*v1.Flavor, error) {
	list, err := f.List()
	if err != nil {
		return nil, err
	}
	for _, flv := range list.Items {
		if flv.Spec.ID == id {
			return &flv, nil
		}
	}

	return nil, errors.NewNotFound(v1.Resource("flavor"), id)
}

func (f *flavors) List() (*v1.FlavorList, error) {
	var result v1.FlavorList
	err := f.cl.Get().
		Namespace(f.ns).Resource("flavors").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (f *flavors) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, "flavors", f.ns, fields.Everything())
}

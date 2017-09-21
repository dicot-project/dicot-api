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

package api

import (
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/dicot-project/dicot-api/pkg/api/v1"
)

func NewFlavorClient(cl *rest.RESTClient, namespace string) *FlavorClient {
	return &FlavorClient{cl: cl, ns: namespace}
}

type FlavorClient struct {
	cl *rest.RESTClient
	ns string
}

func (f *FlavorClient) Create(obj *v1.Flavor) (*v1.Flavor, error) {
	var result v1.Flavor
	err := f.cl.Post().
		Namespace(f.ns).Resource("flavors").
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *FlavorClient) Update(obj *v1.Flavor) (*v1.Flavor, error) {
	var result v1.Flavor
	name := obj.GetObjectMeta().GetName()
	err := f.cl.Put().
		Namespace(f.ns).Resource("flavors").
		Name(name).Body(obj).Do().Into(&result)
	return &result, err
}

func (f *FlavorClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource("flavors").
		Name(name).Body(options).Do().
		Error()
}

func (f *FlavorClient) Get(name string) (*v1.Flavor, error) {
	var result v1.Flavor
	err := f.cl.Get().
		Namespace(f.ns).Resource("flavors").
		Name(name).Do().Into(&result)
	return &result, err
}

func (f *FlavorClient) GetByID(id string) (*v1.Flavor, error) {
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

func (f *FlavorClient) List() (*v1.FlavorList, error) {
	var result v1.FlavorList
	err := f.cl.Get().
		Namespace(f.ns).Resource("flavors").
		Do().Into(&result)
	return &result, err
}

func (f *FlavorClient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, "flavors", f.ns, fields.Everything())
}

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

func NewKeypairClient(cl *rest.RESTClient, namespace string) *KeypairClient {
	return &KeypairClient{cl: cl, ns: namespace}
}

type KeypairClient struct {
	cl *rest.RESTClient
	ns string
}

func (kpc *KeypairClient) Create(obj *v1.Keypair) (*v1.Keypair, error) {
	var result v1.Keypair
	err := kpc.cl.Post().
		Namespace(kpc.ns).Resource("keypairs").
		Body(obj).Do().Into(&result)
	return &result, err
}

func (kpc *KeypairClient) Update(obj *v1.Keypair) (*v1.Keypair, error) {
	var result v1.Keypair
	name := obj.GetObjectMeta().GetName()
	err := kpc.cl.Put().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Body(obj).Do().Into(&result)
	return &result, err
}

func (kpc *KeypairClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return kpc.cl.Delete().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Body(options).Do().
		Error()
}

func (kpc *KeypairClient) Get(name string) (*v1.Keypair, error) {
	var result v1.Keypair
	err := kpc.cl.Get().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Do().Into(&result)
	return &result, err
}

func (kpc *KeypairClient) Exists(name string) (bool, error) {
	_, err := kpc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (kpc *KeypairClient) List() (*v1.KeypairList, error) {
	var result v1.KeypairList
	err := kpc.cl.Get().
		Namespace(kpc.ns).Resource("keypairs").
		Do().Into(&result)
	return &result, err
}

func (kpc *KeypairClient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(kpc.cl, "keypairs", kpc.ns, fields.Everything())
}

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

func NewKeypairClient(cl rest.Interface, namespace string) KeypairInterface {
	return &keypairs{cl: cl, ns: namespace}
}

type keypairs struct {
	cl rest.Interface
	ns string
}

type KeypairGetter interface {
	Keypairs(namespace string) KeypairInterface
}

type KeypairInterface interface {
	Create(obj *v1.Keypair) (*v1.Keypair, error)
	Update(obj *v1.Keypair) (*v1.Keypair, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.Keypair, error)
	Exists(name string) (bool, error)
	List() (*v1.KeypairList, error)
	NewListWatch() *cache.ListWatch
}

func (kpc *keypairs) Create(obj *v1.Keypair) (*v1.Keypair, error) {
	var result v1.Keypair
	err := kpc.cl.Post().
		Namespace(kpc.ns).Resource("keypairs").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (kpc *keypairs) Update(obj *v1.Keypair) (*v1.Keypair, error) {
	var result v1.Keypair
	name := obj.GetObjectMeta().GetName()
	err := kpc.cl.Put().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (kpc *keypairs) Delete(name string, options *meta_v1.DeleteOptions) error {
	return kpc.cl.Delete().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Body(options).Do().
		Error()
}

func (kpc *keypairs) Get(name string) (*v1.Keypair, error) {
	var result v1.Keypair
	err := kpc.cl.Get().
		Namespace(kpc.ns).Resource("keypairs").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (kpc *keypairs) Exists(name string) (bool, error) {
	_, err := kpc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (kpc *keypairs) List() (*v1.KeypairList, error) {
	var result v1.KeypairList
	err := kpc.cl.Get().
		Namespace(kpc.ns).Resource("keypairs").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (kpc *keypairs) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(kpc.cl, "keypairs", kpc.ns, fields.Everything())
}

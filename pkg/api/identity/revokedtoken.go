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

func NewRevokedTokenClient(cl rest.Interface, namespace string) RevokedTokenInterface {
	return &revokedTokens{cl: cl, ns: namespace}
}

type revokedTokens struct {
	cl rest.Interface
	ns string
}

type RevokedTokenInterface interface {
	Create(obj *v1.RevokedToken) (*v1.RevokedToken, error)
	Update(obj *v1.RevokedToken) (*v1.RevokedToken, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.RevokedToken, error)
	GetByUID(id string) (*v1.RevokedToken, error)
	Exists(name string) (bool, error)
	List() (*v1.RevokedTokenList, error)
	NewListWatch() *cache.ListWatch
}

func (pc *revokedTokens) Create(obj *v1.RevokedToken) (*v1.RevokedToken, error) {
	var result v1.RevokedToken
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("revokedtokens").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *revokedTokens) Update(obj *v1.RevokedToken) (*v1.RevokedToken, error) {
	var result v1.RevokedToken
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("revokedtokens").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *revokedTokens) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("revokedtokens").
		Name(name).Body(options).Do().
		Error()
}

func (pc *revokedTokens) Get(name string) (*v1.RevokedToken, error) {
	var result v1.RevokedToken
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("revokedtokens").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *revokedTokens) GetByUID(uid string) (*v1.RevokedToken, error) {
	list, err := pc.List()
	if err != nil {
		return nil, err
	}
	for _, revokedtoken := range list.Items {
		if string(revokedtoken.ObjectMeta.UID) == uid {
			return &revokedtoken, nil
		}
	}
	return nil, errors.NewNotFound(v1.Resource("revokedtoken"), uid)
}

func (pc *revokedTokens) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *revokedTokens) List() (*v1.RevokedTokenList, error) {
	var result v1.RevokedTokenList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("revokedtokens").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *revokedTokens) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "revokedtokens", pc.ns, fields.Everything())
}

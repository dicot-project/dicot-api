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

func NewUserClient(cl rest.Interface, namespace string) UserInterface {
	return &users{cl: cl, ns: namespace}
}

type users struct {
	cl rest.Interface
	ns string
}

type UserInterface interface {
	Create(obj *v1.User) (*v1.User, error)
	Update(obj *v1.User) (*v1.User, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.User, error)
	GetByUID(id string) (*v1.User, error)
	Exists(name string) (bool, error)
	List() (*v1.UserList, error)
	NewListWatch() *cache.ListWatch
}

func (pc *users) Create(obj *v1.User) (*v1.User, error) {
	var result v1.User
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("users").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *users) Update(obj *v1.User) (*v1.User, error) {
	var result v1.User
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("users").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *users) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("users").
		Name(name).Body(options).Do().
		Error()
}

func (pc *users) Get(name string) (*v1.User, error) {
	var result v1.User
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("users").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *users) GetByUID(uid string) (*v1.User, error) {
	list, err := pc.List()
	if err != nil {
		return nil, err
	}
	for _, user := range list.Items {
		if string(user.ObjectMeta.UID) == uid {
			return &user, nil
		}
	}
	return nil, errors.NewNotFound(v1.Resource("user"), uid)
}

func (pc *users) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *users) List() (*v1.UserList, error) {
	var result v1.UserList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("users").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *users) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "users", pc.ns, fields.Everything())
}

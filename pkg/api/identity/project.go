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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/dicot-project/dicot-api/pkg/api/identity/v1"
)

func NewProjectClient(cl rest.Interface, namespace string) ProjectInterface {
	return &projects{cl: cl, ns: namespace}
}

func FormatProjectNamespace(domainName, projectName string) string {
	return fmt.Sprintf("dicot-project-%s-%s", domainName, projectName)
}

func FormatDomainNamespace(domainName string) string {
	return fmt.Sprintf("dicot-domain-%s", domainName)
}

type projects struct {
	cl rest.Interface
	ns string
}

type ProjectGetter interface {
	Projects(namespace string) ProjectInterface
}

type ProjectInterface interface {
	Create(obj *v1.Project) (*v1.Project, error)
	Update(obj *v1.Project) (*v1.Project, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	Get(name string) (*v1.Project, error)
	GetByUID(id string) (*v1.Project, error)
	Exists(name string) (bool, error)
	List() (*v1.ProjectList, error)
	NewListWatch() *cache.ListWatch
}

func (pc *projects) Create(obj *v1.Project) (*v1.Project, error) {
	var result v1.Project
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("projects").
		Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *projects) Update(obj *v1.Project) (*v1.Project, error) {
	var result v1.Project
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("projects").
		Name(name).Body(obj).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *projects) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("projects").
		Name(name).Body(options).Do().
		Error()
}

func (pc *projects) Get(name string) (*v1.Project, error) {
	var result v1.Project
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("projects").
		Name(name).Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *projects) GetByUID(uid string) (*v1.Project, error) {
	list, err := pc.List()
	if err != nil {
		return nil, err
	}
	for _, project := range list.Items {
		if string(project.ObjectMeta.UID) == uid {
			return &project, nil
		}
	}
	return nil, errors.NewNotFound(v1.Resource("project"), uid)
}

func (pc *projects) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *projects) List() (*v1.ProjectList, error) {
	var result v1.ProjectList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("projects").
		Do().Into(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil

}

func (pc *projects) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "projects", pc.ns, fields.Everything())
}

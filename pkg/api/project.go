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
	"fmt"

	"github.com/dicot-project/dicot-api/pkg/api/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func NewProjectClient(cl *rest.RESTClient, namespace string) *ProjectClient {
	return &ProjectClient{cl: cl, ns: namespace}
}

func FormatProjectNamespace(domainName, projectName string) string {
	return fmt.Sprintf("dicot-project-%s-%s", domainName, projectName)
}

func FormatDomainNamespace(domainName string) string {
	return fmt.Sprintf("dicot-domain-%s", domainName)
}

type ProjectClient struct {
	cl *rest.RESTClient
	ns string
}

func (pc *ProjectClient) Create(obj *v1.Project) (*v1.Project, error) {
	var result v1.Project
	err := pc.cl.Post().
		Namespace(pc.ns).Resource("projects").
		Body(obj).Do().Into(&result)
	return &result, err
}

func (pc *ProjectClient) Update(obj *v1.Project) (*v1.Project, error) {
	var result v1.Project
	name := obj.GetObjectMeta().GetName()
	err := pc.cl.Put().
		Namespace(pc.ns).Resource("projects").
		Name(name).Body(obj).Do().Into(&result)
	return &result, err
}

func (pc *ProjectClient) Delete(name string, options *meta_v1.DeleteOptions) error {
	return pc.cl.Delete().
		Namespace(pc.ns).Resource("projects").
		Name(name).Body(options).Do().
		Error()
}

func (pc *ProjectClient) Get(name string) (*v1.Project, error) {
	var result v1.Project
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("projects").
		Name(name).Do().Into(&result)
	return &result, err
}

func (pc *ProjectClient) GetByUID(uid string) (*v1.Project, error) {
	list, err := pc.List()
	if err != nil {
		return nil, err
	}
	for _, project := range list.Items {
		if string(project.ObjectMeta.UID) == uid {
			return &project, nil
		}
	}
	return nil, nil
}

func (pc *ProjectClient) Exists(name string) (bool, error) {
	_, err := pc.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pc *ProjectClient) List() (*v1.ProjectList, error) {
	var result v1.ProjectList
	err := pc.cl.Get().
		Namespace(pc.ns).Resource("projects").
		Do().Into(&result)
	return &result, err
}

func (pc *ProjectClient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(pc.cl, "projects", pc.ns, fields.Everything())
}

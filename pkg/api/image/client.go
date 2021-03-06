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

package image

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/dicot-project/dicot-api/pkg/api/image/v1"
)

type Interface interface {
	RESTClient() rest.Interface
	ImageGetter
}

type image struct {
	cl rest.Interface
}

func New(c *rest.Config) (Interface, error) {
	cCopy := *c
	cCopy.GroupVersion = &v1.GroupVersion
	cCopy.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	cCopy.APIPath = "/apis"
	cCopy.ContentType = runtime.ContentTypeJSON

	cl, err := rest.RESTClientFor(&cCopy)
	if err != nil {
		return nil, err
	}

	return &image{cl}, err
}

func (c *image) RESTClient() rest.Interface {
	return c.cl
}

func (c *image) Images(namespace string) ImageInterface {
	return NewImageClient(c.cl, namespace)
}

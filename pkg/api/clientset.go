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
	"github.com/dicot-project/dicot-api/pkg/api/compute"
	"github.com/dicot-project/dicot-api/pkg/api/identity"
	"github.com/dicot-project/dicot-api/pkg/api/image"

	"k8s.io/client-go/rest"
)

type Interface interface {
	Compute() compute.Interface
	Identity() identity.Interface
	Image() image.Interface
}

type clientset struct {
	compute  compute.Interface
	identity identity.Interface
	image    image.Interface
}

func (c *clientset) Compute() compute.Interface {
	return c.compute
}

func (c *clientset) Identity() identity.Interface {
	return c.identity
}

func (c *clientset) Image() image.Interface {
	return c.image
}

func NewClientset(c *rest.Config) (Interface, error) {
	cCopy := *c
	computeClient, err := compute.New(&cCopy)
	if err != nil {
		return nil, err
	}
	identityClient, err := identity.New(&cCopy)
	if err != nil {
		return nil, err
	}
	imageClient, err := image.New(&cCopy)
	if err != nil {
		return nil, err
	}
	return &clientset{
		computeClient,
		identityClient,
		imageClient,
	}, nil
}

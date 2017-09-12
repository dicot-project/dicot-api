# Dicot

[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/kubevirt/kubevirt.svg)](https://www.apache.org/licenses/LICENSE-2.0)

**Dicot** is provides an OpenStack API compatibility layer above the KubeVirt
and Kubernetes APIs.

**Note:** Dicot is a heavy work in progress.

# Introduction

The Kubernetes project provides a cluster based management service for
container workloads. The KubeVirt project builds on this to provide a
way to manage virtual machine workloads using the Kubernetes APIs.

Recognising the widespread adoption of OpenStack, the Dicot project
aims to provide an Openstack API compatibility layer. This should
allow existing tools written against OpenStack APIs to be run against
Kubernetes and KubeVirt.

The API compatibility does not aim to cover all OpenStack projects.
Initially it is just targetting the compute (Nova), identity (Keystone),
image (Glance), and block (Cinder) services. Assuming it provides a
faithful implementation of these APIs, the other OpenStack services
could in theory be run talking to the Dicot API.

## Submitting patches

When sending patches to the project, the submitter is required to certify that
they have the legal right to submit the code. This is achieved by adding a line

    Signed-off-by: Real Name <email@address.com>

to the bottom of every commit message. Existence of such a line certifies
that the submitter has complied with the Developer's Certificate of Origin 1.1,
(as defined in the file docs/developer-certificate-of-origin).

This line can be automatically added to a commit in the correct format, by
using the '-s' option to 'git commit'.

## License

Dicot is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright 2017

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

[//]: # (Reference links)
   [k8s]: https://kubernetes.io
   [kubevirt]: https://kubevirt.github.io
   [openstack]: https://openstack.org

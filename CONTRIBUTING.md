# Introduction

Let's start with the relationship between the three important components:

* **Kubernetes** is a container orchestration system, and is used to run
  containers on a cluster
* **KubeVirt** is an add-on which is installed on-top of Kubernetes, to be able
  to add basic virtualization functionality to Kubernetes.
* **Dicot** is an add-on which is installed on-top of KubeVirt, to provide a
  REST API that is compatible with the OpenStack APIs.

## Contributing to Dicot

Contributing to Dicot should be as simple as possible. Have a question? Want
to discuss something? Want to contribute something? Just open an
[Issue](https://github.com/dicot-project/dicot-api/issues), a [Pull
Request](https://github.com/dicot-project/dicot-api/pulls), or send a mail to our
[Google Group](https://groups.google.com/forum/#!forum/dicot-dev).

If you spot a bug or want to change something pretty simple, just go
ahead and open an Issue and/or a Pull Request, including your changes
at [dicot-project/dicot-api](https://github.com/dicot-project/dicot-api).

For bigger changes, please create a tracker Issue, describing what you want to
do. Then either as the first commit in a Pull Request, or as an independent
Pull Request, provide an **informal** design proposal of your intended changes.
The location for such propoals is [/docs](docs/) in the Dicot
core repository. Make sure that all your Pull Requests link back to the
relevant Issues.

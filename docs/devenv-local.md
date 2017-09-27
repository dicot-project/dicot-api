 Running a local development env
 ===============================

This document describes how to setup a development environment
running Kubernetes, KubeVirt and Dicot, with everything on the
same local host. This is not how a production deployment would
exist, but is good enough for the majority of day-to-day dev
and test. It assumes Fedora 26 or equivalently modern OS, so
that it has a recent version of QEMU, libvirt and the Golang
runtime version needed by Kubernetes.

If running this inside a virtual machine it is recommended
that the VM has a minimum of 8 GB of RAM, 4 vCPUs and 40GB
of disk storage. Enabling nested KVM is also recommended.

Software prerequisites
======================

Starting from a default Fedora 26 Server install the following
extra software needs installing and configuring

```bash
sudo dnf install etcd docker git libvirt-devel
sudo dnf install golang glide golang-googlecode-tools-goimports
sudo groupadd docker
sudo gpasswd -a $USER docker
sudo systemctl enable docker
sudo systemctl start docker
newgrp docker
```

XXXX 'setenforce 0'

Setting up Go
=============

A standard Go development setup is expected. IOW set the GOPATH
env variable to a dir where all the GIT checkouts will live

```bash
cat >> $HOME/.bashrc <<EOF
export GOPATH=\$HOME/dev
export PATH=\$GOPATH/bin:\$GOPATH/src/k8s.io/kubernetes/_output/bin:\$PATH
EOF
. $HOME/.bashrc
mkdir -p $GOPATH/{src,pkg,bin}
```

Setting up Kubernetes
=====================

Due to ongoing refactoring of Kubernetes it is not advised to
follow the master git branch. Dicot thus aims to be compatible
only with the most recent stable release branch. This is
currently the 1.7 stream. There is a race condition in this
branch that often causes DNS setup to fail, so we must also
cherry-pick a patch

```bash
cd $GOPATH/src
git clone git://github.com/kubernetes/kubernetes k8s.io/kubernetes
cd k8s.io/kubernetes
git checkout release-1.7
git cherry-pick -x 413ab26df92c3da66bc6eb60c1d1105f6ac267fc
```

When launching k8s a few options need configuring. We want its
API server running on the default IP address with a sensible
hostname set, rather than 127.0.0.1. We always want DNS enabled
and the ability to launch privileged containers

```bash
DEV=`ip -4 route| grep default | awk '{print $5}'`
IP=`ip -4 addr | grep $DEV | grep inet | awk '{print $2}' | sed -e 's,/.*,,'`
HOST=`hostname`

cat > k8s.sh <<EOF
#!/bin/sh

export API_HOST_IP=$IP
export KUBELET_HOST=$IP
export HOSTNAME_OVERRIDE=$HOST
export KUBE_ENABLE_CLUSTER_DNS=true
export ALLOW_PRIVILEGED=1
export ALLOW_SECURITY_CONTEXT=1
export ENABLE_RBAC=true

if [ -d _output/local/bin/linux/amd64/ ]
then
  ./hack/local-up-cluster.sh -o _output/local/bin/linux/amd64/
else
  ./hack/local-up-cluster.sh
fi
EOF
chmod +x k8s.sh
./k8s.sh
```

Once k8s is running it will remain in the foreground. To stop
later it just Ctrl-C.

In another terminal now setup the k8s client app

```bash
cd $HOME/.kube
ln -s /var/run/kubernetes/admin.kubeconfig config
```

It is now possible to query k8s

```bash
kubectl get --all-namespaces all
```


Setting up KubeVirt
===================

This will use the master branch of KubeVirt

```bash
cd $GOPATH/src
git clone git://github.com/kubevirt/kubevirt kubevirt.io/kubevirt
cd kubevirt.io/kubevirt
```

Configure it with the local host info and then build

```bash
DEV=`ip -4 route| grep default | awk '{print $5}'`
IP=`ip -4 addr | grep $DEV | grep inet | awk '{print $2}' | sed -e 's,/.*,,'`
HOST=`hostname`

cat > hack/config-local.sh <<EOF
master_ip=$IP
primary_nic=$DEV
primary_node_name=$HOST
docker_tag=latest
EOF
make manifests docker
```

Running kubevirt is now simply a matter of loading the k8s
manifests

```bash
for i in manifests/*.yaml
do
  kubectl create -f $i
done
```


Setting up Dicot
================

With KubeVirt and Kubernetes running the final step is get Dicot
itself:

```bash
cd $GOPATH/src
git clone git://github.com/dicot-project/dicot-api github.com/dicot-project/dicot-api
cd github.com/dicot-project/dicot-api
```

It can be built and launched with:

```bash
make
for i in manifests/*.yaml
do
  kubectl create -f $i
done

./bin/dicot-api --kubeconfig $HOME/.kube/config -d -v 1 --logtostderr
```

In this case

Using OpenStack
===============

The Dicot build put a config containing Keystone credentials
in conf/keystone_admin. This can be used with the standard
OpenStack client tool

```bash
sudo dnf install python-openstackclient
. conf/identity_admin
openstack flavor list
openstack flavor show m1.small
```

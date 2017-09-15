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

package v2_1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	k8sv1 "k8s.io/client-go/pkg/api/v1"
)

type HypervisorListRes struct {
	Hypervisors []HypervisorInfo `json:"hypervisors"`
	Links       []LinkInfo       `json:"hypervisor_links"`
}

type HypervisorInfo struct {
	ID       string `json:"id"`
	Hostname string `json:"hypervisor_hostname"`
	State    string `json:"state"`
	Status   string `json:"status"`
}

type HypervisorListDetailRes struct {
	Hypervisors []HypervisorInfoDetail `json:"hypervisors"`
	Links       []LinkInfo             `json:"hypervisor_links"`
}

type HypervisorShowRes struct {
	Hypervisor HypervisorInfoDetail `json:"hypervisor"`
}

type HypervisorInfoDetail struct {
	ID                 string      `json:"id"`
	Hostname           string      `json:"hypervisor_hostname"`
	Type               string      `json:"hypervisor_type"`
	Version            string      `json:"hypervisor_version"`
	State              string      `json:"state"`
	Status             string      `json:"status"`
	CPUInfo            CPUInfo     `json:"cpu_info"`
	CurrentWorkload    uint64      `json:"current_workload"`
	DiskAvailableLeast uint64      `json:"disk_available_least"`
	HostIP             string      `json:"host_ip"`
	FreeDiskGB         uint64      `json:"free_disk_gb"`
	FreeRamMB          uint64      `json:"free_ram_mb"`
	LocalGB            uint64      `json:"local_gb"`
	LocalGBUsed        uint64      `json:"local_gb_used"`
	MemoryMB           uint64      `json:"memory_mb"`
	MemoryMBUsed       uint64      `json:"memory_mb_used"`
	RunningVMs         uint64      `json:"running_vm"`
	VCPUs              uint64      `json:"vcpus"`
	VCPUsUsed          uint64      `json:"vcpus_used"`
	Service            ServiceInfo `json:"service"`
}

type ServiceInfo struct {
	ID             string  `json:"id"`
	Host           string  `json:"host"`
	DisabledReason *string `json:"disabled_reason"`
}

type CPUInfo struct {
	Arch     string          `json:"arch"`
	Model    string          `json:"model"`
	Vendor   string          `json:"vendor"`
	Features []string        `json:"features"`
	Topology CPUInfoTopology `json:"topology"`
}

type CPUInfoTopology struct {
	Cores   uint `json:"cores"`
	Threads uint `json:"threads"`
	Sockets uint `json:"sockets"`
}

func (svc *service) getHypervisorList(c *gin.Context) ([]k8sv1.Pod, error) {
	marker := c.Query("marker")
	filterLimit, limit := GetFilterUInt(c, "limit")

	selector, err := labels.Parse("daemon in (virt-handler)")
	if err != nil {
		return []k8sv1.Pod{}, err
	}

	pods, err := svc.Clientset.CoreV1().Pods(k8smetav1.NamespaceAll).List(
		k8smetav1.ListOptions{
			LabelSelector: selector.String()})
	if err != nil {
		return []k8sv1.Pod{}, err
	}

	res := []k8sv1.Pod{}

	count := uint64(0)
	seenMarker := false
	if marker == "" {
		seenMarker = true
	}
	// XXX Links field
	for _, pod := range pods.Items {
		if marker != "" {
			if marker == string(pod.ObjectMeta.UID) {
				seenMarker = true
				marker = ""
				continue
			}
		}
		if !seenMarker {
			continue
		}

		res = append(res, pod)

		count = count + 1
		if filterLimit && count >= limit {
			break
		}
	}

	return res, nil
}

func (svc *service) HypervisorList(c *gin.Context) {

	pods, err := svc.getHypervisorList(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	res := HypervisorListRes{}

	for _, pod := range pods {
		state := "down"
		if pod.Status.Phase == "Running" {
			state = "up"
		}

		res.Hypervisors = append(res.Hypervisors, HypervisorInfo{
			Hostname: pod.Spec.NodeName,
			ID:       string(pod.ObjectMeta.UID),
			Status:   "enabled",
			State:    state,
		})
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) HypervisorListDetails(c *gin.Context) {

	pods, err := svc.getHypervisorList(c)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	res := HypervisorListDetailRes{}

	for _, pod := range pods {
		state := "down"
		if pod.Status.Phase == "Running" {
			state = "up"
		}

		res.Hypervisors = append(res.Hypervisors, HypervisorInfoDetail{
			Hostname: pod.Spec.NodeName,
			ID:       string(pod.ObjectMeta.UID),
			Status:   "enabled",
			State:    state,
			Type:     "QEMU",
		})
	}
	c.JSON(http.StatusOK, res)
}

func (svc *service) HypervisorShow(c *gin.Context) {
	name := c.Param("name")

	if name == "detail" {
		svc.HypervisorListDetails(c)
		return
	}

	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + name)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	labelSelector, err := labels.Parse("daemon in (virt-handler)")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pods, err := svc.Clientset.CoreV1().Pods(k8smetav1.NamespaceAll).List(
		k8smetav1.ListOptions{
			LabelSelector: labelSelector.String(),
			FieldSelector: fieldSelector.String(),
		})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(pods.Items) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if len(pods.Items) > 1 {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	pod := pods.Items[0]

	state := "down"
	if pod.Status.Phase == "Running" {
		state = "up"
	}

	res := HypervisorShowRes{
		Hypervisor: HypervisorInfoDetail{
			Hostname: pod.Spec.NodeName,
			ID:       string(pod.ObjectMeta.UID),
			Status:   "enabled",
			State:    state,
			Type:     "QEMU",
		},
	}

	c.JSON(http.StatusOK, res)
}

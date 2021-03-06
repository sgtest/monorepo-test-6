/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cm

import (
	"github.com/golang/glog"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-15/pkg/api/v1"
)

type containerManagerStub struct{}

var _ ContainerManager = &containerManagerStub{}

func (cm *containerManagerStub) Start(_ *v1.Node, _ ActivePodsFunc) error {
	glog.V(2).Infof("Starting stub container manager")
	return nil
}

func (cm *containerManagerStub) SystemCgroupsLimit() v1.ResourceList {
	return v1.ResourceList{}
}

func (cm *containerManagerStub) GetNodeConfig() NodeConfig {
	return NodeConfig{}
}

func (cm *containerManagerStub) GetMountedSubsystems() *CgroupSubsystems {
	return &CgroupSubsystems{}
}

func (cm *containerManagerStub) GetQOSContainersInfo() QOSContainersInfo {
	return QOSContainersInfo{}
}

func (cm *containerManagerStub) UpdateQOSCgroups() error {
	return nil
}

func (cm *containerManagerStub) Status() Status {
	return Status{}
}

func (cm *containerManagerStub) GetNodeAllocatableReservation() v1.ResourceList {
	return nil
}

func (cm *containerManagerStub) NewPodContainerManager() PodContainerManager {
	return &podContainerManagerStub{}
}

func NewStubContainerManager() ContainerManager {
	return &containerManagerStub{}
}

/*
Copyright 2016 The Kubernetes Authors.

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

package priorities

import (
	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/pkg/api/v1"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/plugin/pkg/scheduler/schedulercache"
)

// priorityMetadata is a type that is passed as metadata for priority functions
type priorityMetadata struct {
	nonZeroRequest *schedulercache.Resource
	podTolerations []v1.Toleration
	affinity       *v1.Affinity
}

// PriorityMetadata is a MetadataProducer.  Node info can be nil.
func PriorityMetadata(pod *v1.Pod, nodeNameToInfo map[string]*schedulercache.NodeInfo) interface{} {
	// If we cannot compute metadata, just return nil
	if pod == nil {
		return nil
	}
	tolerationsPreferNoSchedule := getAllTolerationPreferNoSchedule(pod.Spec.Tolerations)
	return &priorityMetadata{
		nonZeroRequest: getNonZeroRequests(pod),
		podTolerations: tolerationsPreferNoSchedule,
		affinity:       schedulercache.ReconcileAffinity(pod),
	}
}

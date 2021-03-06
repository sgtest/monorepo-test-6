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

package secret

import (
	"github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/api/v1"
)

// fakeManager implements Manager interface for testing purposes.
// simple operations to apiserver.
type fakeManager struct {
}

func NewFakeManager() Manager {
	return &fakeManager{}
}

func (s *fakeManager) GetSecret(namespace, name string) (*v1.Secret, error) {
	return nil, nil
}

func (s *fakeManager) RegisterPod(pod *v1.Pod) {
}

func (s *fakeManager) UnregisterPod(pod *v1.Pod) {
}

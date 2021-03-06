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

package e2e

import (
	"time"

	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/pkg/api"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/pkg/api/v1"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/test/e2e/framework"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-5/test/e2e/perf"

	. "github.com/onsi/ginkgo"
)

var _ = framework.KubeDescribe("Empty [Feature:Empty]", func() {
	f := framework.NewDefaultFramework("empty")

	BeforeEach(func() {
		c := f.ClientSet
		ns := f.Namespace.Name

		// TODO: respect --allow-notready-nodes flag in those functions.
		framework.ExpectNoError(framework.WaitForAllNodesSchedulable(c, framework.TestContext.NodeSchedulableTimeout))
		framework.WaitForAllNodesHealthy(c, time.Minute)

		err := framework.CheckTestingNSDeletedExcept(c, ns)
		framework.ExpectNoError(err)
	})

	It("starts a pod", func() {
		configs, _ := perf.GenerateConfigsForGroup([]*v1.Namespace{f.Namespace}, "empty-pod", 1, 1, framework.GetPauseImageName(f.ClientSet), []string{}, api.Kind("ReplicationController"), 0)
		if len(configs) != 1 {
			framework.Failf("generateConfigs should have generated single config")
		}
		config := configs[0]
		config.SetClient(f.ClientSet)
		framework.ExpectNoError(config.Run())
	})
})

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

package e2e

import (
	"testing"

	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/autoscaling"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/cluster-logging"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/framework"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/perf"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/scheduling"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-12/test/e2e/storage"
)

func init() {
	framework.ViperizeFlags()
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

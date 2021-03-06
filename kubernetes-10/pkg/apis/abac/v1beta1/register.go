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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	api "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/apis/abac"
)

const GroupName = "abac.authorization.kubernetes.io"

// SchemeGroupVersion is the API group and version for abac v1beta1
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1beta1"}

func init() {
	// TODO: delete this, abac should not have its own scheme.
	if err := addKnownTypes(api.Scheme); err != nil {
		// Programmer error.
		panic(err)
	}
	if err := addConversionFuncs(api.Scheme); err != nil {
		// Programmer error.
		panic(err)
	}
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes, addConversionFuncs, RegisterDefaults)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Policy{},
	)
	return nil
}

func (obj *Policy) GetObjectKind() schema.ObjectKind { return &obj.TypeMeta }

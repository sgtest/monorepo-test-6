/*
Copyright 2017 The Kubernetes Authors.

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

package v1beta1_test

import (
	"reflect"
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/api"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/api/install"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/api/v1"
	_ "github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/apis/apps/install"
	. "github.com/sourcegraph/monorepo-test-1/kubernetes-14/pkg/apis/apps/v1beta1"
)

func TestSetDefaultDeployment(t *testing.T) {
	defaultIntOrString := intstr.FromString("25%")
	differentIntOrString := intstr.FromInt(5)
	period := int64(v1.DefaultTerminationGracePeriodSeconds)
	defaultTemplate := v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			DNSPolicy:                     v1.DNSClusterFirst,
			RestartPolicy:                 v1.RestartPolicyAlways,
			SecurityContext:               &v1.PodSecurityContext{},
			TerminationGracePeriodSeconds: &period,
			SchedulerName:                 api.DefaultSchedulerName,
		},
	}
	tests := []struct {
		original *Deployment
		expected *Deployment
	}{
		{
			original: &Deployment{},
			expected: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(1),
					Strategy: DeploymentStrategy{
						Type: RollingUpdateDeploymentStrategyType,
						RollingUpdate: &RollingUpdateDeployment{
							MaxSurge:       &defaultIntOrString,
							MaxUnavailable: &defaultIntOrString,
						},
					},
					RevisionHistoryLimit:    newInt32(2),
					ProgressDeadlineSeconds: newInt32(600),
					Template:                defaultTemplate,
				},
			},
		},
		{
			original: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						RollingUpdate: &RollingUpdateDeployment{
							MaxSurge: &differentIntOrString,
						},
					},
				},
			},
			expected: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						Type: RollingUpdateDeploymentStrategyType,
						RollingUpdate: &RollingUpdateDeployment{
							MaxSurge:       &differentIntOrString,
							MaxUnavailable: &defaultIntOrString,
						},
					},
					RevisionHistoryLimit:    newInt32(2),
					ProgressDeadlineSeconds: newInt32(600),
					Template:                defaultTemplate,
				},
			},
		},
		{
			original: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(3),
					Strategy: DeploymentStrategy{
						Type:          RollingUpdateDeploymentStrategyType,
						RollingUpdate: nil,
					},
				},
			},
			expected: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(3),
					Strategy: DeploymentStrategy{
						Type: RollingUpdateDeploymentStrategyType,
						RollingUpdate: &RollingUpdateDeployment{
							MaxSurge:       &defaultIntOrString,
							MaxUnavailable: &defaultIntOrString,
						},
					},
					RevisionHistoryLimit:    newInt32(2),
					ProgressDeadlineSeconds: newInt32(600),
					Template:                defaultTemplate,
				},
			},
		},
		{
			original: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						Type: RecreateDeploymentStrategyType,
					},
					RevisionHistoryLimit: newInt32(0),
				},
			},
			expected: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						Type: RecreateDeploymentStrategyType,
					},
					RevisionHistoryLimit:    newInt32(0),
					ProgressDeadlineSeconds: newInt32(600),
					Template:                defaultTemplate,
				},
			},
		},
		{
			original: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						Type: RecreateDeploymentStrategyType,
					},
					ProgressDeadlineSeconds: newInt32(30),
					RevisionHistoryLimit:    newInt32(2),
				},
			},
			expected: &Deployment{
				Spec: DeploymentSpec{
					Replicas: newInt32(5),
					Strategy: DeploymentStrategy{
						Type: RecreateDeploymentStrategyType,
					},
					ProgressDeadlineSeconds: newInt32(30),
					RevisionHistoryLimit:    newInt32(2),
					Template:                defaultTemplate,
				},
			},
		},
	}

	for _, test := range tests {
		original := test.original
		expected := test.expected
		obj2 := roundTrip(t, runtime.Object(original))
		got, ok := obj2.(*Deployment)
		if !ok {
			t.Errorf("unexpected object: %v", got)
			t.FailNow()
		}
		if !apiequality.Semantic.DeepEqual(got.Spec, expected.Spec) {
			t.Errorf("object mismatch!\nexpected:\n\t%+v\ngot:\n\t%+v", got.Spec, expected.Spec)
		}
	}
}

func TestDefaultDeploymentAvailability(t *testing.T) {
	d := roundTrip(t, runtime.Object(&Deployment{})).(*Deployment)

	maxUnavailable, err := intstr.GetValueFromIntOrPercent(d.Spec.Strategy.RollingUpdate.MaxUnavailable, int(*(d.Spec.Replicas)), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *(d.Spec.Replicas)-int32(maxUnavailable) <= 0 {
		t.Fatalf("the default value of maxUnavailable can lead to no active replicas during rolling update")
	}
}

func roundTrip(t *testing.T, obj runtime.Object) runtime.Object {
	data, err := runtime.Encode(api.Codecs.LegacyCodec(SchemeGroupVersion), obj)
	if err != nil {
		t.Errorf("%v\n %#v", err, obj)
		return nil
	}
	obj2, err := runtime.Decode(api.Codecs.UniversalDecoder(), data)
	if err != nil {
		t.Errorf("%v\nData: %s\nSource: %#v", err, string(data), obj)
		return nil
	}
	obj3 := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(runtime.Object)
	err = api.Scheme.Convert(obj2, obj3, nil)
	if err != nil {
		t.Errorf("%v\nSource: %#v", err, obj2)
		return nil
	}
	return obj3
}

func newInt32(val int32) *int32 {
	p := new(int32)
	*p = val
	return p
}

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

package tests

import (
	"encoding/json"
	"reflect"
	"testing"

	// TODO: Ideally we should create the necessary package structure in e.g.,
	// pkg/conversion/test/... instead of importing pkg/api here.
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/api"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/api/testapi"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/apis/extensions"
)

func TestV1EncodeDecodeStatus(t *testing.T) {
	status := &metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    200,
		Reason:  metav1.StatusReasonUnknown,
		Message: "",
	}

	v1Codec := testapi.Default.Codec()

	encoded, err := runtime.Encode(v1Codec, status)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	typeMeta := metav1.TypeMeta{}
	if err := json.Unmarshal(encoded, &typeMeta); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if typeMeta.Kind != "Status" {
		t.Errorf("Kind is not set to \"Status\". Got %v", string(encoded))
	}
	if typeMeta.APIVersion != "v1" {
		t.Errorf("APIVersion is not set to \"v1\". Got %v", string(encoded))
	}
	decoded, err := runtime.Decode(v1Codec, encoded)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(status, decoded) {
		t.Errorf("expected: %v, got: %v", status, decoded)
	}
}

func TestExperimentalEncodeDecodeStatus(t *testing.T) {
	status := &metav1.Status{
		Status:  metav1.StatusFailure,
		Code:    200,
		Reason:  metav1.StatusReasonUnknown,
		Message: "",
	}
	// TODO: caesarxuchao: use the testapi.Extensions.Codec() once the PR that
	// moves experimental from v1 to v1beta1 got merged.
	expCodec := api.Codecs.LegacyCodec(extensions.SchemeGroupVersion)
	encoded, err := runtime.Encode(expCodec, status)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	typeMeta := metav1.TypeMeta{}
	if err := json.Unmarshal(encoded, &typeMeta); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if typeMeta.Kind != "Status" {
		t.Errorf("Kind is not set to \"Status\". Got %s", encoded)
	}
	if typeMeta.APIVersion != "v1" {
		t.Errorf("APIVersion is not set to \"\". Got %s", encoded)
	}
	decoded, err := runtime.Decode(expCodec, encoded)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(status, decoded) {
		t.Errorf("expected: %v, got: %v", status, decoded)
	}
}

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

package rest

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/api"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/apis/extensions"
	extensionsapiv1beta1 "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/apis/extensions/v1beta1"
	extensionsclient "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/client/clientset_generated/internalclientset/typed/extensions/internalversion"
	expcontrollerstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/controller/storage"
	daemonstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/daemonset/storage"
	deploymentstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/deployment/storage"
	ingressstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/ingress/storage"
	networkpolicystore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/networkpolicy/storage"
	pspstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/podsecuritypolicy/storage"
	replicasetstore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/replicaset/storage"
	thirdpartyresourcestore "github.com/sourcegraph/monorepo-test-1/kubernetes-10/pkg/registry/extensions/thirdpartyresource/storage"
)

type RESTStorageProvider struct {
	ResourceInterface ResourceInterface
}

func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, bool) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(extensions.GroupName, api.Registry, api.Scheme, api.ParameterCodec, api.Codecs)

	if apiResourceConfigSource.AnyResourcesForVersionEnabled(extensionsapiv1beta1.SchemeGroupVersion) {
		apiGroupInfo.VersionedResourcesStorageMap[extensionsapiv1beta1.SchemeGroupVersion.Version] = p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
		apiGroupInfo.GroupMeta.GroupVersion = extensionsapiv1beta1.SchemeGroupVersion
	}

	return apiGroupInfo, true
}

func (p RESTStorageProvider) v1beta1Storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) map[string]rest.Storage {
	version := extensionsapiv1beta1.SchemeGroupVersion

	storage := map[string]rest.Storage{}

	// This is a dummy replication controller for scale subresource purposes.
	// TODO: figure out how to enable this only if needed as a part of scale subresource GA.
	controllerStorage := expcontrollerstore.NewStorage(restOptionsGetter)
	storage["replicationcontrollers"] = controllerStorage.ReplicationController
	storage["replicationcontrollers/scale"] = controllerStorage.Scale

	if apiResourceConfigSource.ResourceEnabled(version.WithResource("thirdpartyresources")) {
		thirdPartyResourceStorage := thirdpartyresourcestore.NewREST(restOptionsGetter)
		storage["thirdpartyresources"] = thirdPartyResourceStorage
	}

	if apiResourceConfigSource.ResourceEnabled(version.WithResource("daemonsets")) {
		daemonSetStorage, daemonSetStatusStorage := daemonstore.NewREST(restOptionsGetter)
		storage["daemonsets"] = daemonSetStorage
		storage["daemonsets/status"] = daemonSetStatusStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("deployments")) {
		deploymentStorage := deploymentstore.NewStorage(restOptionsGetter)
		storage["deployments"] = deploymentStorage.Deployment
		storage["deployments/status"] = deploymentStorage.Status
		storage["deployments/rollback"] = deploymentStorage.Rollback
		storage["deployments/scale"] = deploymentStorage.Scale
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("ingresses")) {
		ingressStorage, ingressStatusStorage := ingressstore.NewREST(restOptionsGetter)
		storage["ingresses"] = ingressStorage
		storage["ingresses/status"] = ingressStatusStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("podsecuritypolicy")) || apiResourceConfigSource.ResourceEnabled(version.WithResource("podsecuritypolicies")) {
		podSecurityExtensionsStorage := pspstore.NewREST(restOptionsGetter)
		storage["podSecurityPolicies"] = podSecurityExtensionsStorage
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("replicasets")) {
		replicaSetStorage := replicasetstore.NewStorage(restOptionsGetter)
		storage["replicasets"] = replicaSetStorage.ReplicaSet
		storage["replicasets/status"] = replicaSetStorage.Status
		storage["replicasets/scale"] = replicaSetStorage.Scale
	}
	if apiResourceConfigSource.ResourceEnabled(version.WithResource("networkpolicies")) {
		networkExtensionsStorage := networkpolicystore.NewREST(restOptionsGetter)
		storage["networkpolicies"] = networkExtensionsStorage
	}

	return storage
}

func (p RESTStorageProvider) PostStartHook() (string, genericapiserver.PostStartHookFunc, error) {
	return "extensions/third-party-resources", p.postStartHookFunc, nil
}
func (p RESTStorageProvider) postStartHookFunc(hookContext genericapiserver.PostStartHookContext) error {
	clientset, err := extensionsclient.NewForConfig(hookContext.LoopbackClientConfig)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to initialize clusterroles: %v", err))
		return nil
	}

	thirdPartyControl := ThirdPartyController{
		master: p.ResourceInterface,
		client: clientset,
	}
	go wait.Forever(func() {
		if err := thirdPartyControl.SyncResources(); err != nil {
			glog.Warningf("third party resource sync failed: %v", err)
		}
	}, 10*time.Second)

	return nil
}

func (p RESTStorageProvider) GroupName() string {
	return extensions.GroupName
}

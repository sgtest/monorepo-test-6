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

package kubefed

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	federationapi "github.com/sourcegraph/monorepo-test-1/kubernetes-6/federation/apis/federation"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/federation/pkg/kubefed/util"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/api"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/client/clientset_generated/internalclientset"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/kubectl/cmd/templates"
	cmdutil "github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/kubectl/cmd/util"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-6/pkg/kubectl/resource"

	"github.com/spf13/cobra"
)

var (
	unjoin_long = templates.LongDesc(`
		Unjoin removes a cluster from a federation.

        Current context is assumed to be a federation endpoint.
        Please use the --context flag otherwise.`)
	unjoin_example = templates.Examples(`
		# Unjoin removes the specified cluster from a federation.
		# Federation control plane's host cluster context name
		# must be specified via the --host-cluster-context flag
		# to properly cleanup the credentials.
		kubectl unjoin foo --host-cluster-context=bar`)
)

type unjoinFederation struct {
	commonOptions util.SubcommandOptions
}

// NewCmdUnjoin defines the `unjoin` command that removes a cluster
// from a federation.
func NewCmdUnjoin(f cmdutil.Factory, cmdOut, cmdErr io.Writer, config util.AdminConfig) *cobra.Command {
	opts := &unjoinFederation{}

	cmd := &cobra.Command{
		Use:     "unjoin CLUSTER_NAME --host-cluster-context=HOST_CONTEXT",
		Short:   "Unjoins a cluster from a federation",
		Long:    unjoin_long,
		Example: unjoin_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(opts.commonOptions.SetName(cmd, args))
			cmdutil.CheckErr(opts.Run(f, cmdOut, cmdErr, config))
		},
	}

	flags := cmd.Flags()
	opts.commonOptions.Bind(flags)

	return cmd
}

// unjoinFederation is the implementation of the `unjoin` command.
func (u *unjoinFederation) Run(f cmdutil.Factory, cmdOut, cmdErr io.Writer, config util.AdminConfig) error {
	cluster, err := popCluster(f, u.commonOptions.Name)
	if err != nil {
		return err
	}
	if cluster == nil {
		fmt.Fprintf(cmdErr, "WARNING: cluster %q not found in federation, so its credentials' secret couldn't be deleted", u.commonOptions.Name)
		return nil
	}

	// We want a separate client factory to communicate with the
	// federation host cluster. See join_federation.go for details.
	hostFactory := config.ClusterFactory(u.commonOptions.Host, u.commonOptions.Kubeconfig)
	hostClientset, err := hostFactory.ClientSet()
	if err != nil {
		return err
	}

	secretName := cluster.Spec.SecretRef.Name
	secret, err := hostClientset.Core().Secrets(u.commonOptions.FederationSystemNamespace).Get(secretName, metav1.GetOptions{})
	if isNotFound(err) {
		// If this is the case, we cannot get the cluster clientset to delete the
		// config map from that cluster and obviously cannot delete the not existing secret.
		// We just publish the warning as cluster has already been removed from federation.
		fmt.Fprintf(cmdErr, "WARNING: secret %q not found in the host cluster, so it couldn't be deleted", secretName)
	} else if err != nil {
		fmt.Fprintf(cmdErr, "WARNING: Error retrieving secret from the base cluster")
	} else {
		err := deleteSecret(hostClientset, cluster.Spec.SecretRef.Name, u.commonOptions.FederationSystemNamespace)
		if err != nil {
			fmt.Fprintf(cmdErr, "WARNING: secret %q could not be deleted: %v", secretName, err)
			// We anyways continue to try and delete the config map but with above warning
		}

		// We need to ensure deleting the config map created in the deregistered cluster
		// This configmap was created when the cluster joined this federation to aid
		// the kube-dns of that cluster to aid service discovery.
		err = deleteConfigMapFromCluster(hostClientset, secret, cluster, u.commonOptions.FederationSystemNamespace)
		if err != nil {
			fmt.Fprintf(cmdErr, "WARNING: Encountered error in deleting kube-dns configmap, %v", err)
			// We anyways continue to print success message but with above warning
		}
	}

	_, err = fmt.Fprintf(cmdOut, "Successfully removed cluster %q from federation\n", u.commonOptions.Name)
	return err
}

// popCluster fetches the cluster object with the given name, deletes
// it and returns the deleted cluster object.
func popCluster(f cmdutil.Factory, name string) (*federationapi.Cluster, error) {
	mapper, typer := f.Object()
	gvks, _, err := typer.ObjectKinds(&federationapi.Cluster{})
	if err != nil {
		return nil, err
	}
	gvk := gvks[0]
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return nil, err
	}
	client, err := f.ClientForMapping(mapping)
	if err != nil {
		return nil, err
	}

	rh := resource.NewHelper(client, mapping)
	obj, err := rh.Get("", name, false)

	if isNotFound(err) {
		// Cluster isn't registered, there isn't anything to be done here.
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	cluster, ok := obj.(*federationapi.Cluster)
	if !ok {
		return nil, fmt.Errorf("unexpected object type: expected \"federation/v1beta1.Cluster\", got %T: obj: %#v", obj, obj)
	}

	// Remove the cluster resource in the federation API server by
	// calling rh.Delete()
	return cluster, rh.Delete("", name)
}

func deleteConfigMapFromCluster(hostClientset internalclientset.Interface, secret *api.Secret, cluster *federationapi.Cluster, fedSystemNamespace string) error {
	clientset, err := getClientsetFromCluster(secret, cluster)
	if err != nil {
		return err
	}

	cmDep, err := getCMDeployment(hostClientset, fedSystemNamespace)
	if err != nil {
		return err
	}
	domainMap, ok := cmDep.Annotations[util.FedDomainMapKey]
	if !ok {
		return fmt.Errorf("kube-dns config map data missing from controller manager annotations")
	}

	configMap, err := clientset.Core().ConfigMaps(metav1.NamespaceSystem).Get(util.KubeDnsConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, ok := configMap.Data[util.FedDomainMapKey]; !ok {
		return nil
	}
	configMap.Data[util.FedDomainMapKey] = removeConfigMapString(configMap.Data[util.FedDomainMapKey], domainMap)

	_, err = clientset.Core().ConfigMaps(metav1.NamespaceSystem).Update(configMap)
	return err
}

// deleteSecret deletes the secret with the given name from the host
// cluster.
func deleteSecret(clientset internalclientset.Interface, name, namespace string) error {
	orphanDependents := false
	return clientset.Core().Secrets(namespace).Delete(name, &metav1.DeleteOptions{OrphanDependents: &orphanDependents})
}

// isNotFound checks if the given error is a NotFound status error.
func isNotFound(err error) bool {
	statusErr := err
	if urlErr, ok := err.(*url.Error); ok {
		statusErr = urlErr.Err
	}
	return errors.IsNotFound(statusErr)
}

func getClientsetFromCluster(secret *api.Secret, cluster *federationapi.Cluster) (*internalclientset.Clientset, error) {
	serverAddress, err := util.GetServerAddress(cluster)
	if err != nil {
		return nil, err
	}
	if serverAddress == "" {
		return nil, fmt.Errorf("failed to get server address for the cluster: %s", cluster.Name)
	}

	clientset, err := util.GetClientsetFromSecret(secret, serverAddress)
	if err != nil {
		return nil, err
	}
	if clientset == nil {
		// There is a possibility that the clientset is nil without any error reported
		return nil, fmt.Errorf("failed for get client to access cluster: %s", cluster.Name)
	}

	return clientset, nil
}

// removeConfigMapString returns an empty string if last value is removed
// or returns the remaining comma separated strings minus the one to be removed
func removeConfigMapString(str string, toRemove string) string {
	if str == "" {
		return ""
	}

	values := strings.Split(str, ",")
	if len(values) == 1 {
		if values[0] == toRemove {
			return ""
		} else {
			// Somehow our federation string is not here
			// Dont do anything further
			return values[0]
		}
	}

	for i, v := range values {
		if v == toRemove {
			values = append(values[:i], values[i+1:]...)
			break
		}
	}
	return strings.Join(values, ",")
}

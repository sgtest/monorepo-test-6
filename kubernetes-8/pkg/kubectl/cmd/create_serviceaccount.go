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

package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/sourcegraph/monorepo-test-1/kubernetes-8/pkg/kubectl"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-8/pkg/kubectl/cmd/templates"
	cmdutil "github.com/sourcegraph/monorepo-test-1/kubernetes-8/pkg/kubectl/cmd/util"
	"github.com/sourcegraph/monorepo-test-1/kubernetes-8/pkg/util/i18n"
)

var (
	serviceAccountLong = templates.LongDesc(i18n.T(`
		Create a service account with the specified name.`))

	serviceAccountExample = templates.Examples(i18n.T(`
	  # Create a new service account named my-service-account
	  kubectl create serviceaccount my-service-account`))
)

// NewCmdCreateServiceAccount is a macro command to create a new service account
func NewCmdCreateServiceAccount(f cmdutil.Factory, cmdOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serviceaccount NAME [--dry-run]",
		Aliases: []string{"sa"},
		Short:   i18n.T("Create a service account with the specified name"),
		Long:    serviceAccountLong,
		Example: serviceAccountExample,
		Run: func(cmd *cobra.Command, args []string) {
			err := CreateServiceAccount(f, cmdOut, cmd, args)
			cmdutil.CheckErr(err)
		},
	}
	cmdutil.AddApplyAnnotationFlags(cmd)
	cmdutil.AddValidateFlags(cmd)
	cmdutil.AddPrinterFlags(cmd)
	cmdutil.AddInclude3rdPartyFlags(cmd)
	cmdutil.AddGeneratorFlags(cmd, cmdutil.ServiceAccountV1GeneratorName)
	return cmd
}

// CreateServiceAccount implements the behavior to run the create service account command
func CreateServiceAccount(f cmdutil.Factory, cmdOut io.Writer, cmd *cobra.Command, args []string) error {
	name, err := NameFromCommandArgs(cmd, args)
	if err != nil {
		return err
	}
	var generator kubectl.StructuredGenerator
	switch generatorName := cmdutil.GetFlagString(cmd, "generator"); generatorName {
	case cmdutil.ServiceAccountV1GeneratorName:
		generator = &kubectl.ServiceAccountGeneratorV1{Name: name}
	default:
		return cmdutil.UsageError(cmd, fmt.Sprintf("Generator: %s not supported.", generatorName))
	}
	return RunCreateSubcommand(f, cmd, cmdOut, &CreateSubcommandOptions{
		Name:                name,
		StructuredGenerator: generator,
		DryRun:              cmdutil.GetDryRunFlag(cmd),
		OutputFormat:        cmdutil.GetFlagString(cmd, "output"),
	})
}

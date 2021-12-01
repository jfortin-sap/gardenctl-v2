/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/component-base/cli/flag"

	"github.com/gardener/gardenctl-v2/internal/util"
	"github.com/gardener/gardenctl-v2/pkg/cmd/base"

	"github.com/spf13/cobra"
)

// NewCmdConfigSetGarden returns a new (config) set-garden command.
func NewCmdConfigSetGarden(f util.Factory, o *SetGardenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-garden",
		Short: "modify or add Garden to gardenctl configuration.",
		Long:  "Modify or add Garden to gardenctl configuration. E.g. \"gardenctl config set-config my-garden --kubeconfig ~/.kube/kubeconfig.yaml\" to configure or add my-cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, cmd, args); err != nil {
				return fmt.Errorf("failed to complete command options: %w", err)
			}
			if err := o.Validate(); err != nil {
				return err
			}

			return runSetGardenCommand(f, o)
		},
	}

	cmd.Flags().Var(&o.KubeconfigFile, "kubeconfig", "path to kubeconfig file for this Garden cluster. If used without --context, current-context of kubeconfig will be set as context")
	cmd.Flags().Var(&o.ContextName, "context", "use specific context of kubeconfig")
	cmd.Flags().Var(&o.Identity, "identity", "identity is the cluster identity of the Garden cluster")
	cmd.Flags().StringArrayVar(&o.Aliases, "aliases", nil, "aliases")

	return cmd
}

func runSetGardenCommand(f util.Factory, opt *SetGardenOptions) error {
	/*
		kubeconfigFile, err := homedir.Expand(opt.KubeconfigFile)
		if err != nil {
			return fmt.Errorf("failed to resolve ~ in kubeconfig path: %w", err)
		}

		kubeConfig, err := clientcmd.LoadFromFile(kubeconfigFile)
		if err != nil {
			return fmt.Errorf("failed to load kubeconfig file %q: %w", opt.KubeconfigFile, err)
		}

		var contextName string
		if *opt.ContextName != "" {
			contextName = *opt.ContextName
		} else if kubeConfig.CurrentContext != "" {
			contextName = kubeConfig.CurrentContext
		} else {
			return fmt.Errorf("failed to add Garden: No current contextName found for kubeconfig %q", kubeconfigFile)
		}

		if opt.Name == "" {
			opt.Name = contextName
		}



		var clusterConfig *v1.ConfigMap

		if !opt.DisableDownload {
			gardenClient, err := manager.GardenClientForKubeconfig(kubeconfigFile, contextName)
			if err != nil {
				return fmt.Errorf("failed to create client for cluster configuration download: %w", err)
			}

			clusterConfig, err = gardenClient.GetConfigMap(f.Context(), "clusterconfig", "gardenctl-system")
			if err != nil {
				statusError, ok := err.(*apiError.StatusError)
				if !ok || statusError.ErrStatus.Code != 404 {
					return fmt.Errorf("failed to download cluster configuration: %w", err)
				}
			}
		}
	*/

	manager, err := f.Manager()
	if err != nil {
		return err
	}

	return manager.Configuration().SetGarden(opt.Name, opt.KubeconfigFile, opt.ContextName, opt.Identity, opt.Aliases, f.GetConfigFile())
}

// SetGardenOptions is a struct to support view command
type SetGardenOptions struct {
	base.Options

	// Name identifies a garden cluster
	Name string

	// KubeconfigFile is the path to the kubeconfig file of the Garden cluster that shall be added
	KubeconfigFile flag.StringFlag

	// Aliases is a list of alternative names to identify this cluster
	Aliases []string

	// Identity is the cluster identity of the Garden cluster
	Identity flag.StringFlag

	// Context to use for kubeconfig
	ContextName flag.StringFlag
}

// NewSetGardenOptions returns initialized SetGardenOptions
func NewSetGardenOptions(ioStreams util.IOStreams) *SetGardenOptions {
	return &SetGardenOptions{
		Options: base.Options{
			IOStreams: ioStreams,
		},
	}
}

// Complete adapts from the command line args to the data required.
func (o *SetGardenOptions) Complete(_ util.Factory, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.Name = strings.TrimSpace(args[0])
	}

	return nil
}

// Validate validates the provided options
func (o *SetGardenOptions) Validate() error {
	if o.Name == "" {
		return errors.New("garden name is required")
	}

	return nil
}

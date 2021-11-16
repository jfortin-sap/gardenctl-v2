/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	apiError "k8s.io/apimachinery/pkg/api/errors"

	"github.com/mitchellh/go-homedir"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/gardener/gardenctl-v2/internal/util"
	"github.com/gardener/gardenctl-v2/pkg/cmd/base"

	"github.com/spf13/cobra"
)

// NewCmdConfigAddGarden returns a new (config) view command.
func NewCmdConfigAddGarden(f util.Factory, o *AddOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-garden",
		Short: "Add Garden to gardenctl configuration. E.g. \"gardenctl config add kubeconfig.yaml\" to add the Garden cluster that kubeconfig.yaml points to",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(f, cmd, args); err != nil {
				return fmt.Errorf("failed to complete command options: %w", err)
			}
			if err := o.Validate(); err != nil {
				return err
			}

			return runAddGardenCommand(f, o)
		},
	}

	cmd.Flags().StringVar(&o.Name, "name", o.Name, "Set name of new cluster. Must be unique. Default is cluster context name")
	cmd.Flags().StringVar(&o.ContextName, "use-context", o.ContextName, "Use specific context of kubeconfig")
	cmd.Flags().BoolVar(&o.DisableDownload, "disable-download", o.DisableDownload, "If true, the automatic settings download is disabled. Use this e.g. to add a Garden that is not reachable")

	o.AddOutputFlags(cmd)

	return cmd
}

func runAddGardenCommand(f util.Factory, opt *AddOptions) error {
	kubeconfigFile, err := homedir.Expand(opt.KubeconfigFile)
	if err != nil {
		return fmt.Errorf("failed to resolve ~ in kubeconfig path: %w", err)
	}

	kubeConfig, err := clientcmd.LoadFromFile(kubeconfigFile)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig file %q: %w", opt.KubeconfigFile, err)
	}

	var contextName string
	if opt.ContextName != "" {
		contextName = opt.ContextName
	} else if kubeConfig.CurrentContext != "" {
		contextName = kubeConfig.CurrentContext
	} else {
		return fmt.Errorf("failed to add Garden: No current contextName found for kubeconfig %q", kubeconfigFile)
	}

	if opt.Name == "" {
		opt.Name = contextName
	}

	manager, err := f.Manager()
	if err != nil {
		return err
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

	return manager.Configuration().AddGarden(opt.Name, kubeconfigFile, opt.ContextName, clusterConfig, f.GetConfigFile())
}

// AddOptions is a struct to support view command
type AddOptions struct {
	base.Options

	// KubeconfigFile is the path to the kubeconfig file of the Garden cluster that shall be added
	KubeconfigFile string

	// DisableDownload disables the automatic settings download
	DisableDownload bool

	// Name set name for new garden cluster
	Name string

	// Context set name for new garden cluster
	ContextName string
}

// NewAddOptions returns initialized AddOptions
func NewAddOptions(ioStreams util.IOStreams) *AddOptions {
	return &AddOptions{
		Options: base.Options{
			IOStreams: ioStreams,
		},
	}
}

// Complete adapts from the command line args to the data required.
func (o *AddOptions) Complete(_ util.Factory, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		o.KubeconfigFile = strings.TrimSpace(args[0])
	}

	return nil
}

// Validate validates the provided options
func (o *AddOptions) Validate() error {
	if o.KubeconfigFile == "" {
		return errors.New("no kubeconfig path specified")
	}

	return nil
}

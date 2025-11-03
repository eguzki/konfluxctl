package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	konfluxNS string
)

func sourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Find sources used to build provided asset",
		Long:  "Find sources used to build provided asset",
		RunE:  runSources,
	}

	cmd.Flags().StringVarP(&konfluxNS, "namespace", "n", "", "Namespace of the topology ConfigMap")
	if err := cmd.MarkFlagRequired("namespace"); err != nil {
		panic(err)
	}
	return cmd
}

func runSources(cmd *cobra.Command, args []string) error {
	//if konfluxNS == "" {
	//	return errors.New("namespace must be provided")
	//}

	_, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	configuration, err := config.GetConfig()
	if err != nil {
		return err
	}

	_, err = client.New(configuration, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return err
	}

	//topologyKey := client.ObjectKey{Name: "topology", Namespace: topologyNS}
	//topologyConfigMap := &corev1.ConfigMap{}
	//err = k8sClient.Get(ctx, topologyKey, topologyConfigMap)
	//logf.Log.V(1).Info("Reading topology ConfigMap", "object", topologyKey, "error", err)
	//if err != nil {
	//	return err
	//}

	return nil
}

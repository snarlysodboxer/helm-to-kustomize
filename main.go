package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/snarlysodboxer/helm-to-kustomize/internal/processor"
)

func main() {
	var inputFile string
	var outputDir string

	rootCmd := &cobra.Command{
		Use:   "helm-to-kustomize",
		Short: "An opinionated tool that converts helm template output to kustomize files",
		Long: `An opinionated tool that converts 'helm template' output into kustomize-ready YAML files.

Each resource is written to its own file named <kind>.<metadata.name>.yaml.
Common Helm labels and annotations are removed from each resource.
A kustomization.yaml is generated listing all output resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return processor.Run(inputFile, outputDir)
		},
	}

	rootCmd.Flags().StringVar(&inputFile, "input-file", "", "Input YAML file (output of 'helm template')")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for kustomize files")
	if err := rootCmd.MarkFlagRequired("input-file"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if err := rootCmd.MarkFlagRequired("output-dir"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

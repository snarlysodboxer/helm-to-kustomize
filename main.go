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
	var opts processor.Options

	rootCmd := &cobra.Command{
		Use:   "helm-to-kustomize",
		Short: "An opinionated tool that converts helm template output to kustomize files",
		Long: `An opinionated tool that converts 'helm template' output into kustomize-ready YAML files.

Each resource is written to its own file named <kind>.<metadata.name>.yaml.
Common Helm labels and annotations are removed from each resource.
A kustomization.yaml is generated listing all output resources; this can be
skipped with --skip-kustomization, or merged into an existing file with
--merge-kustomization.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return processor.Run(inputFile, outputDir, opts)
		},
	}

	rootCmd.Flags().StringVar(&inputFile, "input-file", "", "Input YAML file (output of 'helm template')")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory for kustomize files")
	rootCmd.Flags().BoolVar(&opts.SkipKustomization, "skip-kustomization", false, "Do not create or update kustomization.yaml")
	rootCmd.Flags().BoolVar(&opts.MergeKustomization, "merge-kustomization", false, "Add resources to an existing kustomization.yaml instead of overwriting it (creates it if missing)")
	rootCmd.MarkFlagsMutuallyExclusive("skip-kustomization", "merge-kustomization")
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

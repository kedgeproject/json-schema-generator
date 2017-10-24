package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/kedgeproject/json-schema-generator/pkg"
	"github.com/spf13/cobra"
)

var (
	verbose           bool
	kedgeSpecLocation string
	kubernetesSchema  string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "schemagen",
	Short: "Generate OpenAPI schema for Kedge.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Add extra logging when verbosity is passed
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := pkg.Conversion(kedgeSpecLocation, kubernetesSchema); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.Flags().StringVarP(&kedgeSpecLocation, "kedgespec", "k", "types.go", "Specify the location of Kedge spec file")
	RootCmd.Flags().StringVarP(&kubernetesSchema, "k8sSchema", "s", "swagger.json", "Specify the location of Kuberenetes Schema file")
}

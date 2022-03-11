package cmd

import (
	"log"

	"beryju.org/kube-dhcp/convert/ms_dhcp"
	"github.com/spf13/cobra"
)

var (
	inputFile string
	outputDir string
)

// convertMSDHCP represents the base command when called without any subcommands
var convertMSDHCP = &cobra.Command{
	Use: "ms-dhcp",
	Run: func(cmd *cobra.Command, args []string) {
		conv, err := ms_dhcp.New(inputFile, outputDir)
		if err != nil {
			log.Fatal(err)
		}
		conv.Run()
	},
}

func init() {
	convert.AddCommand(convertMSDHCP)
	convertMSDHCP.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input XML file")
	convertMSDHCP.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory")
}

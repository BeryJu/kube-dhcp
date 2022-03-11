package cmd

import (
	"github.com/spf13/cobra"
)

var convert = &cobra.Command{
	Use: "convert",
}

func init() {
	rootCmd.AddCommand(convert)
}

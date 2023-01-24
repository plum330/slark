package main

import (
	"fmt"
	"github.com/go-slark/slark/cmd/slark/proto"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "slark",
	Short: "slark plugin",
	Long:  "slark plugin",
}

func init() {
	rootCmd.AddCommand(proto.CreateCmd)
	rootCmd.AddCommand(proto.InstallCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Print(err)
		return
	}
}

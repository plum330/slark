package proto

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "install relative plugin",
	Long:  "install relative plugin",
	Run: func(cmd *cobra.Command, args []string) {
		plugins := []string{
			"google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0", // TODO -> latest
			"google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0", // TODO -> latest
			"github.com/envoyproxy/protoc-gen-validate@latest",
			"github.com/google/gnostic/cmd/protoc-gen-openapi@ade94e0", // TODO -> latest,
			"github.com/go-slark/slark/cmd/protoc-gen-http@latest",
			"github.com/go-slark/slark/cmd/protoc-gen-errors@latest",
			"github.com/google/wire/cmd/wire@latest",
			"github.com/rakyll/statik@latest",
			"github.com/mikefarah/yq/v4@latest",
		}
		for _, plugin := range plugins {
			fmt.Printf("go install %s\n", plugin)
			c := exec.Command("go", "install", plugin)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			err := c.Run()
			if err != nil {
				fmt.Printf("insatll %s fail, err:%+v\n", plugin, err)
				return
			}
		}
	},
}

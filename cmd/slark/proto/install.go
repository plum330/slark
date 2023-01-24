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
			"github.com/favadi/protoc-go-inject-tag@latest",
			"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
			"google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
			"github.com/envoyproxy/protoc-gen-validate@latest",
			"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest",
			"github.com/smallfish-root/common-pkg/xcmd/protoc-gen-gin@latest",
			"github.com/google/wire/cmd/wire@latest",
			"github.com/rakyll/statik@latest",
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

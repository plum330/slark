package proto

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const release = "v1.4.0"

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create the proto code",
	Long:  "create the proto code",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Enter proto files or directory")
			return
		}

		plugins := []string{
			"protoc-gen-go", "protoc-gen-go-grpc",
			"protoc-gen-http", "protoc-gen-openapi",
			"protoc-gen-validate", "protoc-gen-errors",
			"wire", "statik", "yq",
		}
		err := find(plugins...)
		if err != nil {
			cmd := exec.Command("slark", "install")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return
			}
		} else {
			cmd := exec.Command("protoc-gen-http", "--version")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return
			}
			version := strings.TrimSpace(strings.Split(out.String(), " ")[1])
			fmt.Printf("history version:%s\ncurrent version:%s\n", version, release)

			if strings.Compare(version, release) != 0 {
				cmd := exec.Command("slark", "install")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					return
				}
			}
		}

		debug = len(args) >= 2

		err = walk(strings.TrimSpace(args[0]))
		if err != nil {
			fmt.Println(err)
		}
	},
}

func find(name ...string) error {
	var err error
	for _, n := range name {
		_, err = exec.LookPath(n)
		if err != nil {
			break
		}
	}
	return err
}

func walk(dir string) error {
	if len(dir) == 0 {
		return errors.New("dir invalid")
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".proto" || strings.HasPrefix(path, "third_party") {
			return nil
		}
		return create(path, dir)
	})
}

var debug bool

func create(path, dir string) error {
	cmd := []string{
		"-I=.",
		"-I=" + "../third_party",
		"--go_out=" + dir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + dir,
		"--go-grpc_opt=paths=source_relative",
		"--http_out=" + dir,
		"--http_opt=paths=source_relative",
		"--errors_out=" + dir,
		"--errors_opt=paths=source_relative",
		"--openapi_out=" + dir,
		"--openapi_opt=output_mode=source_relative,naming=proto,fq_schema_naming=true,default_response=false",
	}
	protoBytes, err := os.ReadFile(path)
	if err == nil && len(protoBytes) > 0 {
		ok, _ := regexp.Match(`\n[^/]*(import)\s+"validate/validate.proto"`, protoBytes)
		if ok {
			cmd = append(cmd, "--validate_out="+dir, "--validate_opt=paths=source_relative,lang=go")
		}
	}
	cmd = append(cmd, path)
	fd := exec.Command("protoc", cmd...)
	if debug {
		fmt.Println(fd.String())
	}
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	return fd.Run()
}

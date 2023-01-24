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
			"protoc-gen-gin", "protoc-gen-openapiv2",
			"protoc-gen-validate", "protoc-go-inject-tag",
			"wire", "statik",
		}
		err := find(plugins...)
		if err != nil {
			cmd := exec.Command("hack", "install")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return
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

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".proto" || strings.HasPrefix(path, "third_party") {
			return nil
		}
		return create(path, dir)
	})
	if err != nil {
		return err
	}

	return injectTag(dir)
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
		"--gin_out=" + dir,
		"--gin_opt=paths=source_relative",
		"--openapiv2_out=" + dir,
		"--openapiv2_opt=logtostderr=true",
		"--openapiv2_opt=json_names_for_fields=false",
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

func injectTag(dir string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("find %s -name *.pb.go -type f ! -name *_http.pb.go -type f ! -name *_grpc.pb.go", dir))
	var stdOut, stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if debug {
		fmt.Println(cmd.String())
	}
	err := cmd.Run()
	outStr, _ := stdOut.String(), stdErr.String()
	if err != nil {
		fmt.Printf("find file fail, err:%+v\n", err)
		return err
	}

	for _, file := range strings.Split(outStr, "\n") {
		if len(file) == 0 {
			continue
		}

		cmd = exec.Command("protoc-go-inject-tag", "-input", file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if debug {
			fmt.Println(cmd.String())
		}
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

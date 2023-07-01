package base

import (
	"fmt"
	"github.com/fatih/color"
	"os/exec"
	"strings"
)

func Lint(dir string) {
	if !Has("golangci-lint") {
		if !Has("curl") {
			fmt.Println(color.YellowString("WARNING no curl"))
			return
		}
		fd := exec.Command("bash", "-c", "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "+goPath()+"/bin")
		fmt.Println(color.WhiteString("install golangci-lint wait...\n%s", fd.String()))
		err := fd.Run()
		if err != nil {
			fmt.Println(color.YellowString("WARNING install golangci-lint failed: %s\n", err.Error()))
			return
		}
	}
	fd := exec.Command("golangci-lint", "run", "--fix")
	fd.Dir = dir
	err := fd.Run()
	if err != nil {
		fmt.Println(color.YellowString("WARNING lint failed: %s\n", err.Error()))
	}
}

func Has(name string) (ok bool) {
	cmd := exec.Command("command", "-v", name)
	_, err := cmd.Output()
	if err != nil {
		return
	}
	ok = true
	return
}

func goPath() string {
	cmd := exec.Command("go", "env", "GOPATH")
	bs, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.Trim(string(bs), "\n")
}

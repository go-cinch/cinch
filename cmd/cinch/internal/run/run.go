package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

// CmdRun run project command.
var CmdRun = &cobra.Command{
	Use:   "run",
	Short: "Run project",
	Long:  "Run project. Example: cinch run",
	Run:   run,
}
var workDir string

func init() {
	CmdRun.Flags().StringVarP(&workDir, "dir", "d", "", "working directory")
}

func run(cmd *cobra.Command, args []string) {
	var dir string
	cmdArgs, programArgs := splitArgs(cmd, args)
	if len(cmdArgs) > 0 {
		dir = cmdArgs[0]
	}
	base, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", err)
		return
	}
	if dir == "" {
		// find the directory containing the cmd/*
		cmdPath, e := findCMD(base)
		if e != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", err)
			return
		}
		switch len(cmdPath) {
		case 0:
			fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", "The cmd directory cannot be found in the current directory")
			return
		case 1:
			for _, v := range cmdPath {
				dir = v
			}
		default:
			var cmdPaths []string
			for k := range cmdPath {
				cmdPaths = append(cmdPaths, k)
			}
			prompt := &survey.Select{
				Message:  "Which directory do you want to run?",
				Options:  cmdPaths,
				PageSize: 10,
			}
			e = survey.AskOne(prompt, &dir)
			if e != nil || dir == "" {
				return
			}
			dir = cmdPath[dir]
		}
	}
	fd := exec.Command("go", append([]string{"run", dir}, programArgs...)...)
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	fd.Dir = dir
	changeWorkingDirectory(fd, workDir)
	if err = fd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mERROR: %s\033[m\n", err.Error())
		return
	}
}

func splitArgs(cmd *cobra.Command, args []string) (cmdArgs, programArgs []string) {
	dashAt := cmd.ArgsLenAtDash()
	if dashAt >= 0 {
		return args[:dashAt], args[dashAt:]
	}
	return args, []string{}
}

func findCMD(base string) (m map[string]string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	if !strings.HasSuffix(wd, "/") {
		wd += "/"
	}
	var root bool
	next := func(dir string) (map[string]string, error) {
		cmdPath := make(map[string]string)
		err = filepath.Walk(dir, func(walkPath string, info os.FileInfo, err error) (e error) {
			// multi level directory is not allowed under the cmdPath directory, so it is judged that the path ends with cmdPath.
			if strings.HasSuffix(walkPath, "cmd") {
				var paths []os.DirEntry
				paths, e = os.ReadDir(walkPath)
				if e != nil {
					return
				}
				for _, fileInfo := range paths {
					if fileInfo.IsDir() {
						abs := filepath.Join(walkPath, fileInfo.Name())
						cmdPath[strings.TrimPrefix(abs, wd)] = abs
					}
				}
				return
			}
			if info.Name() == "go.mod" {
				root = true
			}
			return
		})
		return cmdPath, err
	}
	for i := 0; i < 5; i++ {
		tmp := base
		m, err = next(tmp)
		if err != nil {
			return
		}
		if len(m) > 0 {
			return
		}
		if root {
			break
		}
		_ = filepath.Join(base, "..")
	}
	m = map[string]string{"": base}
	return
}

func changeWorkingDirectory(cmd *exec.Cmd, targetDir string) {
	targetDir = strings.TrimSpace(targetDir)
	if targetDir != "" {
		cmd.Dir = targetDir
	}
}

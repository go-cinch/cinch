package project

import (
	"context"
	"fmt"
	"github.com/go-cinch/cinch/cmd/cinch/internal/base"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// Project is a project template.
type Project struct {
	Name string
	Path string
}

// New new a project from remote repo.
func (p *Project) New(ctx context.Context, dir string, layout string, branch string) (err error) {
	to := filepath.Join(dir, p.Name)
	if _, err = os.Stat(to); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Name)
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		var override bool
		err = survey.AskOne(prompt, &override)
		if err != nil {
			return
		}
		if !override {
			return
		}
		os.RemoveAll(to)
	}
	fmt.Printf("🚀 Creating service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)
	repo := base.NewRepo(layout, branch)
	if err = repo.CopyTo(ctx, to, p.Path, []string{".git", ".github"}); err != nil {
		return
	}
	err = p.customChange(to)
	if err != nil {
		return
	}
	base.Tree(to, dir)

	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(p.Name))
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")

	fmt.Println(color.WhiteString("$ cd %s", p.Name))
	fmt.Println(color.WhiteString("$ make all"))
	fmt.Println(color.WhiteString("$ cd cmd/%s", p.Name))
	fmt.Println(color.WhiteString("$ go run main.go wire_gen.go\n"))

	fmt.Println("			🤝 Thanks for using Cinch")
	fmt.Println("	📚 Tutorial: https://go-cinch.github.io/docs/#/started/0.init")
	return
}

func (p *Project) customChange(to string) (err error) {
	err = replaceContent(filepath.Join(to, "cmd", "server", "main.go"), "layout", p.Name)
	if err != nil {
		return
	}

	err = os.Rename(filepath.Join(to, "cmd", "server"), filepath.Join(to, "cmd", p.Name))
	if err != nil {
		return
	}

	if p.Name == "game" {
		return
	}
	contents := []string{
		filepath.Join(to, "api", "game", "game.pb.go"),
		filepath.Join(to, "api", "game", "game.pb.validate.go"),
		filepath.Join(to, "api", "game", "game_grpc.pb.go"),
		filepath.Join(to, "api", "game", "game_http.pb.go"),
		filepath.Join(to, "cmd", p.Name, "wire_gen.go"),
		filepath.Join(to, "internal", "biz", "biz.go"),
		filepath.Join(to, "internal", "biz", "game.go"),
		filepath.Join(to, "internal", "conf", "db", "2022081510-game.sql"),
		filepath.Join(to, "internal", "data", "data.go"),
		filepath.Join(to, "internal", "data", "game.go"),
		filepath.Join(to, "internal", "pkg", "task", "task.go"),
		filepath.Join(to, "internal", "server", "middleware", "whitelist.go"),
		filepath.Join(to, "internal", "server", "grpc.go"),
		filepath.Join(to, "internal", "server", "http.go"),
		filepath.Join(to, "internal", "service", "game.go"),
		filepath.Join(to, "internal", "service", "service.go"),
	}

	for _, item := range contents {
		err = replaceContent(item, "game", p.Name)
		if err != nil {
			return
		}
		err = replaceContent(item, "Game", camelCase(p.Name))
		if err != nil {
			return
		}
	}

	renames := [][2]string{
		{
			filepath.Join(to, "api", "game", "game.pb.go"),
			filepath.Join(to, "api", "game", p.Name+".pb.go"),
		},
		{
			filepath.Join(to, "api", "game", "game.pb.validate.go"),
			filepath.Join(to, "api", "game", p.Name+".pb.validate.go"),
		},
		{
			filepath.Join(to, "api", "game", "game_grpc.pb.go"),
			filepath.Join(to, "api", "game", p.Name+"_grpc.pb.go"),
		},
		{
			filepath.Join(to, "api", "game", "game_http.pb.go"),
			filepath.Join(to, "api", "game", p.Name+"_http.pb.go"),
		},
		{
			filepath.Join(to, "api", "game"),
			filepath.Join(to, "api", p.Name),
		},
		{
			filepath.Join(to, "internal", "biz", "game.go"),
			filepath.Join(to, "internal", "biz", p.Name+".go"),
		},
		{
			filepath.Join(to, "internal", "conf", "db", "2022081510-game.sql"),
			filepath.Join(to, "internal", "conf", "db", "2022081510-"+p.Name+".sql"),
		},
		{
			filepath.Join(to, "internal", "data", "game.go"),
			filepath.Join(to, "internal", "data", p.Name+".go"),
		},
		{
			filepath.Join(to, "internal", "service", "game.go"),
			filepath.Join(to, "internal", "service", p.Name+".go"),
		},
	}

	for _, item := range renames {
		err = os.Rename(item[0], item[1])
		if err != nil {
			return
		}
	}

	batchGofmt(to)

	return
}

func replaceContent(src, o, n string) (err error) {
	read, err := os.ReadFile(src)
	if err != nil {
		return
	}

	newContents := strings.Replace(string(read), o, n, -1)

	err = os.WriteFile(src, []byte(newContents), 0)
	if err != nil {
		return
	}
	return
}

var (
	camelRe = regexp.MustCompile("(_)([a-zA-Z]+)")
	caser   = cases.Title(language.Und)
)

func camelCase(str string) string {
	camel := camelRe.ReplaceAllString(str, " $2")
	camel = caser.String(str)
	camel = strings.Replace(camel, " ", "", -1)
	return camel
}

func batchGofmt(path string) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) (e error) {
		if err == nil && info != nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			exec.Command("gofmt", "-l", "-w", "-s", path).Run()
		}
		return
	})
}

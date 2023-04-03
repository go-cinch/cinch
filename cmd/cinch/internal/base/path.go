package base

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func cinchHome() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	home := filepath.Join(dir, ".cinch")
	if _, err = os.Stat(home); os.IsNotExist(err) {
		if err = os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func cinchHomeWithDir(dir string) string {
	home := filepath.Join(cinchHome(), dir)
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err = os.MkdirAll(home, 0o700); err != nil {
			log.Fatal(err)
		}
	}
	return home
}

func copyFile(src, dst string, replaces []string) (err error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}
	buf, err := os.ReadFile(src)
	if err != nil {
		return
	}
	var old string
	for i, next := range replaces {
		if i%2 == 0 {
			old = next
			continue
		}
		buf = bytes.ReplaceAll(buf, []byte(old), []byte(next))
	}
	return os.WriteFile(dst, buf, srcInfo.Mode())
}

func copyDir(src, dst string, replaces, ignores []string) (err error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return
	}

	fds, err := os.ReadDir(src)
	if err != nil {
		return
	}
	for _, fd := range fds {
		if hasSets(fd.Name(), ignores) {
			continue
		}
		srcFilePath := filepath.Join(src, fd.Name())
		dstFilePath := filepath.Join(dst, fd.Name())
		if fd.IsDir() {
			err = copyDir(srcFilePath, dstFilePath, replaces, ignores)
		} else {
			err = copyFile(srcFilePath, dstFilePath, replaces)
		}
		if err != nil {
			return
		}
	}
	return
}

func hasSets(name string, sets []string) bool {
	for _, ig := range sets {
		if ig == name {
			return true
		}
	}
	return false
}

func Tree(path string, dir string) {
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) (e error) {
		if err == nil && info != nil && !info.IsDir() {
			fmt.Printf("%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.Replace(path, dir+"/", "", -1), info.Size())
		}
		return
	})
}

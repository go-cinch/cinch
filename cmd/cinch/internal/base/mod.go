package base

import (
	"golang.org/x/mod/modfile"
	"os"
)

// ModulePath returns go module path.
func ModulePath(filename string) (dir string, err error) {
	modBytes, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	dir = modfile.ModulePath(modBytes)
	return
}

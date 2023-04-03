package main

import (
	"github.com/go-cinch/cinch/cmd/cinch/internal/project"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "cinch",
	Short:   "Cinch: An elegant toolkit for Go Cinch.",
	Long:    "Cinch: An elegant toolkit for Go Cinch. More: https://go-cinch.github.io/docs",
	Version: release,
}

func init() {
	rootCmd.AddCommand(project.CmdNew)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

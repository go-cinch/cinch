package main

import (
	"github.com/go-cinch/cinch/cmd/cinch/internal"
	"github.com/go-cinch/cinch/cmd/cinch/internal/gen"
	"github.com/go-cinch/cinch/cmd/cinch/internal/project"
	"github.com/go-cinch/cinch/cmd/cinch/internal/run"
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{
	Use:     "cinch",
	Short:   "Cinch: An elegant toolkit for Go Cinch.",
	Long:    "Cinch: An elegant toolkit for Go Cinch. More: https://go-cinch.github.io/docs",
	Version: internal.Release,
}

func init() {
	rootCmd.AddCommand(project.CmdNew)
	rootCmd.AddCommand(run.CmdRun)
	rootCmd.AddCommand(gen.CmdGen)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

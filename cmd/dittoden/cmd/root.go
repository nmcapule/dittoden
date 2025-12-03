package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dataDir string
)

var rootCmd = &cobra.Command{
	Use:   "dittoden",
	Short: "Dittoden CLI for domain modeling",
	Long:  `Dittoden is a prototyping system for crowd-sourced domain modeling.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dataDir, "dir", "examples", "Path to the folder containing .txtpb files")
}

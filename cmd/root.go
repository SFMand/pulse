package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pulse",
	Short: "A lightweight, real-time service health monitor for the command line.",
	Long: `Pulse is a minimalist infrastructure monitoring tool built in Go.
It allows you to track the availability and latency of endpoints and services directly from your terminal.`,

	Run: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}

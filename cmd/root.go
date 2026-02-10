package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "pulse",
	Short: "A lightweight, real-time service health monitor for the command line.",
	Long: `Pulse is a minimalist monitoring tool built in Go.
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
	cobra.OnInitialize(initConfig)
	viper.SetDefault("interval", "30s")
	viper.SetDefault("targets", []map[string]string{})
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/pulse")
		viper.SetConfigName("pulse")
	}

	viper.SetEnvPrefix("PULSE")
	viper.BindEnv("config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

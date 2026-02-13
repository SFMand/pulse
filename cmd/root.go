package cmd

import (
	"log/slog"
	"os"

	"github.com/SFMand/pulse/logic/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)
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
		slog.Error("command failed", "errmsg", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initConfig)
	viper.SetDefault("interval", config.DefaultInterval)
	viper.SetDefault("targets", []map[string]string{})
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
}

func initLogger() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	Logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(Logger)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Unable to find home directory", "error", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/pulse")
		viper.SetConfigName("pulse")
	}

	viper.SetEnvPrefix("PULSE")
	viper.BindEnv("config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		slog.Debug("using config file", "file", viper.ConfigFileUsed())
	} else {
		slog.Warn("no config file found, using defaults and environment variables", "errmsg", err)
	}
}

package cmd

import (
	"time"

	"github.com/SFMand/pulse/logic"

	"github.com/SFMand/pulse/logic/config"
	"github.com/spf13/cobra"
)

var intervalOverridden time.Duration

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor a service endpoint",
	Long:  `Monitor continuously pings service endpoints listed in pulse.yaml and reports their health status.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagOverridden := cmd.Flags().Changed("interval")
		return logic.StartCheck(flagOverridden, intervalOverridden)
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().DurationVarP(&intervalOverridden, "interval", "i", config.DefaultInterval, "Polling interval (overrides config)")
}

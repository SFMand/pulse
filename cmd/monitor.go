package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

var monitorInterval time.Duration

var monitorCmd = &cobra.Command{
	Use:   "monitor [URL]",
	Short: "Monitor a service endpoint",
	Long: `Monitor continuously pings a service endpoint and reports its health status.
	
	Example:
	  pulse monitor https://example.com
	  pulse monitor -i 10s https://example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceURL := args[0]
		return startMonitoring(serviceURL, monitorInterval)
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().DurationVarP(&monitorInterval, "interval", "i", 30*time.Second, "Polling interval")
}

func startMonitoring(serviceURL string, pollInterval time.Duration) error {
	// validate URL
	parsedURL, err := url.Parse(serviceURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if pollInterval <= 0 {
		return fmt.Errorf("interval must be positive, got %v", pollInterval)
	}

	if verbose {
		fmt.Printf("Starting monitor: %s with interval %v\n", parsedURL.String(), pollInterval)
	}

	fmt.Printf("Monitor placeholder: would ping %s every %v\n", serviceURL, pollInterval)
	return nil
}

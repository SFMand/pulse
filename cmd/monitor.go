package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SFMand/pulse/logic/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var monitorInterval time.Duration

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor a service endpoint",
	Long:  `Monitor continuously pings service endpoints listed in pulse.yaml and reports their health status.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagOverridden := cmd.Flags().Changed("interval")
		return startMonitoring(flagOverridden)
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)
	monitorCmd.Flags().DurationVarP(&monitorInterval, "interval", "i", 30*time.Second, "Polling interval (overrides config)")
}

func startMonitoring(flagOverridden bool) error {
	cfgInterval := viper.GetDuration("interval")
	timeout := viper.GetDuration("timeout")
	if timeout == 0 {
		timeout = config.DefaultTimeout
	}

	var rawTargets []config.Target
	if err := viper.UnmarshalKey("targets", &rawTargets); err != nil {
		return fmt.Errorf("failed to read targets from config: %w", err)
	}
	if len(rawTargets) == 0 {
		return fmt.Errorf("no targets found in configuration")
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if verbose {
			fmt.Println("Received interrupt, shutting down monitor...")
		}
		cancel()
	}()

	var wg sync.WaitGroup

	for _, t := range rawTargets {
		name := t.Name
		if name == "" {
			name = t.URL
		}
		parsed, err := url.Parse(t.URL)
		if err != nil {
			fmt.Printf("Skipping target with invalid URL %q: %v\n", t.URL, err)
			continue
		}

		var tgtInterval time.Duration
		if flagOverridden {
			tgtInterval = monitorInterval
		} else if t.Interval != "" {
			d, err := time.ParseDuration(t.Interval)
			if err != nil {
				fmt.Printf("Invalid interval for target %s (%s), using default\n", name, t.Interval)
				tgtInterval = config.DefaultInterval
			} else {
				tgtInterval = d
			}
		} else if cfgInterval != 0 {
			tgtInterval = cfgInterval
		} else {
			tgtInterval = config.DefaultInterval
		}

		if verbose {
			fmt.Printf("Verbose: Starting monitor for %s with interval %v\n", name, tgtInterval)
		}

		wg.Add(1)
		go func(ctx context.Context, name string, u *url.URL, interval time.Duration) {
			defer wg.Done()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			checkOnce := func() {
				start := time.Now()
				resp, err := client.Get(u.String())
				latency := time.Since(start)
				if err != nil {
					fmt.Printf("%s: DOWN (%v)\n", name, err)
					return
				}
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					fmt.Printf("%s: OK (%s) %d %v\n", name, u.Host, resp.StatusCode, latency)
				} else {
					fmt.Printf("%s: UNHEALTHY (%d) %v\n", name, resp.StatusCode, latency)
				}
			}

			// Run one immediate check, then check on ticker interval
			checkOnce()

			for {
				select {
				case <-ctx.Done():
					if verbose {
						fmt.Printf("Stopping monitor for %s\n", name)
					}
					return
				case <-ticker.C:
					checkOnce()
				}
			}
		}(ctx, name, parsed, tgtInterval)
	}

	wg.Wait()
	return nil
}

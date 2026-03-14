package logic

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SFMand/pulse/logic/config"
)

func StartCheck(flagOverridden bool, intervalOverridden time.Duration) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: cfg.Timeout,
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
		slog.Debug("received interrupt")
		cancel()
	}()

	var wg sync.WaitGroup

	for _, trg := range cfg.Targets {
		if flagOverridden {
			trg.Interval = intervalOverridden
		}
		if trg.Interval <= 0 {
			slog.Warn("invalid target interval, falling back to default", "name", trg.Name, "url", trg.URL, "interval", trg.Interval.String(), "default", config.DefaultInterval.String())
			trg.Interval = config.DefaultInterval
		}
		slog.Debug("start monitoring target", "name", trg.Name, "url", trg.URL, "interval", trg.Interval.String())

		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			ticker := time.NewTicker(trg.Interval)
			defer ticker.Stop()

			checkOnce := func() {
				start := time.Now()
				resp, err := client.Get(trg.URL)
				latency := time.Since(start)
				if err != nil {
					slog.Warn("DOWN", slog.Group("target", "name", trg.Name, "url", trg.URL), "error", err)
					return
				}
				resp.Body.Close()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
					slog.Info("UP", slog.Group("target", "name", trg.Name, "url", trg.URL), "status", resp.StatusCode, "latency", latency.String())
				} else {
					slog.Warn("UNHEALTHY", slog.Group("target", "name", trg.Name, "url", trg.URL), "status", resp.StatusCode, "latency", latency.String())
				}
			}

			// Run one immediate check, then check on ticker interval
			checkOnce()

			for {
				select {
				case <-ctx.Done():
					slog.Info("stop monitoring target", "name", trg.Name, "url", trg.URL)
					return
				case <-ticker.C:
					checkOnce()
				}
			}
		}(ctx)
	}

	wg.Wait()
	return nil
}

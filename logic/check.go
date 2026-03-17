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

	tea "charm.land/bubbletea/v2"
	"github.com/SFMand/pulse/logic/config"
)

func pushStatusUpdate(updates chan<- config.StatusUpdate, update config.StatusUpdate) {
	select {
	case updates <- update:
	default:
	}
}

func StartCheck(flagOverridden bool, intervalOverridden time.Duration) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	updates := make(chan config.StatusUpdate, len(cfg.Targets)*4)
	done := make(chan struct{})

	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer close(updates)
	go func() {
		<-ctx.Done()
		close(done)
	}()

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
				now := time.Now()
				if err != nil {
					slog.Warn("DOWN", slog.Group("target", "name", trg.Name, "url", trg.URL), "error", err)
					pushStatusUpdate(updates, config.StatusUpdate{
						Name:      trg.Name,
						State:     "DOWN",
						Detail:    err.Error(),
						Latency:   0,
						LastCheck: now,
					})
					return
				}
				resp.Body.Close()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
					slog.Info("UP", slog.Group("target", "name", trg.Name, "url", trg.URL), "status", resp.StatusCode, "latency", latency.String())
					pushStatusUpdate(updates, config.StatusUpdate{
						Name:      trg.Name,
						State:     "UP",
						Detail:    http.StatusText(resp.StatusCode),
						Latency:   latency,
						LastCheck: now,
					})
				} else {
					slog.Warn("UNHEALTHY", slog.Group("target", "name", trg.Name, "url", trg.URL), "status", resp.StatusCode, "latency", latency.String())
					pushStatusUpdate(updates, config.StatusUpdate{
						Name:      trg.Name,
						State:     "UNHEALTHY",
						Detail:    http.StatusText(resp.StatusCode),
						Latency:   latency,
						LastCheck: now,
					})
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

	model := config.InitializeModel(cfg.Targets, updates, done)
	program := tea.NewProgram(model)

	_, err = program.Run()
	cancel()

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

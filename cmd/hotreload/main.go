package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/VaibhavPrakash0503/hotreload/internal/builder"
	"github.com/VaibhavPrakash0503/hotreload/internal/debouncer"
	"github.com/VaibhavPrakash0503/hotreload/internal/runner"
	"github.com/VaibhavPrakash0503/hotreload/internal/watcher"
	"github.com/spf13/cobra"
)

func main() {
	var (
		root       string
		buildCmd   string
		execCmd    string
		debounceMs int
		verbose    bool
	)

	rootCmd := &cobra.Command{
		Use:   "hotreload",
		Short: "A file watcher that rebuilds and restarts your project on changes",
		Example: `  hotreload --root ./myapp --build "go build -o ./bin/app ." --exec "./bin/app"
  hotreload --root . --exec "python main.py"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Configure logging level
			logLevel := slog.LevelInfo
			if verbose {
				logLevel = slog.LevelDebug
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: logLevel,
			})))

			// Root context — cancelled only on SIGINT/SIGTERM
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set up file watcher
			w, err := watcher.NewWatcher(root)
			if err != nil {
				return err
			}
			defer w.Close()

			b := builder.NewBuilder(buildCmd)
			r := runner.NewRunner(execCmd)

			// cancelBuild holds the cancel func for the currently running build.
			// Whenever a new reload fires, we call this to discard the old build.
			var (
				buildMu     sync.Mutex
				cancelBuild context.CancelFunc
			)

			// reload is the core action triggered on every file change.
			// Only the latest invocation wins — any in-progress build is cancelled first.
			reload := func() {
				// Cancel any in-progress build and create a fresh context for this one.
				buildMu.Lock()
				if cancelBuild != nil {
					cancelBuild() // kills the previous go build process immediately
				}
				buildCtx, newCancel := context.WithCancel(ctx)
				cancelBuild = newCancel
				buildMu.Unlock()

				slog.Info("Change detected — reloading...")

				if err := r.Stop(); err != nil {
					slog.Error("Failed to stop process", "error", err)
				}

				if buildCmd != "" {
					if err := b.Build(buildCtx); err != nil {
						// Check if the build was cancelled by a newer reload.
						if buildCtx.Err() != nil {
							slog.Info("Build cancelled — newer change detected, skipping")
							return
						}
						slog.Error("Build failed — not restarting", "error", err)
						return
					}
				}

				// Start the new server.
				if err := r.Start(ctx); err != nil {
					slog.Error("Failed to start process", "error", err)
				}
			}

			deb := debouncer.NewDebouncer(debounceMs, reload)
			defer deb.Stop()

			// Initial build + start on launch (no waiting for a file change).
			slog.Info("Starting hotreload", "root", root, "build", buildCmd, "exec", execCmd)
			if buildCmd != "" {
				initCtx, initCancel := context.WithCancel(ctx)
				buildMu.Lock()
				cancelBuild = initCancel
				buildMu.Unlock()

				if err := b.Build(initCtx); err != nil {
					slog.Error("Initial build failed", "error", err)
					return err
				}
			}
			if err := r.Start(ctx); err != nil {
				return err
			}

			// Watch for file changes in the background.
			events := make(chan string, 64)
			go w.Watch(events)
			go func() {
				for range events {
					deb.Trigger()
				}
			}()

			// Block until SIGINT or SIGTERM.
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit

			slog.Info("Shutting down...")
			deb.Stop()
			if err := r.Stop(); err != nil {
				slog.Error("Error stopping process", "error", err)
			}

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&root, "root", "r", ".", "Root directory to watch for changes")
	rootCmd.Flags().StringVarP(&buildCmd, "build", "b", "", "Build command to run before (re)starting (optional)")
	rootCmd.Flags().StringVarP(&execCmd, "exec", "e", "", "Command to run (required)")
	rootCmd.Flags().IntVar(&debounceMs, "debounce", 500, "Debounce delay in milliseconds")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug/verbose logging")
	rootCmd.MarkFlagRequired("exec")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

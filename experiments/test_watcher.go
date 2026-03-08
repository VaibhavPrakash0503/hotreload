//go:build ignore

package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/VaibhavPrakash0503/hotreload/internal/watcher"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Create watcher
	w, err := watcher.NewWatcher("./testserver")
	if err != nil {
		fmt.Println("Error creating watcher:", err)
		return
	}
	defer w.Close()

	// Create events channel
	events := make(chan string)

	// Start watching
	go w.Watch(events)

	logger.Info("Watching ./testserver for changes...")
	logger.Info("Edit a .go file to see it detect changes!")
	logger.Info("Press Ctrl+C to stop")

	// Print events
	for path := range events {
		logger.Info("🔥 FILE CHANGED:", path)
	}
}

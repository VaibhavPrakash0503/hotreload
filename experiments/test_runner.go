//go:build ignore

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/VaibhavPrakash0503/hotreload/internal/runner"
)

func main() {
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create runner
	r := runner.NewRunner("go run ./testserver")

	fmt.Println("🚀 Starting server...")
	ctx := context.Background()

	err := r.Start(ctx)
	if err != nil {
		fmt.Println("❌ Failed to start:", err)
		return
	}

	fmt.Println("✅ Server started!")
	fmt.Println("📝 Logs should appear below:")
	fmt.Println("🌐 Visit http://localhost:8080")
	fmt.Println()

	// Let it run for 10 seconds
	time.Sleep(10 * time.Second)

	fmt.Println("\n🛑 Stopping server...")
	err = r.Stop()
	if err != nil {
		fmt.Println("❌ Failed to stop:", err)
		return
	}

	fmt.Println("✅ Server stopped!")

	// Verify process is dead
	time.Sleep(1 * time.Second)
	fmt.Println("✅ Test complete!")
}

//go:build ignore

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/VaibhavPrakash0503/hotreload/internal/debouncer"
)

func main() {
	callCount := 0

	// Create debouncer with 300ms delay
	deb := debouncer.NewDebouncer(300, func() {
		callCount++
		fmt.Printf("✅ Callback executed! (call #%d)\n", callCount)
	})

	fmt.Println("Test 1: Rapid triggers (should only call once)")
	fmt.Println("Triggering 10 times rapidly...")

	for i := 0; i < 10; i++ {
		deb.Trigger()
		fmt.Printf("  Trigger %d\n", i+1)
		time.Sleep(50 * time.Millisecond) // 50ms between triggers
	}

	// Wait for callback
	time.Sleep(500 * time.Millisecond)

	fmt.Printf("\nResult: Callback called %d time(s)\n", callCount)
	fmt.Println("Expected: 1 time")

	fmt.Println("\n" + strings.Repeat("-", 50))

	fmt.Println("\nTest 2: Trigger, wait, trigger again")
	deb.Trigger()
	fmt.Println("  Triggered")

	time.Sleep(400 * time.Millisecond) // Wait longer than 300ms
	fmt.Println("  (waited 400ms)")

	deb.Trigger()
	fmt.Println("  Triggered again")

	time.Sleep(400 * time.Millisecond)

	fmt.Printf("\nResult: Callback called %d time(s) total\n", callCount)
	fmt.Println("Expected: 3 times total (1 from test 1, 2 from test 2)")
}

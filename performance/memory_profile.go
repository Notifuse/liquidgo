// +build ignore

package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/Notifuse/liquidgo/performance"
)

type memStats struct {
	alloc      uint64
	totalAlloc uint64
	mallocs    uint64
}

func getMemStats() memStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return memStats{
		alloc:      m.Alloc,
		totalAlloc: m.TotalAlloc,
		mallocs:    m.Mallocs,
	}
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	fmt.Println("=== Liquid Go Memory Profiler ===\n")

	// Initialize theme runner
	fmt.Println("Initializing ThemeRunner...")
	runner, err := performance.NewThemeRunner()
	if err != nil {
		log.Fatal("Failed to initialize ThemeRunner:", err)
	}

	// Force GC before starting
	runtime.GC()

	// Profile parse phase
	fmt.Println("\nProfiling: Parse phase...")
	beforeParse := getMemStats()
	
	if err := runner.Compile(); err != nil {
		log.Fatal("Parse phase failed:", err)
	}
	
	runtime.GC() // Force GC to get accurate retained memory
	afterParse := getMemStats()

	parseAllocated := afterParse.totalAlloc - beforeParse.totalAlloc
	parseMallocs := afterParse.mallocs - beforeParse.mallocs

	fmt.Println("Done.")

	// Profile render phase
	fmt.Println("\nProfiling: Render phase...")
	beforeRender := getMemStats()
	
	if err := runner.Render(); err != nil {
		log.Fatal("Render phase failed:", err)
	}
	
	runtime.GC() // Force GC to get accurate retained memory
	afterRender := getMemStats()

	renderAllocated := afterRender.totalAlloc - beforeRender.totalAlloc
	renderMallocs := afterRender.mallocs - beforeRender.mallocs

	fmt.Println("Done.")

	// Display results table
	fmt.Println("\n╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║           Memory Profiling Results                    ║")
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-20s ║ %-15s ║ %-12s ║\n", "Phase", "Parse", "Render")
	fmt.Println("╠═══════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-20s ║ %-15s ║ %-12s ║\n", 
		"Total allocated", 
		formatBytes(parseAllocated), 
		formatBytes(renderAllocated))
	fmt.Printf("║ %-20s ║ %-15d ║ %-12d ║\n", 
		"Total allocations", 
		parseMallocs, 
		renderMallocs)
	fmt.Println("╚═══════════════════════════════════════════════════════╝")

	fmt.Println("\nNote: These measurements include allocations made during")
	fmt.Println("      the phase execution. Use runtime/pprof for detailed")
	fmt.Println("      allocation traces.")
}


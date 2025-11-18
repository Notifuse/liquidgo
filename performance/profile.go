//go:build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/Notifuse/liquidgo/performance"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write memory profile to file")
	iterations = flag.Int("iterations", 200, "number of iterations to run")
)

func main() {
	flag.Parse()

	// Initialize theme runner
	fmt.Println("Initializing ThemeRunner...")
	runner, err := performance.NewThemeRunner()
	if err != nil {
		log.Fatal("Failed to initialize ThemeRunner:", err)
	}

	// Warm up
	fmt.Println("Warming up...")
	if err := runner.Run(); err != nil {
		log.Fatal("Warmup failed:", err)
	}

	// CPU profiling
	if *cpuprofile != "" {
		fmt.Printf("Starting CPU profiling (writing to %s)...\n", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("Could not create CPU profile:", err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("Could not start CPU profile:", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Run benchmark iterations
	fmt.Printf("Running %d iterations...\n", *iterations)
	for i := 0; i < *iterations; i++ {
		if err := runner.Run(); err != nil {
			log.Fatal("Iteration failed:", err)
		}
		if (i+1)%50 == 0 {
			fmt.Printf("Completed %d/%d iterations\n", i+1, *iterations)
		}
	}

	// Memory profiling
	if *memprofile != "" {
		fmt.Printf("Writing memory profile to %s...\n", *memprofile)
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("Could not create memory profile:", err)
		}
		defer f.Close()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("Could not write memory profile:", err)
		}
	}

	fmt.Println("Profiling complete!")
	fmt.Println("\nTo analyze the profiles, use:")
	if *cpuprofile != "" {
		fmt.Printf("  go tool pprof %s\n", *cpuprofile)
	}
	if *memprofile != "" {
		fmt.Printf("  go tool pprof %s\n", *memprofile)
	}
}


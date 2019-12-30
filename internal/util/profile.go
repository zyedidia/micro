package util

import (
	"fmt"
	"log"
	"runtime"
	"time"

	humanize "github.com/dustin/go-humanize"
)

// GetMemStats returns a string describing the memory usage and gc time used so far
func GetMemStats() string {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	return fmt.Sprintf("Alloc: %s, Sys: %s, GC: %d, PauseTotalNs: %dns", humanize.Bytes(memstats.Alloc), humanize.Bytes(memstats.Sys), memstats.NumGC, memstats.PauseTotalNs)
}

func Tic(s string) time.Time {
	log.Println("START:", s)
	return time.Now()
}

func Toc(start time.Time) {
	end := time.Now()
	log.Println("END: ElapsedTime in seconds:", end.Sub(start))
}

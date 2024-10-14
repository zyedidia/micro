package util

import (
	"fmt"
	"runtime"

	humanize "github.com/dustin/go-humanize"
)

// GetMemStats returns a string describing the memory usage and gc time used so far
func GetMemStats() string {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	return fmt.Sprintf("Alloc: %s, Sys: %s, GC: %d, PauseTotalNs: %dns", humanize.Bytes(memstats.Alloc), humanize.Bytes(memstats.Sys), memstats.NumGC, memstats.PauseTotalNs)
}

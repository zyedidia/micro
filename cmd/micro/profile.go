package main

import (
	"fmt"
	"runtime"

	humanize "github.com/dustin/go-humanize"
)

func GetMemStats() string {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)
	return fmt.Sprintf("Alloc: %s, Sys: %s, GC: %d, PauseTotalNs: %dns", humanize.Bytes(memstats.Alloc), humanize.Bytes(memstats.Sys), memstats.NumGC, memstats.PauseTotalNs)
}

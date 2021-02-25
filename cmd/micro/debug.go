package main

import (
	"log"
	"os"

	"github.com/zyedidia/micro/v2/internal/util"
)

// NullWriter simply sends writes into the void
type NullWriter struct{}

// Write is empty
func (NullWriter) Write(data []byte) (n int, err error) {
	return 0, nil
}

// InitLog sets up the debug log system for micro if it has been enabled by compile-time variables
func InitLog() {
	if util.Debug == "ON" {
		f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		log.SetOutput(f)
		log.Println("Micro started")
	} else {
		log.SetOutput(NullWriter{})
	}
}

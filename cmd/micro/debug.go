package main

import (
	"log"
	"os"
)

type NullWriter struct{}

func (NullWriter) Write(data []byte) (n int, err error) {
	return 0, nil
}

func InitLog() {
	if Debug == "ON" {
		f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		log.SetOutput(f)
		log.Println("Micro started")
	} else {
		log.SetOutput(NullWriter{})
		log.Println("Micro started")
	}
}

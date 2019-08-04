package config

import (
	"log"
	"time"
)

var Autosave chan bool

func init() {
	Autosave = make(chan bool)
}

func StartAutoSave() {
	go func() {
		for {
			autotime := time.Duration(GlobalSettings["autosave"].(float64))
			if autotime < 1 {
				break
			}
			time.Sleep(autotime * time.Second)
			log.Println("Autosave")
			Autosave <- true
		}
	}()
}

func StopAutoSave() {
	GlobalSettings["autosave"] = float64(0)
}

package config

import (
	"sync"
	"time"
)

var Autosave chan bool
var autotime float64

// lock for autosave
var autolock sync.Mutex

func init() {
	Autosave = make(chan bool)
}

func SetAutoTime(a float64) {
	autolock.Lock()
	autotime = a
	autolock.Unlock()
}

func StartAutoSave() {
	go func() {
		autolock.Lock()
		a := autotime
		autolock.Unlock()
		if a <= 0 {
			return
		}
		for {
			time.Sleep(time.Duration(a * float64(time.Second)))
			autolock.Lock()
			a := autotime
			autolock.Unlock()
			if a <= 0 {
				break
			}
			Autosave <- true
		}
	}()
}

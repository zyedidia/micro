package config

import (
	"sync"
	"time"
)

var Autosave chan bool
var autotime int

// lock for autosave
var autolock sync.Mutex

func init() {
	Autosave = make(chan bool)
}

func SetAutoTime(a int) {
	autolock.Lock()
	autotime = a
	autolock.Unlock()
}

func GetAutoTime() int {
	autolock.Lock()
	a := autotime
	autolock.Unlock()
	return a
}

func StartAutoSave() {
	go func() {
		for {
			autolock.Lock()
			a := autotime
			autolock.Unlock()
			if a < 1 {
				break
			}
			time.Sleep(time.Duration(a) * time.Second)
			Autosave <- true
		}
	}()
}

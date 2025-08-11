package config

import (
	"time"
)

var Autosave chan bool
var autotime chan float64

func init() {
	Autosave = make(chan bool)
	autotime = make(chan float64)
}

func SetAutoTime(a float64) {
	autotime <- a
}

func StartAutoSave() {
	go func() {
		var a float64
		var t *time.Timer
		var elapsed <-chan time.Time
		for {
			select {
			case a = <-autotime:
				if t != nil {
					t.Stop()
					for len(elapsed) > 0 {
						<-elapsed
					}
				}
				if a > 0 {
					if t != nil {
						t.Reset(time.Duration(a * float64(time.Second)))
					} else {
						t = time.NewTimer(time.Duration(a * float64(time.Second)))
						elapsed = t.C
					}
				}
			case <-elapsed:
				if a > 0 {
					t.Reset(time.Duration(a * float64(time.Second)))
					Autosave <- true
				}
			}
		}
	}()
}

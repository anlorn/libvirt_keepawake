package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

// Monitors all VMs and inhibints/uninhibits sleep when needed
type Orchestrator struct {
	sleepInhibitor           SleepInhibitor
	libvirtWatcher           *LibvirtWatcher
	ticker                   *time.Ticker
	done                     chan bool
	currentInhibitorsCookies map[string]uint32
}

func NewOrchestrator(sleepInhibitor SleepInhibitor, libvirtWatcher *LibvirtWatcher, ticker *time.Ticker) *Orchestrator {
	return &Orchestrator{
		sleepInhibitor:           sleepInhibitor,
		libvirtWatcher:           libvirtWatcher,
		ticker:                   ticker,
		currentInhibitorsCookies: make(map[string]uint32, 1),
	}
}

// Start Run the main loop of the orchestrator and start checking libvirt for VMs to
// inhibit and inhibit sleep
func (o *Orchestrator) Start() {
	o.done = make(chan bool)
	go func() {
		var cookie uint32
		var success bool
		for {
			select {
			case timeNow := <-o.ticker.C:
				activeDomains, err := o.libvirtWatcher.GetActiveDomains()
				if err != nil {
					log.Error("Can't list active domains")
					continue
				}
				if len(activeDomains) > 0 {
					for _, activeDomain := range activeDomains {
						if _, found := o.currentInhibitorsCookies[activeDomain]; !found {
							cookie, success, err = o.sleepInhibitor.Inhibit(activeDomain)
							if err != nil {
								log.Errorf("Can't inhibit sleep with err %s", err)
								continue
							}
							if !success {
								log.Info("Can't inhibit sleep")
								continue
							}
							o.currentInhibitorsCookies[activeDomain] = cookie
						}
					}
				} else if len(activeDomains) == 0 && cookie != 0 {
					err := o.sleepInhibitor.UnInhibit(cookie)
					if err != nil {
						log.Errorf("Can't uninhibit sleep with err %s", err)
					}
				}
				fmt.Println("time", timeNow)
			case <-o.done:
				// TODO cleanup
				log.Debug("Got stop signal for orchestrator, will clean all inhibitors")
				o.ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops main loop of the orchestrator and stop do a cleanup
func (o *Orchestrator) Stop() {
	o.done <- true
	close(o.done)
}

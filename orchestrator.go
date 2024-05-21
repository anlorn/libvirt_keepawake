package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

// Monitors all VMs and inhibints/uninhibits sleep when needed
type Orchestrator struct {
	sleepInhibitor SleepInhibitor
	libvirtWatcher *LibvirtWatcher
	ticker         *time.Ticker
	done           chan bool
}

func NewOrchestrator(sleepInhibitor SleepInhibitor, libvirtWatcher *LibvirtWatcher, ticker *time.Ticker) *Orchestrator {
	return &Orchestrator{sleepInhibitor: sleepInhibitor, libvirtWatcher: libvirtWatcher, ticker: ticker}
}

// Start Run the main loop of the orchestrator and start checking libvirt for VMs to
// inhibit and inhibit sleep
func (o *Orchestrator) Start() {
	o.done = make(chan bool)
	go func() {
		for {
			select {
			case timeNow := <-o.ticker.C:
				// TODO checks VMs and activate/deactive sleep inhibitor

				// TODO activate/deactive sleep inhibitor, monitor VMs
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

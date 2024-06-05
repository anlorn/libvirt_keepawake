package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	libvirtLibrary "libvirt.org/go/libvirt"
	"libvirt_keepawake/internal"
	"libvirt_keepawake/internal/libvirt_watcher"
	"os"
	"os/signal"
	"syscall"
	"time"

	dbus "github.com/godbus/dbus/v5"
)

func main() {
	log.SetLevel(log.InfoLevel)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	conn, err := dbus.SessionBusPrivateNoAutoStartup()
	if err != nil {
		log.Error("Can't connect to session DBUS")
		os.Exit(1)
	} else {
		log.Info("Successfully connected to session DBUS")
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Error("Can't close DBUS connection")
		}
	}()
	sleepInhibitor := internal.NewDbusSleepInhibitor(conn)

	// how to listen for libvirt event
	libVirtConn, libVirtConErr := libvirtLibrary.NewConnect("qemu:///system")
	if libVirtConErr != nil {
		log.Error("Can't connect to libvirt")
		os.Exit(1)
	} else {
		log.Debug("Successfully connected to libvirt")
		defer func() {
			_, err := libVirtConn.Close()
			if err != nil {
				log.Error("Can't close libvirt connection")
			}
		}()
	}
	connAdapter := libvirt_watcher.LibvirtConnectAdapter{Connect: libVirtConn}
	watcher := libvirt_watcher.NewLibvirtWatcher(&connAdapter)

	ticker := time.NewTicker(10 * time.Second)

	orchestrator := internal.NewOrchestrator(sleepInhibitor, watcher, ticker)
	orchestrator.Start()
	defer func() {
		log.Debug("Stopping orchestrator")
		orchestrator.Stop()
	}()
	log.Info("Will wait for interrupt signal")
	v, ok := <-ch
	if !ok {
		fmt.Println(ok)
	}
	fmt.Println(v)
}

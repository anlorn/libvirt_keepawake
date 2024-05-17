package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	libvirtLibrary "libvirt.org/go/libvirt"
	"os"
	"os/signal"
	"syscall"

	dbus "github.com/godbus/dbus/v5"
)

var activeInhibitors = make(map[string]uint32)

func main() {
	log.SetLevel(log.DebugLevel)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	conn, err := dbus.ConnectSessionBus()
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
	sleepInhibitor := NewDbusSleepInhibitor(conn)

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
	connAdapter := LibvirtConnectAdapter{libVirtConn}
	watcher := NewLibvirtWatcher(&connAdapter)
	activeDomainsNames, err := watcher.GetActiveDomains()
	if err != nil {
		log.Error("Can't list active domains")
		os.Exit(1)
	} else {
		log.Debug("Successfully listed active domains")
	}
	for _, domainName := range activeDomainsNames {
		logWithDomain := log.WithFields(log.Fields{"domain_name": domainName})
		logWithDomain.Debug("Found active domain")
		_, found := activeInhibitors[domainName]
		if found {
			logWithDomain.Debug("Already inhibited")
			continue
		}
		cookie, success, err := sleepInhibitor.Inhibit(domainName)
		if err != nil {
			logWithDomain.Error("Can't inhibit sleep")
			continue
		}
		if !success {
			logWithDomain.Info("Can't inhibit sleep")
			continue
		}
		activeInhibitors[domainName] = cookie
	}

	log.Info("Will wait for interrupt signal")
	v, ok := <-ch
	if !ok {
		fmt.Println(ok)
	}
	fmt.Println(v)
}

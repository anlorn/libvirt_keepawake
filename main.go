package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"libvirt.org/go/libvirt"

	log "github.com/sirupsen/logrus"

	dbus "github.com/godbus/dbus/v5"
)

func main() {
	log.SetLevel(log.DebugLevel)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		log.Error("Can't connect to session DBUS")
		os.Exit(1)
	} else {
		log.Info("Succesfully connected to session DBUS")
	}
	defer conn.Close()

	object := conn.Object("org.freedesktop.PowerManagement", "/org/freedesktop/PowerManagement/Inhibit")
	call := object.Call("org.freedesktop.PowerManagement.Inhibit.Inhibit", 0, "Custom Libviccrt script", "Test")
	fmt.Println(call.Body)
	go func() {
		for {
			fmt.Printf("%v+\n", time.Now())
			time.Sleep(time.Second)
		}
	}()

	// how to listen for libvirt event
	libVirtConn, libVirtConErr := libvirt.NewConnect("qemu:///system")
	if libVirtConErr != nil {
		log.Error("Can't connect to libvirt")
		os.Exit(1)
	} else {
		log.Debug("Succesfully connected to libvirt")
	}
	activeDomains, err := libVirtConn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Error("Can't list active domains")
		os.Exit(1)
	} else {
		log.Debug("Succesfully listed active domains")
	}
	for _, domain := range activeDomains {
		name, _ := domain.GetName()
		log.WithFields(
			log.Fields{"domain_name": name},
		).Debug("Found active domain")
	}

	v, ok := <-ch
	if !ok {
		fmt.Println(ok)
	}
	fmt.Println(v)
}

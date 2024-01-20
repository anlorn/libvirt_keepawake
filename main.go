package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import "github.com/godbus/dbus/v5"
import log "github.com/sirupsen/logrus"

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
	call := object.Call("org.freedesktop.PowerManagement.Inhibit.Inhibit", 0, "Custom Libvirt script", "Test")
	fmt.Println(call.Body)
	go func() {
		for {
			fmt.Printf("%v+\n", time.Now())
			time.Sleep(time.Second)
		}
	}()

	v, ok := <-ch
	if !ok {
		fmt.Println(ok)
	}
	fmt.Println(v)
}

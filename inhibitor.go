package main

import "github.com/godbus/dbus/v5"

type SleepInhibitor interface {
	Inhibit(reason string) (cookie int32, success bool, err error)
	UnInhibit(cookie int32) (success bool, err error)
}

type DbusSleepInhibitor struct {
	dbusConnection *dbus.Conn
}

func NewDbusSleepInhibitor(dbusConnection *dbus.Conn) SleepInhibitor {
	return &DbusSleepInhibitor{
		dbusConnection: dbusConnection,
	}
}

func (*DbusSleepInhibitor) Inhibit(reason string) (cookie int32, success bool, err error) {
	return 0, false, nil
}

func (*DbusSleepInhibitor) UnInhibit(cookie int32) (success bool, err error) {
	return false, nil
}

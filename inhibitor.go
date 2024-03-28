package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
)

type SleepInhibitor interface {
	Inhibit(reason string) (cookie int32, success bool, err error)
	// GetInhibitors returns a list of current inhibitors where every element of the list is a string with application
	// name which is inhibiting the sleep. This method is available for xfce4-power-manager and gnome-power-manager.
	// but might not be available in other cases.
	GetInhibitors() (inhibitors []string, err error)
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

func (d *DbusSleepInhibitor) Inhibit(reason string) (cookie int32, success bool, err error) {
	obj := d.dbusConnection.Object(
		"org.freedesktop.PowerManagement.Inhibit",
		"/org/freedesktop/PowerManagement/Inhibit",
	)
	method := "org.freedesktop.PowerManagement.Inhibit.Inhibit"
	call := obj.Call(method, 0, "YourAppName", reason)
	if call.Err != nil {
		logrus.Error("Can't call DBUS method. Err %s", call.Err)
		return 0, false, call.Err
	}
	err = call.Store(&cookie)
	if err != nil {
		logrus.Error("Can't retrieve or store cookie. Err %s", err)
		return 0, false, err
	}
	logrus.Debugf("Inhibit cookie: %d", cookie)
	return cookie, true, nil
}

func (d *DbusSleepInhibitor) GetInhibitors() (inhibitors []string, err error) {
	obj := d.dbusConnection.Object(
		"org.freedesktop.PowerManagement.Inhibit",
		"/org/freedesktop/PowerManagement/Inhibit",
	)
	method := "org.freedesktop.PowerManagement.Inhibit.GetInhibitors"
	call := obj.Call(method, 0)
	if call.Err != nil {
		logrus.Error("Can't call DBUS method %s. Err %s", method, call.Err)
		return inhibitors, call.Err
	}
	err = call.Store(&inhibitors)
	if err != nil {
		logrus.Error("Can't retrieve or store inhibitors. Err %s", err)
		return inhibitors, err
	}
	logrus.Debugf("Current Inhibitors: %v", inhibitors)
	return inhibitors, nil
}

func (d *DbusSleepInhibitor) UnInhibit(cookie int32) (success bool, err error) {
	return false, nil
}

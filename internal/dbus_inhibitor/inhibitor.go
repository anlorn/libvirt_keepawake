package dbus_inhibitor

import (
	dbus "github.com/godbus/dbus/v5"
	"github.com/sirupsen/logrus"
)

const dbusDest string = "org.freedesktop.PowerManagement"
const dbusPath dbus.ObjectPath = "/org/freedesktop/PowerManagement/Inhibit"

type SleepInhibitor interface {
	Inhibit(appName string) (cookie uint32, success bool, err error)
	// GetInhibitors returns a list of current inhibitors where every element of the list is a string with application
	// name which is inhibiting the sleep. This method is available for xfce4-power-manager and gnome-power-manager.
	// but might not be available in other cases.
	GetInhibitors() (inhibitors []string, err error)
	UnInhibit(cookie uint32) (err error)
}

type DbusSleepInhibitor struct {
	dbusConnection *dbus.Conn
}

func NewDbusSleepInhibitor(dbusConnection *dbus.Conn) SleepInhibitor {
	return &DbusSleepInhibitor{
		dbusConnection: dbusConnection,
	}
}

func (d *DbusSleepInhibitor) Inhibit(appName string) (cookie uint32, success bool, err error) {
	obj := d.dbusConnection.Object(
		dbusDest,
		dbusPath,
	)
	dBusMethod := "org.freedesktop.PowerManagement.Inhibit.Inhibit"
	logrus.Debugf("Will inhibit sleep for app %s by calling %s", appName, dBusMethod)
	err = obj.Call(dBusMethod, 0, appName, "VM is running").Store(&cookie)
	logrus.Debugf("Called to inhibit sleep and got cookie: %d", cookie)
	if err != nil {
		logrus.Errorf("Can't retrieve or store cookie. Err %s", err)
		return 0, false, err
	}
	logrus.Debugf("Inhibit cookie: %d", cookie)
	return cookie, true, nil
}

func (d *DbusSleepInhibitor) GetInhibitors() (inhibitors []string, err error) {
	obj := d.dbusConnection.Object(
		dbusDest,
		dbusPath,
	)
	dBusMethod := "org.freedesktop.PowerManagement.Inhibit.GetInhibitors"
	call := obj.Call(dBusMethod, 0)
	if call.Err != nil {
		logrus.Errorf("Can't call DBUS dBusMethod %s. Err %s", dBusMethod, call.Err)
		return inhibitors, call.Err
	}
	err = call.Store(&inhibitors)
	if err != nil {
		logrus.Errorf("Can't retrieve or store inhibitors. Err %s", err)
		return inhibitors, err
	}
	logrus.Debugf("Current Inhibitors: %v", inhibitors)
	return inhibitors, nil
}

func (d *DbusSleepInhibitor) UnInhibit(cookie uint32) (err error) {
	dBusMethod := "org.freedesktop.PowerManagement.Inhibit.UnInhibit"
	obj := d.dbusConnection.Object(dbusDest, dbusPath)
	call := obj.Call(dBusMethod, 0, cookie)
	if call.Err != nil {
		logrus.Infof(
			"Can't call DBUS dBusMethod %s. Might be okay if inhibitor doesn't exists. Err %s",
			dBusMethod, call.Err,
		)
		return call.Err
	}
	return nil
}

package dbus_inhibitor

import "github.com/godbus/dbus/v5"

type DbusConnectionManager struct {
	dbusConnection *dbus.Conn
}

func NewDbusConnectionManager(conn *dbus.Conn) *DbusConnectionManager {
	return &DbusConnectionManager{
		dbusConnection: conn,
	}
}

func (m *DbusConnectionManager) getConnection() (conn *dbus.Conn, err error) {
	return m.dbusConnection, nil
}

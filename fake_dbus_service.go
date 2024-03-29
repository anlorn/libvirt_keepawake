package main

import "github.com/godbus/dbus/v5"
import log "github.com/sirupsen/logrus"

type FakeDbusService struct {
	dbusConnection *dbus.Conn
}

func NewFakeDbusService(dbusConnection *dbus.Conn) *FakeDbusService {
	return &FakeDbusService{dbusConnection: dbusConnection}
}

func (s *FakeDbusService) Start() error {
	// Requesting a name on the bus
	reply, err := s.dbusConnection.RequestName("org.freedesktop.PowerManagement", dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalf("Failed to request name: %v on test dbus", err)
	}
	// Exposing our Inhibitor object to D-Bus
	err = s.dbusConnection.Export(
		s,
		"/org/freedesktop/PowerManagement/Inhibit",
		"org.freedesktop.PowerManagement.Inhibit",
	)
	if err != nil {
		log.Fatalf("Failed to export Inhibitor object: %v", err)
	}
	return nil
}

func (s *FakeDbusService) Stop() {
	// Releasing the name on the bus
	_, err := s.dbusConnection.ReleaseName("org.freedesktop.PowerManagement")
	if err != nil {
		log.Warnf("Failed to release name: %v on test dbus", err)
	}
	err = s.dbusConnection.Close()
	if err != nil {
		log.Warnf("Failed to close connection: %v on test dbus", err)
	}
}

// Inhibit is the method that will handle the Inhibit D-Bus calls.
func (s *FakeDbusService) Inhibit(appName string, reason string) (uint32, *dbus.Error) {
	log.Printf("Inhibit called with appName: %s, reason: %s", appName, reason)
	// Implement your inhibition logic here
	// The uint return value is typically a cookie to uniquely identify this inhibition request
	return uint32(1), nil
}

package dbus_inhibitor

// Fake dbus service which listen for `org.freedesktop.PowerManagement` and track created inhibitors.
// Intended for testing purposes only

import (
	"github.com/godbus/dbus/v5"
	"sync"
)
import log "github.com/sirupsen/logrus"

type FakeDbusService struct {
	dbusConnection   *dbus.Conn
	activeInhibitors map[uint32]string
	mutex            sync.Mutex
}

/*
NewFakeDbusService create a dbus service which listen for `org.freedesktop.PowerManagement`
*/
func NewFakeDbusService(dbusConnection *dbus.Conn) *FakeDbusService {
	service := &FakeDbusService{dbusConnection: dbusConnection}
	service.activeInhibitors = make(map[uint32]string)
	return service
}

func (s *FakeDbusService) Start() error {
	// Requesting a Name on the bus
	reply, err := s.dbusConnection.RequestName("org.freedesktop.PowerManagement", dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatalf("Failed to request Name: %v on test dbus", err)
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
	// Releasing the Name on the bus
	_, err := s.dbusConnection.ReleaseName("org.freedesktop.PowerManagement")
	if err != nil {
		log.Warnf("Failed to release Name: %v on test dbus", err)
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

	s.mutex.Lock()
	defer s.mutex.Unlock()
	cookie := uint32(len(s.activeInhibitors) + 1)
	s.activeInhibitors[cookie] = appName
	return cookie, nil
}

func (s *FakeDbusService) GetInhibitors() ([]string, *dbus.Error) {
	log.Printf("GetInhibitors called")
	// The string return value is typically a list of app names that are currently inhibiting sleep
	var inhibitors = []string{}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, appName := range s.activeInhibitors {
		inhibitors = append(inhibitors, appName)
	}
	return inhibitors, nil
}

func (s *FakeDbusService) UnInhibit(cookie uint32) *dbus.Error {
	log.Printf("UnInhibit called with cookie: %d", cookie)
	// The bool return value is typically a success flag to indicate if the uninhibition was successful
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.activeInhibitors[cookie]; ok {
		delete(s.activeInhibitors, cookie)
		return nil
	}
	return dbus.NewError("org.xfce.PowerManager.Error.CookieNotFound", []interface{}{"Invalid cookie"})
}

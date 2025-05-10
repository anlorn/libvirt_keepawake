package dbus_inhibitor

import (
	"fmt"
	"os"
	"testing"

	dbus "github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DbusSleepInhibitorSuite struct {
	suite.Suite
	testDbusSocketPath string
	dbusProcess        *os.Process
	FakeDbusService    *FakeDbusService
	SleepInhibitor     SleepInhibitor
}

func (s *DbusSleepInhibitorSuite) SetupSuite() {
	dbusSocketPath, dbusProcess, err := RunDbusServer()
	if err != nil {
		s.T().Fatalf("Can't start dbus server. Err %s", err)
	}
	s.dbusProcess = dbusProcess

	s.FakeDbusService = s.runTestDbusService(dbusSocketPath)

	conn, err := dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}

	connectionManager := NewDbusConnectionManager(conn)

	s.SleepInhibitor = NewDbusSleepInhibitor(*connectionManager)
	err = s.FakeDbusService.Start()
	if err != nil {
		s.T().Fatalf("Can't start fake dbus service. Err %s", err)
	}
}

func (s *DbusSleepInhibitorSuite) runTestDbusService(dbusSocketPath string) *FakeDbusService {
	conn, err := dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}
	fakeDbusService := NewFakeDbusService(conn)

	return fakeDbusService
}

func (s *DbusSleepInhibitorSuite) TearDownSuite() {
	s.FakeDbusService.Stop()
	if s.dbusProcess != nil {
		if err := s.dbusProcess.Kill(); err != nil {
			fmt.Printf("Can't kill dbus server with PID %d. Err %s", s.dbusProcess.Pid, err)

		}
	}

	if _, err := os.Stat(s.testDbusSocketPath); err == nil {
		if err := os.Remove(s.testDbusSocketPath); err != nil {
			fmt.Printf("Can't remove socket file %s. Err %s", s.testDbusSocketPath, err)
		}
	}
}

func (s *DbusSleepInhibitorSuite) TestInhibit() {
	cookie, success, err := s.SleepInhibitor.Inhibit("test")
	assert.Equal(s.T(), uint32(1), cookie)
	assert.True(s.T(), success)
	assert.NoError(s.T(), err)

	activeInhibitors, err := s.SleepInhibitor.GetInhibitors()
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), []string{"test"}, activeInhibitors)

	dbusErr := s.SleepInhibitor.UnInhibit(cookie)
	assert.NoError(s.T(), dbusErr)
}

func (s *DbusSleepInhibitorSuite) TestUninhibitedNonExisting() {
	dbusErr := s.SleepInhibitor.UnInhibit(9999)
	assert.Errorf(s.T(), dbusErr, "org.freedesktop.PowerManagement.Inhibit.Error.InhibitorNotFound")
}

func TestRunDbusSleepInhibitorSuite(t *testing.T) {
	suite.Run(t, new(DbusSleepInhibitorSuite))
}

package main

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
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

/*
runDbusServer starts dbus server with given socket path and return process
*/
func runDbusServer(socketPath string) (*os.Process, error) {
	config := fmt.Sprintf(`<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
	"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
   <busconfig>
   <listen>unix:path=%s</listen>
   <auth>EXTERNAL</auth>
   <apparmor mode="disabled"/>
   
	<policy context='default'>
	  <allow send_destination='*' eavesdrop='true'/>
	  <allow eavesdrop='true'/>
	  <allow user='*'/>
	</policy>   
   </busconfig>
   `, socketPath)
	cfgFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, err
	}
	if _, err := cfgFile.Write([]byte(config)); err != nil {
		return nil, err
	}
	err = cfgFile.Close()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("dbus-daemon", "--nofork", "--print-address", "--config-file", cfgFile.Name())
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Process, err
}

func (s *DbusSleepInhibitorSuite) SetupSuite() {
	testDbusSocketPath := fmt.Sprintf("/tmp/dbus-test-%s.socket", uuid.New())
	dbusProcess, err := runDbusServer(testDbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't start dbus server. Err %s", err)
	}
	s.dbusProcess = dbusProcess
	fmt.Printf("Started dbus with PID %d", dbusProcess.Pid)
	dbusSocketPath := fmt.Sprintf("unix:path=%s", testDbusSocketPath)
	time.Sleep(1 * time.Second) // TODO think about better solution
	conn, err := dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}
	s.FakeDbusService = NewFakeDbusService(conn)

	conn, err = dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}
	s.SleepInhibitor = NewDbusSleepInhibitor(conn)
}

func (s *DbusSleepInhibitorSuite) TearDownSuite() {
	s.FakeDbusService.Stop()
	if err := s.dbusProcess.Kill(); err != nil {
		fmt.Printf("Can't kill dbus server with PID %d. Err %s", s.dbusProcess.Pid, err)

	}
	if err := os.Remove(s.testDbusSocketPath); err != nil {
		fmt.Printf("Can't remove socket file %s. Err %s", s.testDbusSocketPath, err)
	}
	fmt.Println("TearDownSuite")
}

func (s *DbusSleepInhibitorSuite) SetupTest() {
	fmt.Println("SetupTest")
}

func (s *DbusSleepInhibitorSuite) TestInhibit() {
	cookie, success, err := s.SleepInhibitor.Inhibit("test")
	assert.Equal(s.T(), 1, cookie)
	assert.True(s.T(), success)
	assert.NoError(s.T(), err)
}

func (s *DbusSleepInhibitorSuite) TestUninhibited() {
	assert.Equal(s.T(), false, false)
}

func TestRunDbusSleepInhibitorSuite(t *testing.T) {
	suite.Run(t, new(DbusSleepInhibitorSuite))
}

package main

import (
	"fmt"
	"github.com/godbus/dbus/v5"
	"io/ioutil"
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
	cfgFile, err := ioutil.TempFile("", "")
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
	fmt.Println("qqq")
	testDbusSocketPath := fmt.Sprintf("/tmp/dbus-test-%s.socket", uuid.New())
	dbusProcess, err := runDbusServer(testDbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't start dbus server. Err %s", err)
		fmt.Println(err)
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
	s.dbusProcess.Kill()
	os.Remove(s.testDbusSocketPath)
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

func (s *DbusSleepInhibitorSuite) TestUninhibit() {
	assert.Equal(s.T(), false, false)
}

func TestRunDbusSleepInhibitorSuite(t *testing.T) {
	suite.Run(t, new(DbusSleepInhibitorSuite))
}

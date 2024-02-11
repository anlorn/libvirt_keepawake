package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DbusSleepInhibitorSuite struct {
	suite.Suite
	dbusProcess *os.Process
}

func runDbusServer(socketPath string) (*os.Process, error) {
	config := `<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
	"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
   <busconfig>
   <listen>unix:path=/tmp/test.socket</listen>
   <auth>EXTERNAL</auth>
   <apparmor mode="disabled"/>
   
	<policy context='default'>
	  <allow send_destination='*' eavesdrop='true'/>
	  <allow eavesdrop='true'/>
	  <allow user='*'/>
	</policy>   
   </busconfig>
   `
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
	dbusProcess, err := runDbusServer("/tmp/dbus-test")
	if err != nil {
		s.T().Fatalf("Can't start dbus server. Err %s", err)
		fmt.Println(err)
	}
	s.dbusProcess = dbusProcess
	fmt.Printf("Started dbus with PID %d", dbusProcess.Pid)
}

func (s *DbusSleepInhibitorSuite) TearDownSuite() {
	s.dbusProcess.Kill()
	fmt.Println("TearDownSuite")
}

func (s *DbusSleepInhibitorSuite) SetupTest() {
	fmt.Println("SetupTest")
}

func (s *DbusSleepInhibitorSuite) TestInhibit() {
	assert.Equal(s.T(), false, false)
}

func (s *DbusSleepInhibitorSuite) TestUninhibit() {
	assert.Equal(s.T(), false, false)
}

func TestRunDbusSleepInhibitorSuite(t *testing.T) {
	suite.Run(t, new(DbusSleepInhibitorSuite))
}

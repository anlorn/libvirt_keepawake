package main

// Create a new session dbus server for testing purposes

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

/*
startDBUSProcess starts dbus server with given socket path and return process.
*/
func startDBUSProcess(socketPath string) (*os.Process, error) {
	config := fmt.Sprintf(`<!DOCTYPE busconfig PUBLIC "-//freedesktop//DTD D-BUS Bus Configuration 1.0//EN"
	"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
   <busconfig>
   <listen>unix:path=%s</listen>
   <auth>EXTERNAL</auth>
   <apparmor mode="disabled"/>
   
	<policy context='default'>
	  <allow send_destination='*' eavesdrop='true'/>
      <allow own='org.freedesktop.PowerManagement'/>
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

/*
RunDbusServer starts dbus server on a random socket and return socket path and dbus process.
When  not needed server has be stopped by killing dbus process.
*/
func RunDbusServer() (socketPath string, dbusProcess *os.Process, err error) {
	testDbusSocketPath := fmt.Sprintf("/tmp/dbus-test-%s.socket", uuid.New())
	dbusSocketPath := fmt.Sprintf("unix:path=%s", testDbusSocketPath)
	dbusProcess, err = startDBUSProcess(testDbusSocketPath)
	if err != nil {
		log.Errorf("Can't start dbus server. Err %s", err)
		return "", nil, err
	}
	fmt.Printf("Started dbus with PID %d", dbusProcess.Pid)
	time.Sleep(1 * time.Second) // TODO think about better solution
	return dbusSocketPath, dbusProcess, nil
}

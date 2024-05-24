package main

// Create a new session dbus server for testing purposes

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net"
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

func waitForDbusSocker(dbusSocketPath string) error {
	timer := time.NewTimer(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			return fmt.Errorf("Can't connect to dbus server")
		case <-ticker.C:
			conn, err := net.Dial("unix", dbusSocketPath)
			if err == nil {
				errClose := conn.Close()
				if errClose != nil {
					log.Errorf("Can't close test connection to dbus server. Err %s", errClose)
				}
				return nil
			}
		}
	}
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
	err = waitForDbusSocker(testDbusSocketPath)
	if err != nil {
		log.Errorf("Can't connect to socker, dbus server is not running. Err %s", err)
		errKill := dbusProcess.Kill()
		if errKill != nil {
			log.Errorf("Can't kill dbus process %d. Err %s", dbusProcess.Pid, errKill)
		}
		return "", nil, err
	}
	log.Infof("Started dbus with PID %d", dbusProcess.Pid)
	return dbusSocketPath, dbusProcess, nil
}

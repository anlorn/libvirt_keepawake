package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type OrchestratorSuite struct {
	suite.Suite
	sleepInhibitor  SleepInhibitor
	watcher         *LibvirtWatcher
	libvirtConnect  *FakeLibvirtConnect
	orchestrator    *Orchestrator
	dbusProcess     *os.Process
	fakeDbusService *FakeDbusService
}

func (s *OrchestratorSuite) SetupTest() {
	s.libvirtConnect = new(FakeLibvirtConnect)
	s.watcher = NewLibvirtWatcher(s.libvirtConnect)

	dbusSocketPath, dbusProcess, err := RunDbusServer()
	s.dbusProcess = dbusProcess

	conn, err := dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}
	s.fakeDbusService = NewFakeDbusService(conn)

	conn, err = dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}
	s.sleepInhibitor = NewDbusSleepInhibitor(conn)
	err = s.fakeDbusService.Start()
	if err != nil {
		s.T().Fatalf("Can't start fake dbus service. Err %s", err)
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	s.orchestrator = NewOrchestrator(s.sleepInhibitor, s.watcher, ticker)
	s.orchestrator.Start()
}

func (s *OrchestratorSuite) TearDownTest() {
	if s.orchestrator != nil {
		s.orchestrator.Stop()
	}
	if s.dbusProcess != nil {
		if err := s.dbusProcess.Kill(); err != nil {
			s.T().Fatalf("Can't kill dbus server with PID %d. Err %s", s.dbusProcess.Pid, err)
		}
	}
}

func (s *OrchestratorSuite) TestInhibitOnDomainActivation() {
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(activeInhibitors))
	s.libvirtConnect.domains = []MinimalLibvirtDomain{FakeLibvirtDomain{"domain1"}}
	time.Sleep(1 * time.Second) // TODO think about better solution
	activeInhibitors, err = s.fakeDbusService.GetInhibitors()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{"domain1"}, activeInhibitors)

}

func TestRunOrchestratorSuite(t *testing.T) {
	suite.Run(t, new(OrchestratorSuite))
}

package internal

import (
	"fmt"
	"libvirt_keepawake/internal/dbus_inhibitor"
	"libvirt_keepawake/internal/libvirt_watcher"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OrchestratorSuite struct {
	suite.Suite
	sleepInhibitor  dbus_inhibitor.SleepInhibitor
	watcher         *libvirt_watcher.LibvirtWatcher
	libvirtConnect  *libvirt_watcher.FakeLibvirtConnect
	orchestrator    *Orchestrator
	dbusProcess     *os.Process
	fakeDbusService *dbus_inhibitor.FakeDbusService
}

// SetupTest is a setup function that initializes the necessary components for testing the Orchestrator.
func (s *OrchestratorSuite) SetupTest() {
	// Initialize a new fake libvirt connection.
	s.libvirtConnect = new(libvirt_watcher.FakeLibvirtConnect)

	// Create a new libvirt watcher using the fake libvirt connection.
	s.watcher = libvirt_watcher.NewLibvirtWatcher(s.libvirtConnect)

	// Run a dbus server for testing and get the socket path and process.
	dbusSocketPath, dbusProcess, err := dbus_inhibitor.RunDbusServer()
	if err != nil {
		s.T().Fatalf("Can't start dbus server. Err %s", err)
	}
	s.dbusProcess = dbusProcess

	// Connect to the test dbus server.
	conn, err := dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}

	// Create a new fake dbus service using the connected dbus connection.
	s.fakeDbusService = dbus_inhibitor.NewFakeDbusService(conn)

	// Connect to the test dbus server again.
	conn, err = dbus.Connect(dbusSocketPath)
	if err != nil {
		s.T().Fatalf("Can't connect to test dbus server. Err %s", err)
	}

	// Create a new dbus sleep inhibitor using the connected dbus connection.
	s.sleepInhibitor = dbus_inhibitor.NewDbusSleepInhibitor(conn)

	// Start the fake dbus service.
	err = s.fakeDbusService.Start()
	if err != nil {
		s.T().Fatalf("Can't start fake dbus service. Err %s", err)
	}

	// Create a new ticker with a 100ms interval.
	ticker := time.NewTicker(500 * time.Millisecond)

	// Create a new orchestrator using the sleep inhibitor, libvirt watcher, and ticker.
	s.orchestrator = NewOrchestrator(s.sleepInhibitor, s.watcher, ticker)

	// Start the orchestrator.
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

// TestInhibitOnDomainActivation tests the orchestrator's ability to inhibit sleep when a domain is activated.
func (s *OrchestratorSuite) TestInhibitOnDomainActivation() {
	// Get the initial list of active inhibitors.
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// Simulate a domain activation by adding a domain to the libvirt connection.
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{libvirt_watcher.FakeLibvirtDomain{Name: "domain1"}},
	)

	// Wait for the orchestrator to process the domain activation.
	s.assertActiveInhibitors([]string{"domain1"})
}

// TestDuplicateDomains tests the orchestrator's ability to handle multiple domains with the same name.
// It is kinda a limitation that I use domain name to identify a domain. but works for me.
// Orchestrator should create one inhibitor for multiple domains with the same name.
// And uninhibit sleep only when all domains with the same name are deactivated.
func (s *OrchestratorSuite) TestDuplicateDomains() {
	// Get the initial list of active inhibitors.
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// Simulate a domain activation by adding a domain to the libvirt connection.
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{libvirt_watcher.FakeLibvirtDomain{Name: "domain1"}},
	)

	// Wait for the orchestrator to process the domain activation.
	s.assertActiveInhibitors([]string{"domain1"})

	// activate second domain with the same name
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"}, libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
		},
	)
	// check we still have only one inhibitor
	s.assertActiveInhibitors([]string{"domain1"})

	// remove one of two duplicate domains
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
		},
	)
	// check we still have inhibitor
	s.assertActiveInhibitors([]string{"domain1"})

	// remove all duplicate domains
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{},
	)
	// inhibitor should be removed
	s.assertActiveInhibitors([]string{})

}

func (s *OrchestratorSuite) TestUnInhibitOnDomainDeactivation() {
	// Get the initial list of active inhibitors.
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// Simulate a domain activation by adding a domain to the libvirt connection.
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{libvirt_watcher.FakeLibvirtDomain{Name: "domain1"}},
	)
	s.assertActiveInhibitors([]string{"domain1"})

	// Deactivate the domain by removing it from the libvirt connection.
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{},
	)
	s.assertActiveInhibitors([]string{})
}

// TestMultipleDomainsActivation tests the orchestrator's ability to inhibit sleep when multiple domains are activated.
// and then remove inhibitors when all domains are deactivated.
func (s *OrchestratorSuite) TestMultipleDomainsActivation() {
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// activate first domain and check inhibitor is activated
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
		},
	)
	s.assertActiveInhibitors([]string{"domain1"})

	// activate second domain and check inhibitor is activated for both domains
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
			libvirt_watcher.FakeLibvirtDomain{Name: "domain2"},
		},
	)
	s.assertActiveInhibitors([]string{"domain1", "domain2"})

	// Deactivate the first domain by removing it from the libvirt connection and check inhibitor is deactivated

	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain2"},
		},
	)
	s.assertActiveInhibitors([]string{"domain2"})

	// deactivate all domains and check inhibitors are deactivated
	s.libvirtConnect.UpdateActiveDomains([]libvirt_watcher.MinimalLibvirtDomain{})
	s.assertActiveInhibitors([]string{})
}

func (s *OrchestratorSuite) TestRemoveInhibitorsOnStoppingOrchestrator() {
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// activate first domain and check inhibitor is activated
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
		},
	)
	s.assertActiveInhibitors([]string{"domain1"})

	s.orchestrator.Stop()
	// domain is still active but all inhibitors are deactivated, because the orchestrator is stopped
	s.assertActiveInhibitors([]string{})
}

func (s *OrchestratorSuite) TestAddInhibitorsOnStoppingOrchestrator() {
	activeInhibitors, err := s.fakeDbusService.GetInhibitors()
	// Assert that there are no errors.
	assert.Nil(s.T(), err)
	// Assert that there are no active inhibitors initially.
	assert.Equal(s.T(), 0, len(activeInhibitors))

	// activate first domain and check inhibitor is activated
	s.libvirtConnect.UpdateActiveDomains(
		[]libvirt_watcher.MinimalLibvirtDomain{
			libvirt_watcher.FakeLibvirtDomain{Name: "domain1"},
		},
	)
	s.assertActiveInhibitors([]string{"domain1"})

	s.orchestrator.Stop()
	// domain is still active but all inhibitors are deactivated, because the orchestrator is stopped
	s.assertActiveInhibitors([]string{})
}

// assertActiveInhibitors asserts that expectedInhibitors is equal to the list of active inhibitors.
// function has some time period when it waits for the inhibitors to be activated.
func (s *OrchestratorSuite) assertActiveInhibitors(expectedInhibitors []string) {
	ticker := time.NewTicker(100 * time.Millisecond)
	timer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			activeInhibitors, err := s.sleepInhibitor.GetInhibitors()
			assert.Nil(s.T(), err)
			sort.Strings(activeInhibitors)
			fmt.Println("active inhibitors", activeInhibitors)
			// use this to compare inhibitors disregard of an order
			sortTransformer := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if cmp.Equal(expectedInhibitors, activeInhibitors, sortTransformer) {
				timer.Stop()
				ticker.Stop()
				return
			}
		case <-timer.C:
			ticker.Stop()
			s.T().Fatalf("Expetected inhibitors %v weren't activated", expectedInhibitors)
		}
	}
}

func TestRunOrchestratorSuite(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	suite.Run(t, new(OrchestratorSuite))
}

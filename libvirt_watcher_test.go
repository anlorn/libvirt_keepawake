package main

import (
	"github.com/stretchr/testify/suite"
	"libvirt.org/go/libvirt"
	"testing"
)

type LibvirtWatcherSuite struct {
	suite.Suite
	libvirtConnect *libvirt.Connect
}

func (s *LibvirtWatcherSuite) TestGetActiveDomains() {
	// prepare
	watcher := NewLibvirtWatcher(s.libvirtConnect)

	// act
	err, activeDomains := watcher.GetActiveDomains()

	// assert
	s.Assert().NoError(err)
	s.Assert().EqualValues(activeDomains, []string{"domain1", "domain2"})

}

func TestRunLibvirtWatcherSuite(t *testing.T) {
	suite.Run(t, new(LibvirtWatcherSuite))
}

package main

import (
	"github.com/stretchr/testify/suite"
	"libvirt.org/go/libvirt"
	"testing"
)

type LibvirtWatcherSuite struct {
	suite.Suite
}

func (s *LibvirtWatcherSuite) TestGetActiveDomains() {
	// prepare
	fakeLibvirtConnect := new(FakeLibvirtConnect)
	domain1 := FakeLibvirtDomain{}
	domain2 := FakeLibvirtDomain{}
	domain1.On("GetName").Return("domain1", nil)
	domain2.On("GetName").Return("domain2", nil)
	fakeLibvirtConnect.On("ListAllDomains", libvirt.CONNECT_LIST_DOMAINS_ACTIVE).Return([]MinimalLibvirtDomain{
		domain1,
		domain2,
	}, nil)
	watcher := NewLibvirtWatcher(fakeLibvirtConnect)

	// act
	activeDomains, err := watcher.GetActiveDomains()

	// assert
	s.Assert().NoError(err)
	s.Assert().EqualValues(activeDomains, []string{"domain1", "domain2"})

}

func TestRunLibvirtWatcherSuite(t *testing.T) {
	suite.Run(t, new(LibvirtWatcherSuite))
}

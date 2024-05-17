package main

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type LibvirtWatcherSuite struct {
	suite.Suite
}

func (s *LibvirtWatcherSuite) TestGetActiveDomains() {
	// prepare
	firstDomainName := "domain1"
	secondDomainName := "domain2"
	domain1 := FakeLibvirtDomain{firstDomainName}
	domain2 := FakeLibvirtDomain{secondDomainName}
	fakeLibvirtConnect := new(FakeLibvirtConnect)
	fakeLibvirtConnect.domains = []MinimalLibvirtDomain{domain1, domain2}
	watcher := NewLibvirtWatcher(fakeLibvirtConnect)

	// act
	activeDomains, err := watcher.GetActiveDomains()

	// assert
	s.Assert().NoError(err)
	s.Assert().EqualValues(activeDomains, []string{firstDomainName, secondDomainName})

}

func TestRunLibvirtWatcherSuite(t *testing.T) {
	suite.Run(t, new(LibvirtWatcherSuite))
}

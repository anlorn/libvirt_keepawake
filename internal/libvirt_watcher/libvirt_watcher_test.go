package libvirt_watcher

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type LibvirtWatcherSuite struct {
	suite.Suite
}

func (s *LibvirtWatcherSuite) TestGetActiveDomains() {
	// prepare
	firstDomainName := "domain1"
	secondDomainName := "domain2"
	domain1 := FakeLibvirtDomain{Name: firstDomainName}
	domain2 := FakeLibvirtDomain{Name: secondDomainName}
	fakeLibvirtConnect := new(FakeLibvirtConnect)
	fakeLibvirtConnect.Domains = []MinimalLibvirtDomain{domain1, domain2}
	watcher := NewLibvirtWatcher(fakeLibvirtConnect)

	// act
	activeDomains, err := watcher.GetActiveDomains()

	// assert
	s.Assert().NoError(err)
	s.Assert().EqualValues(
		activeDomains,
		[]MinimalLibvirtDomain{
			FakeLibvirtDomain{Name: firstDomainName},
			FakeLibvirtDomain{Name: secondDomainName},
		})

}

func TestRunLibvirtWatcherSuite(t *testing.T) {
	suite.Run(t, new(LibvirtWatcherSuite))
}

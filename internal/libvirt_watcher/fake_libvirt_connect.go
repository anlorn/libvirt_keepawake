package libvirt_watcher

import (
	"fmt"
	"libvirt.org/go/libvirt"
	"sync"
)
import "github.com/stretchr/testify/mock"

// fake for libvirt.connect
type FakeLibvirtConnect struct {
	mock.Mock
	mu      sync.Mutex
	domains []MinimalLibvirtDomain
}

func (f *FakeLibvirtConnect) ListAllDomains(
	flags libvirt.ConnectListAllDomainsFlags,
) ([]MinimalLibvirtDomain, error) {
	if flags != libvirt.CONNECT_LIST_DOMAINS_ACTIVE {
		return nil, fmt.Errorf("not implemented, only active Domains ae supported in fake")
	}
	f.mu.Lock()
	activeDomains := make([]MinimalLibvirtDomain, len(f.domains))
	copy(activeDomains, f.domains)
	f.mu.Unlock()
	return activeDomains, nil
}

// UpdateActiveDomains updates the list of active domains in the fake libvirt connection. Should be called
// only from tests
func (f *FakeLibvirtConnect) UpdateActiveDomains(domains []MinimalLibvirtDomain) {
	f.mu.Lock()
	f.domains = domains
	f.mu.Unlock()
}

type FakeLibvirtDomain struct {
	Name string
}

func (f FakeLibvirtDomain) GetName() (string, error) {
	return f.Name, nil
}

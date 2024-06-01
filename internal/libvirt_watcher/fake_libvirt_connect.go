package libvirt_watcher

import (
	"fmt"
	"libvirt.org/go/libvirt"
)
import "github.com/stretchr/testify/mock"

// fake for libvirt.connect
type FakeLibvirtConnect struct {
	mock.Mock
	Domains []MinimalLibvirtDomain
}

func (f *FakeLibvirtConnect) ListAllDomains(
	flags libvirt.ConnectListAllDomainsFlags,
) ([]MinimalLibvirtDomain, error) {
	if flags != libvirt.CONNECT_LIST_DOMAINS_ACTIVE {
		return nil, fmt.Errorf("not implemented, only active Domains ae supported in fake")
	}
	return f.Domains, nil
}

type FakeLibvirtDomain struct {
	Name string
}

func (f FakeLibvirtDomain) GetName() (string, error) {
	return f.Name, nil
}

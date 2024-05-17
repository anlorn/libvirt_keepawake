package main

import (
	"fmt"
	"libvirt.org/go/libvirt"
)
import "github.com/stretchr/testify/mock"

// fake for libvirt.connect
type FakeLibvirtConnect struct {
	mock.Mock
	domains []MinimalLibvirtDomain
}

func (f *FakeLibvirtConnect) ListAllDomains(
	flags libvirt.ConnectListAllDomainsFlags,
) ([]MinimalLibvirtDomain, error) {
	if flags != libvirt.CONNECT_LIST_DOMAINS_ACTIVE {
		return nil, fmt.Errorf("not implemented, only active domains ae supported in fake")
	}
	return f.domains, nil
}

type FakeLibvirtDomain struct {
	name string
}

func (f FakeLibvirtDomain) GetName() (string, error) {
	return f.name, nil
}

package main

import "libvirt.org/go/libvirt"
import "github.com/stretchr/testify/mock"

// fake for libvirt.connect
type FakeLibvirtConnect struct {
	mock.Mock
}

func (f *FakeLibvirtConnect) ListAllDomains(
	flags libvirt.ConnectListAllDomainsFlags,
) ([]MinimalLibvirtDomain, error) {
	args := f.Called(flags)
	return args.Get(0).([]MinimalLibvirtDomain), args.Error(1)
}

type FakeLibvirtDomain struct {
	mock.Mock
}

func (f FakeLibvirtDomain) GetName() (string, error) {
	args := f.Called()
	return args.Get(0).(string), args.Error(1)
}

package main

import "libvirt.org/go/libvirt"

// fake for libvirt.connect
type FakeLibvirtConnect struct {
	mock.Mock
}

func NewFakeLibvirtConnect() *FakeLibvirtConnect {
	return &FakeLibvirtConnect{}
}

func (f *FakeLibvirtConnect) ListAllDomains(
	flags libvirt.ConnectListAllDomainsFlags,
) ([]libvirt.Domain, error) {

}

package main

import "libvirt.org/go/libvirt"

type LibvirtWatcher struct {
	libvirtConnection *libvirt.Connect
	//TODO events to get notified when domains have been started/stopped
	// TODO interface
}

func NewLibvirtWatcher(connection *libvirt.Connect) LibvirtWatcher {
	return LibvirtWatcher{libvirtConnection: connection}
}

func (c *LibvirtWatcher) GetActiveDomains() (error, []string) {
	return nil, []string{}
}

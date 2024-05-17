package main

import "libvirt.org/go/libvirt"

type MinimalLibvirtConnect interface {
	ListAllDomains(flags libvirt.ConnectListAllDomainsFlags) ([]MinimalLibvirtDomain, error)
}

type LibvirtConnectAdapter struct {
	connect *libvirt.Connect
}

func (a *LibvirtConnectAdapter) ListAllDomains(flags libvirt.ConnectListAllDomainsFlags) ([]MinimalLibvirtDomain, error) {
	domains, err := a.connect.ListAllDomains(flags)
	if err != nil {
		return nil, err
	}
	domainsAdapter := make([]MinimalLibvirtDomain, len(domains))
	for i, domain := range domains {
		domainsAdapter[i] = LibvirtDomainAdapter{&domain}
	}
	return domainsAdapter, nil
}

type MinimalLibvirtDomain interface {
	GetName() (string, error)
}

type LibvirtDomainAdapter struct {
	domain *libvirt.Domain
}

func (a LibvirtDomainAdapter) GetName() (string, error) {
	return a.domain.GetName()
}

type LibvirtWatcher struct {
	libvirtConnection MinimalLibvirtConnect
	//TODO events to get notified when domains have been started/stopped
	// TODO interface
}

func NewLibvirtWatcher(connection MinimalLibvirtConnect) LibvirtWatcher {
	return LibvirtWatcher{libvirtConnection: connection}
}

func (c *LibvirtWatcher) GetActiveDomains() ([]string, error) {
	domains, err := c.libvirtConnection.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return nil, err
	}
	domainsNames := make([]string, len(domains))
	for i, domain := range domains {
		domainName, err := domain.GetName()
		if err != nil {
			return nil, err
		}
		domainsNames[i] = domainName
	}
	return domainsNames, nil
}

package libvirt_watcher

import (
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

type MinimalLibvirtConnect interface {
	ListAllDomains(flags libvirt.ConnectListAllDomainsFlags) ([]MinimalLibvirtDomain, error)
}

type LibvirtConnectAdapter struct {
	Connect *libvirt.Connect
}

func (a *LibvirtConnectAdapter) ListAllDomains(flags libvirt.ConnectListAllDomainsFlags) ([]MinimalLibvirtDomain, error) {
	domains, err := a.Connect.ListAllDomains(flags)
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

func (a LibvirtDomainAdapter) String() string {
	name, err := a.GetName()
	if err != nil {
		log.WithError(err).Error("Can't get name of domain")
		return "unknown"
	}
	return name
}

type LibvirtWatcher struct {
	libvirtConnection MinimalLibvirtConnect
	//TODO events to get notified when domains have been started/stopped
	// TODO interface
}

func NewLibvirtWatcher(connection MinimalLibvirtConnect) *LibvirtWatcher {
	return &LibvirtWatcher{libvirtConnection: connection}
}

func (c *LibvirtWatcher) GetActiveDomains() ([]MinimalLibvirtDomain, error) {
	domains, err := c.libvirtConnection.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		return nil, err
	}
	domainsNames := make([]MinimalLibvirtDomain, len(domains))
	copy(domainsNames, domains)
	return domainsNames, nil
}

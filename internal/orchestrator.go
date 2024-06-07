package internal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"libvirt_keepawake/internal/libvirt_watcher"
	"time"
)

type InhibitorName string
type InhibitorCookie uint32

// Orchestrator Monitors all VMs and inhibits/uninhibits sleep when needed
type Orchestrator struct {
	sleepInhibitor           SleepInhibitor
	libvirtWatcher           *libvirt_watcher.LibvirtWatcher
	ticker                   *time.Ticker
	done                     chan bool
	currentInhibitorsCookies map[InhibitorName]InhibitorCookie
}

func NewOrchestrator(sleepInhibitor SleepInhibitor, libvirtWatcher *libvirt_watcher.LibvirtWatcher, ticker *time.Ticker) *Orchestrator {
	return &Orchestrator{
		sleepInhibitor:           sleepInhibitor,
		libvirtWatcher:           libvirtWatcher,
		ticker:                   ticker,
		currentInhibitorsCookies: make(map[InhibitorName]InhibitorCookie, 1),
	}
}

// Start Run the main loop of the orchestrator and start checking libvirt for VMs to
// inhibit and inhibit sleep
func (o *Orchestrator) Start() {
	o.done = make(chan bool)
	go func() {
		//var cookie uint32
		//var success bool
		for {
			select {
			case <-o.ticker.C:
				log.Debug("Checking for active VMs to inhibit/uninhibit sleep")
				activeDomains, err := o.libvirtWatcher.GetActiveDomains()
				if err != nil {
					log.Error("Can't list active domains")
					continue
				}
				domainsWithoutInhibitors, err := o.determineDomainsWithoutInhibitors(activeDomains)
				if err != nil {
					log.Errorf("Can't determine domains without inhibitors. Err %s", err)
					continue
				}
				inhibitorsWithoutDomains, err := o.determineInhibitorsWithoutDomains(activeDomains)
				if err != nil {
					log.Errorf("Can't determine inhibitors without domains. Err %s", err)
					continue
				}
				for _, domainWithoutInhibitor := range domainsWithoutInhibitors {
					log.Debugf("Will actiave inhibitor for domain %s without inhibitor", domainWithoutInhibitor)
					err := o.activateInhibitorForDomain(domainWithoutInhibitor)
					if err != nil {
						log.Errorf("Can't activate inhibitor for domain %s with err %s", domainWithoutInhibitor, err)
						continue
					}
					log.Infof("Activated inhibitor for domain %s", domainWithoutInhibitor)
				}

				for _, inhibitorWithoutDomain := range inhibitorsWithoutDomains {
					err := o.deactivateInhibitor(inhibitorWithoutDomain)
					if err != nil {
						log.Errorf(
							"Can't deactivate inhibitor for domain %s with err %s",
							inhibitorWithoutDomain, err,
						)
						continue
					}
					log.Infof("Deactivated inhibitor for domain %s", inhibitorWithoutDomain)
				}
			case <-o.done: // On stop signal, clean all inhibitors
				log.Debugf(
					"Got stop signal for orchestrator, will clean all inhibitors %v", o.currentInhibitorsCookies,
				)
				for domainName, cookie := range o.currentInhibitorsCookies {
					err := o.sleepInhibitor.UnInhibit(uint32(cookie))
					if err != nil {
						log.Errorf("Can't uninhibit sleep with err %s", err)
					}
					log.Infof("Uninhibited sleep on stopping for domain %s", domainName)
				}
				o.ticker.Stop()
				break
			}
		}
	}()
}

// Stop stops main loop of the orchestrator and stop do a cleanup
func (o *Orchestrator) Stop() {
	if o.done != nil {
		o.done <- true
	}
}

/*
determineDomainsWithoutInhibitors determines all domains that don't have any active inhibitor
*/
func (o *Orchestrator) determineDomainsWithoutInhibitors(domains []libvirt_watcher.MinimalLibvirtDomain) ([]libvirt_watcher.MinimalLibvirtDomain, error) {
	var domainsWithoutInhibitors []libvirt_watcher.MinimalLibvirtDomain
	log.Debugf(
		"Will determine domains without inhibitors. Domains: %v. Current Inhibitors: %v",
		domains,
		o.currentInhibitorsCookies,
	)
	for _, domain := range domains {
		domainName, err := domain.GetName()
		if err != nil {
			log.Errorf("Can't get name of domain %s with err %s", domain, err)
			return domainsWithoutInhibitors, err
		}
		if _, found := o.currentInhibitorsCookies[InhibitorName(domainName)]; !found {
			domainsWithoutInhibitors = append(domainsWithoutInhibitors, domain)
		}
	}
	log.Debugf("Domains without inhibitors: %v", domainsWithoutInhibitors)
	return domainsWithoutInhibitors, nil
}

/*
determineInhibitorsWithoutDomains determines all inhibitors that are not
associated(doesn't have the same name as domain) with any domain
*/
func (o *Orchestrator) determineInhibitorsWithoutDomains(domains []libvirt_watcher.MinimalLibvirtDomain) ([]InhibitorName, error) {
	var inhibitorsWithoutDomains []InhibitorName
	domainsMap := map[InhibitorName]bool{}
	log.Debugf(
		"Will search for inhibitors without domains. Domains: %v. Current Inhibitors: %v",
		domains,
		o.currentInhibitorsCookies,
	)
	for _, domain := range domains {
		domainName, err := domain.GetName()
		if err != nil {
			log.Errorf("Can't get name of domain %s with err %s", domain, err)
			return inhibitorsWithoutDomains, err
		}
		domainsMap[InhibitorName(domainName)] = true
	}
	for inhibitorName := range o.currentInhibitorsCookies {
		if _, found := domainsMap[inhibitorName]; !found {
			log.Debugf("Found inhibitor %s without domain", inhibitorName)
			inhibitorsWithoutDomains = append(inhibitorsWithoutDomains, inhibitorName)
		}
	}
	log.Debugf("Inhibitors without domains: %v", inhibitorsWithoutDomains)
	return inhibitorsWithoutDomains, nil
}

/*
activateInhibitorForDomain activates an inhibitor for the given domain. Inhibitor name will be the same
as the domain name
*/
func (o *Orchestrator) activateInhibitorForDomain(domain libvirt_watcher.MinimalLibvirtDomain) error {
	domainName, err := domain.GetName()
	if err != nil {
		log.Errorf("Can't get name of domain %s, err %s", domain, err)
		return err
	}
	cookie, success, err := o.sleepInhibitor.Inhibit(domainName)
	if err != nil {
		log.Errorf("Can't inhibit sleep for domain %s with err %s", domainName, err)
		return err
	}
	if !success {
		log.Errorf("Can't inhibit sleep for domain %s", domainName)
		return fmt.Errorf("inhibition for domain %s wasn't succesfull", domainName)
	}
	o.currentInhibitorsCookies[InhibitorName(domainName)] = InhibitorCookie(cookie)
	return nil
}

/*
deactivateInhibitor deactivates an inhibitor for the given domain.
*/
func (o *Orchestrator) deactivateInhibitor(name InhibitorName) error {
	cookie, ok := o.currentInhibitorsCookies[name]
	if !ok {
		errMsg := fmt.Sprintf("Can't find cookie for inhibitor %s", name)
		log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}
	err := o.sleepInhibitor.UnInhibit(uint32(cookie))
	if err != nil {
		log.Errorf("Can't uninhibit sleep for domain %s with err %s", name, err)
		return err
	}
	delete(o.currentInhibitorsCookies, name)
	return nil
}

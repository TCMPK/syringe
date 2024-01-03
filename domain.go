package main

import (
	"fmt"
	"strings"
	"time"

	dns "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

type Domain struct {
	Record_name string `json:"Record_name" example:"google.com"`
	Record_type string `json:"Record_type" example:"A"`
	Refresh_at  int64  `json:"Refresh_at" example:"1234567"`
	index       int
}

func (domain Domain) Validate() bool {
	// we currently don't validate domain names
	if _, ok := dns.StringToType[strings.ToUpper(domain.Record_type)]; ok {
		return true
	}
	return false

}

func (domain Domain) Query(strategyContainer *ResolverStrategies, config *ResolverConfiguration, c chan<- uint) {
	for i := 0; i < len(strategyContainer.ResolveFunctions); i++ {
		strategy := strategyContainer.ResolveFunctions[i]
		ttl, err := strategy(config, &domain)
		log.Trace("resolve ", domain.ToString(), " via strategy index ", i, " yields ttl=", ttl, " err=", err)
		if err != nil {
			continue
		}
		c <- ttl
		return
	}
	c <- uint(config.StaticDelaySeconds)
}

func (domain *Domain) RefreshInSeconds(seconds uint) {
	domain.Refresh_at = time.Now().UnixMilli() + int64(seconds*1000)
}

func (domain *Domain) RefreshInMillis(millis uint64) {
	domain.Refresh_at = time.Now().UnixMilli() + int64(millis)
}

func (domain *Domain) RecordType() uint16 {
	return dns.StringToType[strings.ToUpper(domain.Record_type)]
}

func (domain Domain) MillisUntilDue() int64 {
	current_time_unix_millis := time.Now().UnixMilli()
	return domain.Refresh_at - current_time_unix_millis
}

func (domain Domain) SecondsUntilDue() int64 {
	return int64(domain.MillisUntilDue() / 1000)
}

func (domain Domain) ToString() string {
	return fmt.Sprintf("%s IN %s", domain.Record_name, domain.Record_type)
}

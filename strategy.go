package main

import (
	"errors"
	"math/rand"
	"runtime"
	"strings"

	dnsresolver "github.com/Focinfi/go-dns-resolver"
	dns "github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	domainsResolvedByStrategy = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "domains_resolved_by_strategy",
		Help: "The total number of resolved domains by strategy",
	},
		[]string{"strategy"},
	)
)

func init() {
	prometheus.Register(domainsResolvedByStrategy)
}

type ResolverStrategies struct {
	ResolveFunctions []func(config *ResolverConfiguration, domain *Domain) (uint, error)
}

func TryQueryRegularDomain(config *ResolverConfiguration, domain *Domain) (uint, error) {
	resolver := dnsresolver.NewResolver(config.ResolverIp)
	resolver.Targets(domain.Record_name).Types(dnsresolver.QueryType(domain.RecordType()))
	// Lookup
	res := resolver.Lookup()
	for target := range res.ResMap {
		for _, r := range res.ResMap[target] {
			ttl_resolved := uint(r.Ttl.Seconds())
			pc, _, _, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			domainsResolvedByStrategy.With(prometheus.Labels{"strategy": f.Name()}).Inc()
			if ttl_resolved < config.PinMinTtl {
				return config.PinMinTtl, nil
			} else {
				return uint(r.Ttl.Seconds()), nil
			}
		}
	}
	return 0, errors.New("received no rr for regular lookup")
}

func TryQuerySOADomain(config *ResolverConfiguration, domain *Domain) (uint, error) {
	resolver := dnsresolver.NewResolver(config.ResolverIp)
	resolver.Targets(domain.Record_name).Types(dnsresolver.QueryType(domain.RecordType()))

	domain_split := strings.Split(domain.Record_name, ".")
	if len(domain_split) < 2 {
		return 0, errors.New("received invalid soa domain")
	}
	soa_domain := strings.Join(strings.Split(domain.Record_name, ".")[1:], ".")
	resolver.Targets(soa_domain).Types(dnsresolver.QueryType(dns.TypeSOA))
	res := resolver.Lookup()
	for target := range res.ResMap {
		for _, r := range res.ResMap[target] {
			ttl_resolved := uint(r.Ttl.Seconds())
			pc, _, _, _ := runtime.Caller(0)
			f := runtime.FuncForPC(pc)
			domainsResolvedByStrategy.With(prometheus.Labels{"strategy": f.Name()}).Inc()
			if ttl_resolved < config.PinMinTtl {
				return config.PinMinTtl, nil
			} else {
				return uint(r.Ttl.Seconds()), nil
			}
		}
	}
	return 0, errors.New("received no rr for soa lookup")
}

func TryQueryFlexibleDelayDomain(config *ResolverConfiguration, domain *Domain) (uint, error) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	domainsResolvedByStrategy.With(prometheus.Labels{"strategy": f.Name()}).Inc()
	return uint(rand.Intn(int(config.FlexibleDelayMaxTtlSeconds-config.FlexibleDelayMinTtlSeconds)) + int(config.FlexibleDelayMinTtlSeconds)), nil
}

func TryQueryStaticDelayDomain(config *ResolverConfiguration, domain *Domain) (uint, error) {
	pc, _, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	domainsResolvedByStrategy.With(prometheus.Labels{"strategy": f.Name()}).Inc()
	return uint(config.StaticDelaySeconds), nil
}

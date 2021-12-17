package collector

import (
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	"sync"
)

var (
	resolveDomain = kingpin.Flag("resolve.domain", "Detect the domain name resolved by dns, It is recommended to add two domain names, one public domain name and one k8s service name").Default("www.allsmartcloud.com", "metrics-server.kube-system").Strings()
)

func init() {
	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)
}

type ResolveCollector struct {
	dnsResolveDesc   *prometheus.Desc
	resolveMutex     sync.Mutex
	dnsResolveResult *ResolveResult
}

type ResolveResult struct {
	DnsStatus float64
	Domain    string
}

func NewResolveCollector() prometheus.Collector {
	r := &ResolveCollector{
		dnsResolveDesc: prometheus.NewDesc("resolve_status", "resolve Status, normal when the value is 1", []string{"domain"}, nil),
		dnsResolveResult: &ResolveResult{
			DnsStatus: 0,
			Domain:    "",
		},
	}
	return r
}

func (r *ResolveCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.dnsResolveDesc
}

func (r *ResolveCollector) Collect(ch chan<- prometheus.Metric) {
	r.resolveMutex.Lock()
	defer r.resolveMutex.Unlock()
	for _, domain := range *resolveDomain {
		r.ResolveDns(domain)
		ch <- prometheus.MustNewConstMetric(r.dnsResolveDesc, prometheus.GaugeValue, r.dnsResolveResult.DnsStatus, domain)
	}
}

func (r *ResolveCollector) ResolveDns(domain string) {
	_, err := net.LookupHost(domain)
	if err != nil {
		level.Error(logger).Log("msg", "Error resolve host", "err", err)
		r.dnsResolveResult.DnsStatus = 0

	} else {
		r.dnsResolveResult.DnsStatus = 1
	}
}

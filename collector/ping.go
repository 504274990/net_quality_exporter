package collector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"strconv"
	"sync"
	"time"
)

var (
	logger log.Logger
)

func init() {
	promlogConfig := &promlog.Config{}
	logger = promlog.New(promlogConfig)
}

type PingCollector struct {
	pingStatusDesc *prometheus.Desc
	pingAvgRttDesc *prometheus.Desc
	pingDevRttDesc *prometheus.Desc
	pingMaxRttDesc *prometheus.Desc
	pingMinRttDesc *prometheus.Desc
	pingLossDesc   *prometheus.Desc
	icmpMutex      sync.Mutex
	pings          *PING
}

type PING struct {
	interval    time.Duration
	size        int
	timeout     time.Duration
	count       int
	pkginterval time.Duration
	target      string
	result      map[int]*PingResult
	pingMutex   sync.Mutex
}

type PingResult struct {
	Success   float64
	DestAddr  string
	DropRate  float64
	MinTime   time.Duration
	AvgTime   time.Duration
	MaxTime   time.Duration
	DevTime   time.Duration
	DnsStatus float64
	pkgSize   string
}

func NewPingCollector(interval time.Duration, size int, largesize int, timeout time.Duration, count int, pkginterval time.Duration, target string) prometheus.Collector {

	p := &PingCollector{
		pingStatusDesc: prometheus.NewDesc("ping_status", "Ping Status, normal when the value is 1", []string{"target", "package_size"}, nil),
		pingAvgRttDesc: prometheus.NewDesc("ping_avg_rtt_seconds", "Average Round Trip Time in seconds", []string{"target", "package_size"}, nil),
		pingDevRttDesc: prometheus.NewDesc("ping_dev_rtt_seconds", "Deviate Trip Time in seconds", []string{"target", "package_size"}, nil),
		pingMaxRttDesc: prometheus.NewDesc("ping_max_rtt_seconds", "Max Trip Time in seconds", []string{"target", "package_size"}, nil),
		pingMinRttDesc: prometheus.NewDesc("ping_min_rtt_seconds", "Min Trip Time in seconds", []string{"target", "package_size"}, nil),
		pingLossDesc:   prometheus.NewDesc("ping_loss_percent", "Packet loss in percent", []string{"target", "package_size"}, nil),
		pings: &PING{
			interval:    interval,
			size:        size,
			timeout:     timeout,
			count:       count,
			pkginterval: pkginterval,
			target:      target,
			result: map[int]*PingResult{
				size: {
					Success:   0,
					DestAddr:  target,
					DropRate:  0,
					MinTime:   0,
					AvgTime:   0,
					MaxTime:   0,
					DevTime:   0,
					DnsStatus: 0,
					pkgSize:   strconv.Itoa(size),
				},
				largesize: {
					Success:   0,
					DestAddr:  target,
					DropRate:  0,
					MinTime:   0,
					AvgTime:   0,
					MaxTime:   0,
					DevTime:   0,
					DnsStatus: 0,
					pkgSize:   strconv.Itoa(largesize),
				},
			},
		},
	}
	for _, sizes := range []int{size, largesize} {
		go p.PingRun(sizes)
	}

	return p
}

func (p *PingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.pingStatusDesc
	ch <- p.pingAvgRttDesc
	ch <- p.pingDevRttDesc
	ch <- p.pingMaxRttDesc
	ch <- p.pingMinRttDesc
	ch <- p.pingLossDesc
}

func (p *PingCollector) Collect(ch chan<- prometheus.Metric) {
	p.icmpMutex.Lock()
	defer p.icmpMutex.Unlock()
	for _, metrics := range p.pings.result {
		ch <- prometheus.MustNewConstMetric(p.pingStatusDesc, prometheus.GaugeValue, metrics.Success, metrics.DestAddr, metrics.pkgSize)
		ch <- prometheus.MustNewConstMetric(p.pingAvgRttDesc, prometheus.GaugeValue, metrics.AvgTime.Seconds(), metrics.DestAddr, metrics.pkgSize)
		ch <- prometheus.MustNewConstMetric(p.pingDevRttDesc, prometheus.GaugeValue, metrics.DevTime.Seconds(), metrics.DestAddr, metrics.pkgSize)
		ch <- prometheus.MustNewConstMetric(p.pingMaxRttDesc, prometheus.GaugeValue, metrics.MaxTime.Seconds(), metrics.DestAddr, metrics.pkgSize)
		ch <- prometheus.MustNewConstMetric(p.pingMinRttDesc, prometheus.GaugeValue, metrics.MinTime.Seconds(), metrics.DestAddr, metrics.pkgSize)
		ch <- prometheus.MustNewConstMetric(p.pingLossDesc, prometheus.GaugeValue, metrics.DropRate, metrics.DestAddr, metrics.pkgSize)
	}

}

func (p *PingCollector) PingRun(sizes int) {

	for range time.NewTicker(p.pings.interval).C {

		pinger, err := ping.NewPinger(p.pings.target)
		if err != nil {
			level.Error(logger).Log("msg", "initialization ping instance failed", "err", err)
		}

		pinger.Interval = p.pings.pkginterval
		pinger.Count = p.pings.count
		pinger.Size = p.pings.size
		pinger.Timeout = p.pings.timeout
		pinger.SetPrivileged(true)

		err = pinger.Run()
		if err != nil {
			p.pings.result[sizes].Success = 0
			level.Error(logger).Log("msg", "execute ping progress failed", "err", err)
		} else {
			p.pings.result[sizes].Success = 1
		}
		stats := pinger.Statistics()

		p.pings.result[sizes].AvgTime = stats.AvgRtt
		p.pings.result[sizes].MaxTime = stats.MaxRtt
		p.pings.result[sizes].MinTime = stats.MinRtt
		p.pings.result[sizes].DevTime = stats.StdDevRtt
		p.pings.result[sizes].DropRate = stats.PacketLoss
		p.pings.result[sizes].pkgSize = strconv.Itoa(sizes)
	}
}

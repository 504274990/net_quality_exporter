package main

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"net_quality_exporter/collector"
	"os"
)

const version string = "0.0.1"

var (
	listenAddress   = kingpin.Flag("web.listen-address", "Address to listen on for web interface.").Default(":9231").String()
	metricsPath     = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	pingInterval    = kingpin.Flag("ping.interval", "Each ping runner interval time.").Default("60s").Duration()
	pingSize        = kingpin.Flag("ping.size", "The size of each ping packet.").Default("32").Int()
	pingLargeSize   = kingpin.Flag("ping.largesize", "The size of each large ping packet.").Default("1000").Int()
	pingTimeout     = kingpin.Flag("ping.timeout", "The entire timeout period of each ping runner.").Default("20s").Duration()
	pingCount       = kingpin.Flag("ping.count", "The number of packets sent by each ping runner.").Default("10").Int()
	pingPkgInterval = kingpin.Flag("ping.pkg.interval", "The interval of each ping packet.").Default("500ms").Duration()
	pingTarget      = kingpin.Flag("ping.target", "The interval of each ping packet, must be domain name, default is aliyun.com's address.").Default("106.11.172.9").String()
	logger          log.Logger
)

func main() {
	promlogConfig := &promlog.Config{}
	logger := promlog.New(promlogConfig)
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version)
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector.NewPingCollector(*pingInterval, *pingSize, *pingLargeSize, *pingTimeout, *pingCount, *pingPkgInterval, *pingTarget))
	reg.MustRegister(collector.NewResolveCollector())
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	level.Info(logger).Log("msg", "Starting network_exporter", "version", version)
	http.Handle("/metrics", h)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Net Quality Exporter</title></head>
            <body>
            <h1>Net Quality Exporter</h1>
            <p><a href='` + *metricsPath + `'>Metrics</a></p>
            </body>
            </html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	server := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(server, "", logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/dgzlopes/prototype/pkg/protod"
	"github.com/dgzlopes/prototype/pkg/prototype"
	"github.com/dgzlopes/prototype/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/signals"
)

var (
	protodEnabled bool
	cfg           util.Config
	svcs          []services.Service
	tmpTags       string
)

func init() {
	cfg = util.Config{}
	flag.BoolVar(&protodEnabled, "d", false, "To run protod")

	flag.StringVar(&cfg.PrometheusPath, "prometheus-path", "/metrics", "The path to publish Prometheus metrics to.")
	flag.StringVar(&cfg.PrometheusListenAddress, "prometheus-listen-address", ":80", "The address to listen on for Prometheus scrapes.")

	// Prototype-specific flags
	flag.StringVar(&cfg.PrototypeURL, "prototype-url", "http://localhost:10000", "The URL (scheme://hostname) at which to find Prototype.")

	// Protod-specific flags
	flag.DurationVar(&cfg.RefreshWait, "refresh-duration", 15*time.Second, "The amount of time to pause between config refreshes")
	flag.StringVar(&cfg.Cluster, "cluster", "default", "")
	flag.StringVar(&cfg.Service, "service", "envoy", "")
	flag.StringVar(&tmpTags, "tags", "", "")
}

func main() {
	flag.Parse()
	cfg.Tags = strings.Split(tmpTags, ",")

	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	rand.Seed(time.Now().UnixNano())

	if protodEnabled {
		p, err := protod.New(&cfg, logger)
		if err != nil {
			logger.Error("Error instantiating protod:", err)
			os.Exit(1)
		}
		logger.Info("Starting protod.")
		svcs = append(svcs, p...)
	} else {
		p, err := prototype.New(&cfg, logger)
		if err != nil {
			logger.Error("Error instantiating prototype:", err)
			os.Exit(1)
		}
		logger.Info("Starting prototype.")
		svcs = append(svcs, p...)
	}

	m, _ := services.NewManager(svcs...)
	m.StartAsync(context.Background())
	handler := signals.NewHandler(logging.Global())
	go func() {
		handler.Loop()
		m.StopAsync()
	}()
	m.AwaitStopped(context.Background())
}

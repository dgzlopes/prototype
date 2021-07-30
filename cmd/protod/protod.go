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
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/signals"
)

var (
	cfg     protod.Config
	svcs    []services.Service
	tmpTags string
)

func init() {
	cfg = protod.Config{}
	flag.StringVar(&cfg.PrometheusPath, "prometheus-path", "/metrics", "The path to publish Prometheus metrics to.")
	flag.StringVar(&cfg.PrometheusListenAddress, "prometheus-listen-address", ":80", "The address to listen on for Prometheus scrapes.")

	flag.StringVar(&cfg.PrototypeURL, "prototype-url", "http://localhost:10000", "The URL (scheme://hostname) at which to find Prototype.")
	flag.DurationVar(&cfg.RefreshWait, "refresh-duration", 15*time.Second, "The amount of time to pause between config refreshes")
	flag.StringVar(&cfg.Name, "name", "default", "")
	flag.StringVar(&tmpTags, "tags", "", "")
	cfg.Tags = strings.Split(tmpTags, ",")
}

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	rand.Seed(time.Now().UnixNano())

	p, err := protod.New(&cfg, logger)
	if err != nil {
		logger.Error("Error instantiating protod:", err)
		os.Exit(1)
	}
	logger.Info("Starting protod.")
	svcs = append(svcs, p...)

	m, _ := services.NewManager(svcs...)
	m.StartAsync(context.Background())
	handler := signals.NewHandler(logging.Global())
	go func() {
		handler.Loop()
		m.StopAsync()
	}()
	m.AwaitStopped(context.Background())
}

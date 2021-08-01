package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/dgzlopes/prototype/pkg/prototype"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/signals"
)

var (
	cfg  prototype.Config
	svcs []services.Service
)

func init() {
	cfg = prototype.Config{}
	flag.StringVar(&cfg.PrometheusPath, "prometheus-path", "/metrics", "The path to publish Prometheus metrics to.")
	flag.StringVar(&cfg.PrometheusListenAddress, "prometheus-listen-address", ":80", "The address to listen on for Prometheus scrapes.")
}

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	rand.Seed(time.Now().UnixNano())

	p, err := prototype.New(&cfg, logger)
	if err != nil {
		logger.Error("Error instantiating prototype:", err)
		os.Exit(1)
	}
	logger.Info("Starting prototype.")
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

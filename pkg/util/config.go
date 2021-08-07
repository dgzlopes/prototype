package util

import "time"

type Config struct {
	PrometheusPath          string
	PrometheusListenAddress string

	// Prototype (Control plane)
	PrototypeURL string

	// ProtoD (Data plane)
	Cluster     string
	Service     string
	Tags        []string
	RefreshWait time.Duration
}

package util

import "time"

type Config struct {
	PrometheusPath          string
	PrometheusListenAddress string

	// Prototype (Control plane)
	Cluster string
	Service string
	Tags    []string

	PrototypeURL string
	RefreshWait  time.Duration

	// ProtoD (Data plane)
}

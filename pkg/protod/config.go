package protod

import "time"

type Config struct {
	PrometheusPath          string
	PrometheusListenAddress string

	Cluster string
	Service string
	Tags    []string

	PrototypeURL string
	RefreshWait  time.Duration
}

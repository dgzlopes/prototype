package protod

import "time"

type Config struct {
	PrometheusPath          string
	PrometheusListenAddress string

	Name string
	Tags []string

	PrototypeURL string
	RefreshWait  time.Duration
}

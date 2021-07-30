package util

import (
	"io/ioutil"
)

type PrototypeRequest struct {
	Cluster   string     `json:"cluster"`
	Service   string     `json:"service"`
	ID        string     `json:"id"`
	Tags      []string   `json:"tags"`
	EnvoyInfo *EnvoyInfo `json:"envoy_info"`
}

type EnvoyInfo struct {
	Version string `json:"version"`
	State   string `json:"state"`
	Uptime  string `json:"uptime"`
}

// WriteToDisk writes some data to disk
func WriteToDisk(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

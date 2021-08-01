package util

import (
	"bytes"
	"encoding/gob"
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

type HTTPpayload struct {
	Cluster string `json:"cluster"`
	Service string `json:"service"`
	Type    string `json:"type"`
	Config  string `json:"config"`
}

// WriteToDisk writes some data to disk
func WriteToDisk(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (m PrototypeRequest) EncodeToBytes() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeFromBytes(q []byte) (PrototypeRequest, error) {
	var rm PrototypeRequest
	buffer := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(&rm)
	if err != nil {
		return PrototypeRequest{}, err
	}
	return rm, nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// Configuration base struct
type Configuration struct {
	Node             *Node             `yaml:"node"`
	DynamicResources *DynamicResources `yaml:"dynamic_resources"`
	Admin            *Admin            `yaml:"admin"`
}

// Node is used for instance identification purposes
type Node struct {
	Cluster string `yaml:"cluster"`
	ID      string `yaml:"id"`
}

// DynamicResources specify where to load dynamic configuration from.
type DynamicResources struct {
	CDSConfig *ConfigSource `yaml:"cds_config"`
	LDSConfig *ConfigSource `yaml:"lds_config"`
}

// ConfigSource for each xDS API source
type ConfigSource struct {
	Path string `yaml:"path"`
}

// Admin interface config
type Admin struct {
	AccessLogPath string  `yaml:"access_log_path"`
	Adress        *Adress `yaml:"address"`
}

// Adress is the TCP address that the administration server will listen on.
type Adress struct {
	SocketAdress *SocketAdress `yaml:"socket_address"`
}

// SocketAdress config about the socket
type SocketAdress struct {
	Adress    string `yaml:"address"`
	PortValue int    `yaml:"port_value"`
}

// WriteConfigFile writes a config to disk
func WriteConfigFile(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// GenerateConfig from CLI flags
func GenerateConfig() *Configuration {
	return &Configuration{
		&Node{
			Cluster: "example",
			ID:      "node-1",
		},
		&DynamicResources{
			CDSConfig: &ConfigSource{
				Path: "/tmp/cds.yaml",
			},
			LDSConfig: &ConfigSource{
				Path: "/tmp/lds.yaml",
			},
		},
		&Admin{
			AccessLogPath: "/dev/null",
			Adress: &Adress{
				SocketAdress: &SocketAdress{
					Adress:    "0.0.0.0",
					PortValue: 19000,
				},
			},
		},
	}
}

type ProtoD struct {
	Name       string      `json:"name"`
	Tags       []string    `json:"tags"`
	ServerInfo *ServerInfo `json:"server_info"`
}

type ServerInfo struct {
	Version string `json:"version"`
	State   string `json:"state"`
}

// GetDynamicConfig gets all the configs from the control plane
func (cf *Configuration) GetDynamicConfig(endpoint string) error {
	json_data, err := json.Marshal(ProtoD{
		Name: "quote",
		Tags: []string{"env:production", "version:0.0.6-beta"},
		ServerInfo: &ServerInfo{
			Version: "1",
			State:   "2",
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(endpoint+"/api/protod", "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	for k, v := range res {
		if strings.Contains(k, "/cds/") {
			f, _ := os.Create("/tmp/cds_temp.yaml")
			_, _ = f.WriteString(v)
			f.Close()
			os.Rename("/tmp/cds_temp.yaml", "/tmp/cds.yaml")
		}
		if strings.Contains(k, "/lds/") {
			f, _ := os.Create("/tmp/lds_temp.yaml")
			_, _ = f.WriteString(v)
			f.Close()
			os.Rename("/tmp/lds_temp.yaml", "/tmp/lds.yaml")
		}
	}
	return nil
}

// GetDynamicConfigCycle gets the config from the control plane on some interval
func (cf *Configuration) GetDynamicConfigCycle(wg *sync.WaitGroup, endpoint string, refresh int) {
	defer wg.Done()
	for {
		time.Sleep(time.Duration(refresh) * time.Second)
		err := cf.GetDynamicConfig(endpoint)
		if err != nil {
			fmt.Printf("Failed to get dynamic config %v\n", err)
			continue
		}
		fmt.Println("Getting dynamic config from control plane...")
	}
}

// RunEnvoy runs envoy binary on the background
func RunEnvoy(wg *sync.WaitGroup, args ...string) {
	defer wg.Done()
	cmd := exec.Command("envoy", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
}
func main() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	endpoint := "http://localhost:10000"
	refresh := 10

	// Check if we can find envoy binary
	_, err := exec.LookPath("envoy")
	if err != nil {
		log.Fatalf("Envoy lookpath failed with %s\n", err)
	}

	generatedConfig := GenerateConfig()
	config, err := yaml.Marshal(generatedConfig)
	if err != nil {
		log.Fatalf("Failed to config %s\n", err)
	}

	err = WriteConfigFile("config.yaml", config)
	if err != nil {
		log.Fatalf("Failed to write config to file %s\n", err)
	}

	// Create initial files
	f, err := os.Create("/tmp/cds.yaml")
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	f, err = os.Create("/tmp/lds.yaml")
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	generatedConfig.GetDynamicConfig(endpoint)
	//go RunEnvoy(wg, "-c", "config.yaml", "-l", "debug")
	go generatedConfig.GetDynamicConfigCycle(wg, endpoint, refresh)
	go RunEnvoy(wg, "-c", "config.yaml")

	wg.Wait()
}

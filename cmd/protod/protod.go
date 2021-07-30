package main

import (
	"bytes"
	"encoding/json"
	"flag"
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

var (
	prometheusListenAddress string
	prometheusPath          string

	prototypeURL    string
	refreshDuration time.Duration
)

func init() {
	flag.StringVar(&prometheusPath, "prometheus-path", "/metrics", "The path to publish Prometheus metrics to.")
	flag.StringVar(&prometheusListenAddress, "prometheus-listen-address", ":80", "The address to listen on for Prometheus scrapes.")

	flag.StringVar(&prototypeURL, "prototype-url", "", "The URL (scheme://hostname) at which to find Prototype.")
	flag.DurationVar(&refreshDuration, "refresh-duration", 15*time.Second, "The amount of time to pause between config refreshes")
}

func main() {
	flag.Parse()

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

// WriteConfigFile writes a config to disk
func WriteConfigFile(fileName string, data []byte) error {
	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
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
		return err
	}

	resp, err := http.Post(endpoint+"/api/protod", "application/json",
		bytes.NewBuffer(json_data))

	if err != nil {
		return err
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
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	cmd.Start()

	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	cmd.Wait()
}

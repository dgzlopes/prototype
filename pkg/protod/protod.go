package protod

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dgzlopes/prototype/pkg/util"
	log "github.com/sirupsen/logrus"

	"github.com/cortexproject/cortex/pkg/util/services"
	"github.com/docker/docker/pkg/namesgenerator"
	"gopkg.in/yaml.v2"
)

type ProtoD struct {
	services []services.Service
	logger   *log.Logger

	id string

	envoy    *exec.Cmd
	envoycfg *EnvoyConfig

	cfg *Config
}

func New(cfg *Config, logger *log.Logger) ([]services.Service, error) {
	p := &ProtoD{
		id:     cfg.Service + "_" + namesgenerator.GetRandomName(0),
		cfg:    cfg,
		logger: logger,
	}

	// Add envoy worker
	p.services = append(p.services, services.NewBasicService(p.start, p.run, p.shutDown))

	// Add pulling worker
	p.services = append(p.services, services.NewTimerService(5*time.Second, nil, p.pull, nil))

	return p.services, nil
}

func (p *ProtoD) start(ctx context.Context) error {
	// Check if we can find envoy binary
	_, err := exec.LookPath("envoy")
	if err != nil {
		p.logger.Error("Envoy lookpath failed: %s", err)
		os.Exit(1)
	}

	// Generate and write envoy config (base)
	p.envoycfg = GenerateEnvoyConfig()
	config, err := yaml.Marshal(p.envoycfg)
	if err != nil {
		p.logger.Error("Failed to config %s", err)
		os.Exit(1)
	}
	err = util.WriteToDisk("config.yaml", config)
	if err != nil {
		p.logger.Error("Failed to write config to file %s", err)
		os.Exit(1)
	}

	// Create initial xDS files
	f, err := os.Create("/tmp/cds.yaml")
	f.Close()
	if err != nil {
		p.logger.Error(err)
		os.Exit(1)
	}
	f, err = os.Create("/tmp/lds.yaml")
	f.Close()
	if err != nil {
		p.logger.Error(err)
		os.Exit(1)
	}

	// Get configs from Prototype
	err = p.GetDynamicConfig()
	if err != nil {
		p.logger.Error("Failed to get dynamic configs: %s", err)
	}
	return nil
}

func (p *ProtoD) run(ctx context.Context) error {
	go func() {
		p.envoy = exec.Command("envoy", "--log-format", "'%v'", "-c", "config.yaml")
		stdout, _ := p.envoy.StdoutPipe()
		p.envoy.Stderr = p.envoy.Stdout
		p.envoy.Start()
		p.logger.Info("Started Envoy")
		for {
			tmp := make([]byte, 1024)
			_, err := stdout.Read(tmp)
			p.logger.WithFields(log.Fields{
				"msg": strings.ReplaceAll(strings.ReplaceAll(string(bytes.Trim(tmp, "\x00")), "\n", ""), "'", ""),
			}).Info("Envoy stdout")
			if err != nil {
				break
			}
		}
	}()
	<-ctx.Done()
	return nil
}

func (p *ProtoD) shutDown(_ error) error {
	p.logger.Info("Shutting down Envoy")
	if err := p.envoy.Process.Kill(); err != nil {
		return err
	}
	return nil
}

func (p *ProtoD) pull(ctx context.Context) error {
	p.logger.Info("Pulling configs")
	if err := p.GetDynamicConfig(); err != nil {
		p.logger.Error("Failed to get dynamic configs: %s", err)
	}
	return nil
}

// GetDynamicConfig gets all the configs from the control plane
func (p *ProtoD) GetDynamicConfig() error {
	envoyInfo, err := getEnvoyInfo()
	if err != nil {
		return err
	}

	json_data, err := json.Marshal(util.PrototypeRequest{
		Cluster: p.cfg.Cluster,
		Service: p.cfg.Service,
		ID:      p.id,
		Tags:    p.cfg.Tags,
		EnvoyInfo: &util.EnvoyInfo{
			Version: envoyInfo["version"].(string),
			State:   envoyInfo["state"].(string),
			Uptime:  envoyInfo["uptime_current_epoch"].(string),
		},
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(p.cfg.PrototypeURL+"/api/protod", "application/json", bytes.NewBuffer(json_data))

	if err != nil {
		return err
	}

	var res map[string]string

	json.NewDecoder(resp.Body).Decode(&res)

	if len(res) == 0 {
		p.logger.Info("No configs found")
		return nil
	}

	for k, v := range res {
		if strings.Contains(k, "/cds/") {
			util.WriteToDisk("/tmp/cds_temp.yaml", []byte(v))
			os.Rename("/tmp/cds_temp.yaml", "/tmp/cds.yaml")
			p.logger.Info("Refreshed CDS")
		}
		if strings.Contains(k, "/lds/") {
			util.WriteToDisk("/tmp/lds_temp.yaml", []byte(v))
			os.Rename("/tmp/lds_temp.yaml", "/tmp/lds.yaml")
			p.logger.Info("Refreshed LDS")
		}
	}
	return nil
}

// Query Envoy for the current metadata
func getEnvoyInfo() (map[string]interface{}, error) {
	r, err := http.Get("http://127.0.0.1:19000/server_info")
	if err != nil {
		return nil, err
	}
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)
	return result, nil
}

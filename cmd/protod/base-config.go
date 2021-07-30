package main

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

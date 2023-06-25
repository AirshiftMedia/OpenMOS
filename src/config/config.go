package config

// implements YAML config

// TODO: platform based paths, such as /etc/openmos/mosConfig.yaml

type Config struct {
	common struct {
		mosVersion string `yaml:"mosVersion"` // will not implement legacy features such as 10540/10541 ports
	}
	listener struct {
		mosPort string `yaml:"mosPort"`
		timeout string `yaml:"mosPortTimeout"`
	}
}

func initConfig(configPath string) (*Config, error) {
	config := &Config{}

	// open file

}

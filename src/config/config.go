package config

import (
	"errors"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// implements YAML config
// follows the example in https://github.com/koddr/example-go-config-yaml

type ConfigItems struct {
	common struct {
		mosVersion string `yaml:"mosVersion"` // will not implement legacy features such as 10540/10541 ports
	}
	listener struct {
		mosPort     string `yaml:"mosPort"`
		timeout     string `yaml:"mosPortTimeout"`
		readBuffer  string `yaml:"readBuffer"`
		writeBuffer string `yaml:"readBuffer"`
		sslKeyPath  string `yaml:"sslKeyPath"`
	}
}

type Values interface {
	getValue() string
	setValue(key string, value string)
	getCommonValues() []string
	getListenerValues() []string 
}

func initConfigPath() error {
	// inits config path on /etc/OpenMOS/mosConfig.yaml

	folderPath := "/etc/OpenMOS"

	// look for the OpenMOS folder and create if missing

	if _, err := os.Stat(folderPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}

	// look for the mosConfig.yaml and create empty if missing

	f1, err := os.Create(folderPath, "/mosConfig.yaml")
	if err != nil {
		log.Fatal(err)
	}
	f1.Close()

}

func initConfig(configPath string) (*Config, error) {
	config := &Config{}

	// open file
	file, err := os.Open(configPath)

	if err != nil {

		// if missing, attempt to recreate once (expected in first run)

		err := initConfigPath()

		file, err := os.Open(configPath)

		if err != nil {
			log.Fatal(err)
		}

		return nil, err

	}

	
	}

	defer file.Close()

	// init YAML decode and start decoding

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

}

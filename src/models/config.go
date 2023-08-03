// manage preferences with YAML encoding

package models

type ConfigItems struct {
	debugMode  bool   `yaml:debugMode`
	configPath string `yaml:configPath`
}

/*
type ConfigItems struct {
	debugMode  bool   `yaml:debugMode`
	configPath string `yaml:configPath`
	mosVersion string `yaml:"mosVersion"` // will not implement legacy features such as 10540/10541 ports

	sentryKey         string `yaml:"sentryKey"`
	sentryProject     string `yaml:"sentryProject"`
	sentryEnvironment string `yaml:"sentryEnvironment"`

	mosPort     string `yaml:"mosPort"`
	timeout     string `yaml:"mosPortTimeout"`
	readBuffer  string `yaml:"readBuffer"`
	writeBuffer string `yaml:"readBuffer"`
	sslKeyPath  string `yaml:"sslKeyPath"`
}
*/

type ConfigUtil func(*ConfigItems)

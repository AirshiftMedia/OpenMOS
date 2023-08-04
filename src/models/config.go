// manage preferences with YAML encoding

package models

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

type ConfigUtil func(*ConfigItems)

func writeConfig(c ConfigItems) bool {

	// TODO: encode in YAML and write as a file into configPath

	return true
}

func readConfig() *ConfigItems {
	return &ConfigItems{
		debugMode: true,
	}
}

func readValue(key string) string {
	return "42"
}

func writeValue(key string, value string) bool {
	return true
}

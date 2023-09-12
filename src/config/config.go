package config

import (
	"github.com/spf13/viper"
)

type commonConfig struct {
	sslKeyPath string `mapstructure:"sslKeyPath"`
}

type monitorConfig struct {
	sentryKey         string `mapstructure:"sentryKey"`
	sentryProject     string `mapstructure:"sentryProject"`
	sentryEnvironment string `mapstructure:"sentryEnvironment"`
}

type mosConfig struct {
	mosPort        string `mapstructure:"mosPort"`
	mosTimeout     string `mapstructure:"mosTimeout"`
	mosReadBuffer  string `mapstructure:"mosReadBuffer"`
	mosWriteBuffer string `mapstructure:"mosWriteBuffer"`
}

type Config struct {
	common  commonConfig  `mapstructure:"common"`
	monitor monitorConfig `mapstructure:"monitor"`
	mos     mosConfig     `mapstructure:"mos"`
}

var vconf *viper.Viper

func loadConfig() (Config, error) {

	vconf = viper.New()

	var config Config

	vconf.SetConfigName("openMosConfig")
	vconf.SetConfigType("json")
	vconf.AddConfigPath("./util")

	err := vconf.ReadInConfig()

	if err != nil {
		return Config{}, nil
	}

	err = vconf.Unmarshal(&config)

	return config, nil

}

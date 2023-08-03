package main

import (
	"fmt"
	"openmos/models"
	"os"
)

// buildVersion := 100

func defaultConfig() models.ConfigItems {
	return models.ConfigItems{
		configPath: "/etc/OpenMOS/mosConfig.yaml",
	}
}

func debugMode(conf *models.ConfigItems) {
	debugMode = true
}

type Listener struct {
	models.ConfigItems
}

func newListener(conf ...models.ConfigUtil) *Listener {
	c := defaultConfig()
	for _, fn := range conf {
		fn(&c)
	}
	return &Listener{
		ConfigItems: conf,
	}
}

func main() {

	// load config

	myConfig := readConfig()

	// check if arguments were provided when invoked

	for _, arg := range os.Args {
		fmt.Println("Arguments provided for initialization: ", arg)
	}

	fmt.Println("Starting OpenMOS service: Now listening on port ")

	l := newListener()

	fmt.Printf("%+v\n", l)

}

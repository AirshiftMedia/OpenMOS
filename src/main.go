package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

buildVersion := 100

func main() {

	// load config

	configPath := "/etc/OpenMOS/mosConfig.yaml"

	cfg, err := config.initConfig(configPath)

	if err != nil {
		log.Fatal(err)
	}

	// check if arguments were provided when invoked

	for _, arg := range os.Args {
		fmt.Println("Arguments provided for initialization: ", arg)
	}


	fmt.Println("Starting OpenMOS service: Now listening on port ")

	//

}

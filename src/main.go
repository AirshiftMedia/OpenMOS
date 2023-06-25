package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	// check if arguments were provided when invoked

	for _, arg := range os.Args {
		fmt.Println("Arguments provided for initialization: ", arg)
	}

	// load config

	configPath := "/etc/OpenMOS/mosConfig.yaml"

	cfg, err := config.initConfig(configPath)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting OpenMOS service: Now listening on port ")

	//

}

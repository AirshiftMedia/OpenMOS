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

	// init Sentry.io monitoring
	// see docs https://docs.sentry.io/platforms/go/

	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://<key>@sentry.io/<project>",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)

	hostname, err := os.Hostname()

	if err != nil {
		fmt.Println("Error while getting hostname: ", err)
		os.Exit(1)
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext("character", map[string]interface{}{
			"productName":        "OpenMOS Service",
			"buildVersion":         100,
			"hostname": hostname,
		})
	})

	sentry.CaptureMessage("OpenMOS initialized at ", hostname)

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

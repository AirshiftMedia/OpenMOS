package observer

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

// init Sentry.io monitoring
// see docs https://docs.sentry.io/platforms/go/

func initSentry() {

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
			"productName":  "OpenMOS Service",
			"buildVersion": buildVersion,
			"hostname":     hostname,
		})
	})

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("environment", "prod")
	})

	sentry.CaptureMessage("OpenMOS initialized at ", hostname)

}

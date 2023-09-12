package main

import (
	"fmt"
	"os"
)

func main() {

	// init logging with Sentry

	/* err := sentry.Init(sentry.ClientOptions{
		Dsn:           myDSN,
		EnableTracing: false,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer sentry.Flush(2 * time.Second)

	logger := slog.New(slogsentry.Option{Level: slog.LevelDebug}.NewSentryHandler())
	logger = logger.
		With("environment", "dev").
		With("release", buildVersion)

	logger.Info("Starting OpenMOS server instance ", logger.Int("mos-id", 1), logger.Int("build-version"), buildVersion)
	*/
	// init viper config

	for _, arg := range os.Args {
		fmt.Println("Arguments provided for server initialization: ", arg)
	}

	config, err := utils.loadConfig()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(config)
	}

	fmt.Println("Starting OpenMOS service: Now listening on port ")

}

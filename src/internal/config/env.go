package config

import "time"

// applyEnvOverrides applies environment variables on top of the loaded
// config. Without a YAML file, unset variables fall back to defaults.
func applyEnvOverrides(config *Config, yamlLoaded bool) {
	// App config
	if envVal := getEnv("APP_NAME", ""); envVal != "" || !yamlLoaded {
		config.App.Name = getEnv("APP_NAME", getDefaultString(config.App.Name, "OpenMOS"))
	}
	if envVal := getEnv("APP_VERSION", ""); envVal != "" || !yamlLoaded {
		config.App.Version = getEnv("APP_VERSION", getDefaultString(config.App.Version, "1.0.0"))
	}
	if envVal := getEnv("APP_ENV", ""); envVal != "" || !yamlLoaded {
		config.App.Environment = getEnv("APP_ENV", getDefaultString(config.App.Environment, "development"))
	}

	// Server config
	if envVal := getEnv("SERVER_HOST", ""); envVal != "" || !yamlLoaded {
		config.Server.Host = getEnv("SERVER_HOST", getDefaultString(config.Server.Host, "0.0.0.0"))
	}
	if envVal := getEnv("SERVER_PORT", ""); envVal != "" || !yamlLoaded {
		config.Server.Port = getEnvAsInt("SERVER_PORT", getDefaultInt(config.Server.Port, 10540)) // Default MOS port
	}
	if envVal := getEnv("SERVER_READ_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.ReadTimeout = getEnvAsDuration("SERVER_READ_TIMEOUT", getDefaultDuration(config.Server.ReadTimeout, 5*time.Second))
	}
	if envVal := getEnv("SERVER_WRITE_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.WriteTimeout = getEnvAsDuration("SERVER_WRITE_TIMEOUT", getDefaultDuration(config.Server.WriteTimeout, 5*time.Second))
	}
	if envVal := getEnv("SERVER_SHUTDOWN_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.ShutdownTimeout = getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", getDefaultDuration(config.Server.ShutdownTimeout, 30*time.Second))
	}

	// PostgreSQL config. DATABASE_URL matches the backend01 .env convention.
	if envVal := getEnv("DATABASE_URL", ""); envVal != "" || !yamlLoaded {
		config.Postgres.URL = getEnv("DATABASE_URL", config.Postgres.URL)
	}
	if envVal := getEnv("POSTGRES_MAX_CONNS", ""); envVal != "" || !yamlLoaded {
		config.Postgres.MaxConns = getEnvAsInt32("POSTGRES_MAX_CONNS", getDefaultInt32(config.Postgres.MaxConns, 10))
	}
	if envVal := getEnv("POSTGRES_MIN_CONNS", ""); envVal != "" || !yamlLoaded {
		config.Postgres.MinConns = getEnvAsInt32("POSTGRES_MIN_CONNS", getDefaultInt32(config.Postgres.MinConns, 1))
	}
	if envVal := getEnv("POSTGRES_CONNECT_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Postgres.ConnectTimeout = getEnvAsDuration("POSTGRES_CONNECT_TIMEOUT", getDefaultDuration(config.Postgres.ConnectTimeout, 5*time.Second))
	}
	if envVal := getEnv("POSTGRES_HEALTH_CHECK_PERIOD", ""); envVal != "" || !yamlLoaded {
		config.Postgres.HealthCheckPeriod = getEnvAsDuration("POSTGRES_HEALTH_CHECK_PERIOD", getDefaultDuration(config.Postgres.HealthCheckPeriod, 30*time.Second))
	}

	// MOS config
	if envVal := getEnv("MOS_ID", ""); envVal != "" || !yamlLoaded {
		config.MOS.ID = getEnv("MOS_ID", getDefaultString(config.MOS.ID, "OpenMOS_Server"))
	}
	if envVal := getEnv("MOS_HEARTBEAT_INTERVAL", ""); envVal != "" || !yamlLoaded {
		config.MOS.HeartbeatInterval = getEnvAsDuration("MOS_HEARTBEAT_INTERVAL", getDefaultDuration(config.MOS.HeartbeatInterval, 30*time.Second))
	}
	if envVal := getEnv("MOS_CLIENT_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.MOS.ClientTimeout = getEnvAsDuration("MOS_CLIENT_TIMEOUT", getDefaultDuration(config.MOS.ClientTimeout, 2*time.Minute))
	}

	// Logging config
	if envVal := getEnv("LOG_LEVEL", ""); envVal != "" || !yamlLoaded {
		config.Logging.Level = getEnv("LOG_LEVEL", getDefaultString(config.Logging.Level, "info"))
	}

	// Sentry config
	if envVal := getEnv("SENTRY_DSN", ""); envVal != "" || !yamlLoaded {
		config.Sentry.DSN = getEnv("SENTRY_DSN", getDefaultString(config.Sentry.DSN, ""))
	}
	if envVal := getEnv("SENTRY_ENV", ""); envVal != "" || !yamlLoaded {
		config.Sentry.Environment = getEnv("SENTRY_ENV", getDefaultString(config.Sentry.Environment, config.App.Environment))
	}
	if envVal := getEnv("SENTRY_DEBUG", ""); envVal != "" || !yamlLoaded {
		config.Sentry.Debug = getEnvAsBool("SENTRY_DEBUG", getDefaultBool(config.Sentry.Debug, false))
	}
	if envVal := getEnv("SENTRY_ATTACH_STACKTRACE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.AttachStacktrace = getEnvAsBool("SENTRY_ATTACH_STACKTRACE", getDefaultBool(config.Sentry.AttachStacktrace, true))
	}
	if envVal := getEnv("SENTRY_SAMPLE_RATE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.SampleRate = getEnvAsFloat("SENTRY_SAMPLE_RATE", getDefaultFloat(config.Sentry.SampleRate, 1.0))
	}
	if envVal := getEnv("SENTRY_TRACES_SAMPLE_RATE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.TracesSampleRate = getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", getDefaultFloat(config.Sentry.TracesSampleRate, 0.2))
	}
}

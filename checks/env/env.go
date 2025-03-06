package env

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"otel-checker/checks/utils"
)

// EnvVar represents an environment variable configuration
type EnvVar struct {
	Name         string
	Required     bool
	DefaultValue string
	Validator    func(string) error
	Description  string
}

// Common environment variables used across the project
var (
	// OpenTelemetry common variables
	OtelServiceName = EnvVar{
		Name:        "OTEL_SERVICE_NAME",
		Required:    false,
		Description: "Service name for easier identification",
	}

	OtelExporterOTLPProtocol = EnvVar{
		Name:         "OTEL_EXPORTER_OTLP_PROTOCOL",
		Required:     true,
		DefaultValue: "http/protobuf",
		Validator: func(value string) error {
			if value != "http/protobuf" {
				return fmt.Errorf("must be set to 'http/protobuf'")
			}
			return nil
		},
		Description: "Protocol for OTLP exporter",
	}

	OtelExporterOTLPEndpoint = EnvVar{
		Name:     "OTEL_EXPORTER_OTLP_ENDPOINT",
		Required: true,
		Validator: func(value string) error {
			match, _ := regexp.MatchString("https:\\/\\/.+\\.grafana\\.net\\/otlp", value)
			if !match {
				if strings.Contains(value, "localhost") {
					return fmt.Errorf("endpoint is set to localhost. Update to a Grafana endpoint similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
				}
				return fmt.Errorf("must be set in the format similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
			}
			return nil
		},
		Description: "OTLP exporter endpoint",
	}

	OtelExporterOTLPHeaders = EnvVar{
		Name:     "OTEL_EXPORTER_OTLP_HEADERS",
		Required: true,
		Validator: func(value string) error {
			tokenStart := "Authorization=Basic "
			if !strings.Contains(value, tokenStart) {
				return fmt.Errorf("must contain '%s...'", tokenStart)
			}
			return nil
		},
		Description: "OTLP exporter headers",
	}

	// Metrics, Traces, and Logs exporters
	OtelMetricsExporter = EnvVar{
		Name:         "OTEL_METRICS_EXPORTER",
		Required:     false,
		DefaultValue: "otlp",
		Validator: func(value string) error {
			if value == "none" {
				return fmt.Errorf("cannot be 'none'. Change to 'otlp' or leave unset")
			}
			return nil
		},
		Description: "Metrics exporter configuration",
	}

	OtelTracesExporter = EnvVar{
		Name:         "OTEL_TRACES_EXPORTER",
		Required:     false,
		DefaultValue: "otlp",
		Validator: func(value string) error {
			if value == "none" {
				return fmt.Errorf("cannot be 'none'. Change to 'otlp' or leave unset")
			}
			return nil
		},
		Description: "Traces exporter configuration",
	}

	OtelLogsExporter = EnvVar{
		Name:         "OTEL_LOGS_EXPORTER",
		Required:     false,
		DefaultValue: "otlp",
		Validator: func(value string) error {
			if value == "none" {
				return fmt.Errorf("cannot be 'none'. Change to 'otlp' or leave unset")
			}
			return nil
		},
		Description: "Logs exporter configuration",
	}
)

// CheckEnvVar validates an environment variable against its configuration and reports the result
func CheckEnvVar(envVar EnvVar, reporter *utils.ComponentReporter) {
	value := os.Getenv(envVar.Name)

	if envVar.Required && value == "" {
		reporter.AddError(fmt.Sprintf("%s is required", envVar.Name))
		return
	}

	if value == "" && envVar.DefaultValue != "" {
		value = envVar.DefaultValue
	}

	if envVar.Validator != nil && value != "" {
		if err := envVar.Validator(value); err != nil {
			reporter.AddError(fmt.Sprintf("%s: %s", envVar.Name, err))
			return
		}
	}

	if !envVar.Required && value == "" {
		reporter.AddWarning(fmt.Sprintf("%s is not set", envVar.Name))
		return
	}

	reporter.AddSuccessfulCheck(fmt.Sprintf("%s is set correctly", envVar.Name))
}

// CheckEnvVars validates multiple environment variables and reports the results
func CheckEnvVars(reporter *utils.ComponentReporter, envVars ...EnvVar) {
	for _, envVar := range envVars {
		CheckEnvVar(envVar, reporter)
	}
}

// GetEnvVar returns the value of an environment variable with its default value if not set
func GetEnvVar(envVar EnvVar) string {
	value := os.Getenv(envVar.Name)
	if value == "" && envVar.DefaultValue != "" {
		return envVar.DefaultValue
	}
	return value
}

// IsEnvVarSet checks if an environment variable is set
func IsEnvVarSet(envVar EnvVar) bool {
	return os.Getenv(envVar.Name) != ""
}

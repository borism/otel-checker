package env

import (
	"fmt"
	"os"
	"regexp"
	"strings"
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

	// JavaScript specific variables
	NodeOptions = EnvVar{
		Name:         "NODE_OPTIONS",
		Required:     false,
		DefaultValue: "--require @opentelemetry/auto-instrumentations-node/register",
		Description:  "Node.js options for auto-instrumentation",
	}

	// .NET specific variables
	CoreCLREnableProfiling = EnvVar{
		Name:         "CORECLR_ENABLE_PROFILING",
		Required:     true,
		DefaultValue: "1",
		Description:  "Enable .NET profiling",
	}

	CoreCLRProfiler = EnvVar{
		Name:         "CORECLR_PROFILER",
		Required:     true,
		DefaultValue: "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
		Description:  ".NET profiler GUID",
	}

	CoreCLRProfilerPath = EnvVar{
		Name:        "CORECLR_PROFILER_PATH",
		Required:    true,
		Description: "Path to .NET profiler",
	}

	OtelDotNetAutoHome = EnvVar{
		Name:        "OTEL_DOTNET_AUTO_HOME",
		Required:    true,
		Description: "Home directory for .NET auto-instrumentation",
	}

	// Beyla specific variables
	BeylaServiceName = EnvVar{
		Name:        "BEYLA_SERVICE_NAME",
		Required:    false,
		Description: "Service name for Beyla",
	}

	BeylaOpenPort = EnvVar{
		Name:        "BEYLA_OPEN_PORT",
		Required:    true,
		Description: "Port for Beyla to listen on",
	}

	GrafanaCloudSubmit = EnvVar{
		Name:        "GRAFANA_CLOUD_SUBMIT",
		Required:    true,
		Description: "Types of telemetry to submit to Grafana Cloud",
	}

	GrafanaCloudInstanceID = EnvVar{
		Name:        "GRAFANA_CLOUD_INSTANCE_ID",
		Required:    true,
		Description: "Grafana Cloud instance ID",
	}

	GrafanaCloudAPIKey = EnvVar{
		Name:        "GRAFANA_CLOUD_API_KEY",
		Required:    true,
		Description: "Grafana Cloud API key",
	}
)

// CheckEnvVar validates an environment variable against its configuration
func CheckEnvVar(envVar EnvVar) (string, error) {
	value := os.Getenv(envVar.Name)

	if envVar.Required && value == "" {
		return "", fmt.Errorf("%s is required", envVar.Name)
	}

	if value == "" && envVar.DefaultValue != "" {
		value = envVar.DefaultValue
	}

	if envVar.Validator != nil && value != "" {
		if err := envVar.Validator(value); err != nil {
			return value, fmt.Errorf("%s: %s", envVar.Name, err)
		}
	}

	return value, nil
}

// CheckEnvVars validates multiple environment variables
func CheckEnvVars(envVars ...EnvVar) (map[string]string, []error) {
	values := make(map[string]string)
	var errors []error

	for _, envVar := range envVars {
		value, err := CheckEnvVar(envVar)
		if err != nil {
			errors = append(errors, err)
		} else {
			values[envVar.Name] = value
		}
	}

	return values, errors
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

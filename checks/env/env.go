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
	Name          string
	Required      bool
	Recommended   bool
	DefaultValue  string
	RequiredValue string
	Validator     func(value string, language string, reporter *utils.ComponentReporter)
	Description   string
	Message       string
}

// Common environment variables used across the project
var (
	// OpenTelemetry common variables
	OtelServiceName = EnvVar{
		Name:        "OTEL_SERVICE_NAME",
		Recommended: true,
		Message:     "It's recommended the environment variable OTEL_SERVICE_NAME to be set to your service name, for easier identification",
	}

	OtelExporterOTLPProtocol = EnvVar{
		Name:          "OTEL_EXPORTER_OTLP_PROTOCOL",
		RequiredValue: "http/protobuf",
		Description:   "Protocol for OTLP exporter",
		Message:       "OTEL_EXPORTER_OTLP_PROTOCOL must be set to 'http/protobuf'",
	}

	OtelExporterOTLPEndpoint = EnvVar{
		Name:     "OTEL_EXPORTER_OTLP_ENDPOINT",
		Required: true,
		Validator: func(value string, language string, reporter *utils.ComponentReporter) {
			match, _ := regexp.MatchString("https://.+\\.grafana\\.net/otlp", value)
			if match {
				reporter.AddSuccessfulCheck("OTEL_EXPORTER_OTLP_ENDPOINT set in the format similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
			} else {
				if strings.Contains(value, "localhost") {
					reporter.AddWarning("OTEL_EXPORTER_OTLP_ENDPOINT is set to localhost. Update to a Grafana endpoint similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp to be able to send telemetry to your Grafana Cloud instance")
				} else {
					reporter.AddError("OTEL_EXPORTER_OTLP_ENDPOINT is not set in the format similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
				}
			}
		},
		Description: "OTLP exporter endpoint",
	}

	OtelExporterOTLPHeaders = EnvVar{
		Name:     "OTEL_EXPORTER_OTLP_HEADERS",
		Required: true,
		Validator: func(value string, language string, reporter *utils.ComponentReporter) {
			tokenStart := "Authorization=Basic "
			if language == "python" {
				tokenStart = "Authorization=Basic%20"
			}
			if strings.Contains(value, tokenStart) {
				reporter.AddSuccessfulCheck("OTEL_EXPORTER_OTLP_HEADERS is set correctly")
			} else {
				reporter.AddError(fmt.Sprintf("OTEL_EXPORTER_OTLP_HEADERS is not set. Value should have '%s...'", tokenStart))
			}
		},
		Description: "OTLP exporter headers",
	}

	OtelMetricsExporter = exporterEnvVar("OTEL_METRICS_EXPORTER", "Metrics")
	OtelTracesExporter  = exporterEnvVar("OTEL_TRACES_EXPORTER", "Traces")
	OtelLogsExporter    = exporterEnvVar("OTEL_LOGS_EXPORTER", "Logs")
)

func exporterEnvVar(key string, name string) EnvVar {
	return EnvVar{
		Name:         key,
		Required:     false,
		DefaultValue: "otlp",
		Validator: func(value string, language string, reporter *utils.ComponentReporter) {
			if value == "none" {
				reporter.AddError(fmt.Sprintf("The value of %s cannot be 'none'. Change the value to 'otlp' or leave it unset", key))
			} else {
				if value == "" {
					reporter.AddSuccessfulCheck(fmt.Sprintf("%s is unset, with a default value of 'otlp'", key))
				} else {
					reporter.AddSuccessfulCheck(fmt.Sprintf("The value of %s is set to '%s' (default value)", key, value))
				}
			}
		},
		Description: name + " exporter configuration",
	}
}

// CheckEnvVar validates an environment variable against its configuration and reports the result
func CheckEnvVar(language string, envVar EnvVar, reporter *utils.ComponentReporter) {
	value := GetValue(envVar)
	if envVar.Validator != nil {
		envVar.Validator(value, language, reporter)
	} else {
		if envVar.RequiredValue != "" && !envVar.Recommended {
			envVar.Required = true
		}

		if envVar.Required && checkValue(envVar, value, reporter.AddError) {
			return
		}
		if envVar.Recommended && checkValue(envVar, value, reporter.AddWarning) {
			return
		}
		reporter.AddSuccessfulCheck(fmt.Sprintf("%s is set to '%s'", envVar.Name, value))
	}
}

func checkValue(e EnvVar, value string, report func(string)) bool {
	if e.RequiredValue != "" {
		if value != e.RequiredValue {
			if e.Message == "" {
				report(fmt.Sprintf("%s must be set to '%s'", e.Name, e.RequiredValue))
			} else {
				report(e.Message)
			}
			return true
		}
	} else {
		if value == "" {
			description := e.Message
			if description == "" {
				description = fmt.Sprintf("%s is not set", e.Name)
			}
			report(description)
			return true
		}
	}
	return false
}

// CheckEnvVars validates multiple environment variables and reports the results
func CheckEnvVars(reporter *utils.ComponentReporter, language string, envVars ...EnvVar) {
	for _, envVar := range envVars {
		CheckEnvVar(language, envVar, reporter)
	}
}

// GetValue returns the value of an environment variable with its default value if not set
func GetValue(envVar EnvVar) string {
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

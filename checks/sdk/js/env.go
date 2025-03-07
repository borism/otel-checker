package js

import (
	"otel-checker/checks/env"
	"otel-checker/checks/utils"
	"strings"
)

// JavaScript specific environment variables
var (
	NodeOptions = env.EnvVar{
		Name:          "NODE_OPTIONS",
		RequiredValue: "--require @opentelemetry/auto-instrumentations-node/register",
		Description:   "Node.js options for auto-instrumentation",
	}

	NodeResourceDetectors = env.EnvVar{
		Name: "OTEL_NODE_RESOURCE_DETECTORS",
		Validator: func(value string, language string, reporter *utils.ComponentReporter) {
			if value == "" ||
				!strings.Contains(value, "env") ||
				!strings.Contains(value, "host") ||
				!strings.Contains(value, "os") ||
				!strings.Contains(value, "serviceinstance") {
				reporter.AddWarning("It's recommended the environment variable OTEL_NODE_RESOURCE_DETECTORS to be set to at least `env,host,os,serviceinstance`")
			} else {
				reporter.AddSuccessfulCheck("OTEL_NODE_RESOURCE_DETECTORS has recommended values")
			}
		},
		Description: "at least `env,host,os,serviceinstance`",
	}
)

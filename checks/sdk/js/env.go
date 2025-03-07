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
		Recommended:   true,
		RequiredValue: "--require @opentelemetry/auto-instrumentations-node/register",
		Description:   `NODE_OPTIONS not set. You can set it by running 'export NODE_OPTIONS="--require @opentelemetry/auto-instrumentations-node/register"' or add the same '--require ...' when starting your application`,
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

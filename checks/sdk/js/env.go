package js

import (
	"fmt"
	"otel-checker/checks/env"
	"strings"
)

// JavaScript specific environment variables
var (
	NodeOptions = env.EnvVar{
		Name:         "NODE_OPTIONS",
		Required:     false,
		DefaultValue: "--require @opentelemetry/auto-instrumentations-node/register",
		Description:  "Node.js options for auto-instrumentation",
	}

	NodeResourceDetectors = env.EnvVar{
		Name:     "OTEL_NODE_RESOURCE_DETECTORS",
		Required: false,
		Validator: func(value string) error {
			requiredDetectors := []string{"env", "host", "os", "serviceinstance"}
			for _, detector := range requiredDetectors {
				if !strings.Contains(value, detector) {
					return fmt.Errorf("It's recommended the environment variable OTEL_NODE_RESOURCE_DETECTORS to be set to at least `env,host,os,serviceinstance`")
				}
			}
			return nil
		},
		Description: "at least `env,host,os,serviceinstance`",
	}
)

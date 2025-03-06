package js

import (
	"fmt"
	"os"
	"os/exec"
	"otel-checker/checks/env"
	"otel-checker/checks/utils"
	"strconv"
	"strings"
)

func CheckJSSetup(reporter *utils.ComponentReporter, commands utils.Commands) {
	checkEnvVars(reporter)
	checkNodeVersion(reporter)
	if commands.ManualInstrumentation {
		checkJSCodeBasedInstrumentation(reporter, commands.PackageJsonPath, commands.InstrumentationFile)
	} else {
		checkJSAutoInstrumentation(reporter, commands.PackageJsonPath)
	}
	checkSupportedLibraries(reporter, commands)
}

func checkEnvVars(reporter *utils.ComponentReporter) {
	// Check Node resource detectors
	nodeDetectors := env.EnvVar{
		Name:     "OTEL_NODE_RESOURCE_DETECTORS",
		Required: false,
		Validator: func(value string) error {
			requiredDetectors := []string{"env", "host", "os", "serviceinstance"}
			for _, detector := range requiredDetectors {
				if !strings.Contains(value, detector) {
					return fmt.Errorf("must include '%s'", detector)
				}
			}
			return nil
		},
		Description: "Node resource detectors",
	}

	value, err := env.CheckEnvVar(nodeDetectors)
	if err != nil {
		reporter.AddWarning(fmt.Sprintf("It's recommended the environment variable OTEL_NODE_RESOURCE_DETECTORS to be set to at least `env,host,os,serviceinstance`: %s", err))
	} else {
		reporter.AddSuccessfulCheck("OTEL_NODE_RESOURCE_DETECTORS has recommended values")
	}
}

func checkNodeVersion(reporter *utils.ComponentReporter) {
	cmd := exec.Command("node", "-v")
	stdout, err := cmd.Output()

	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check minimum node version: %s", err))
		return
	}
	versionInfo := strings.Split(string(stdout), ".")
	v, err := strconv.Atoi(versionInfo[0][1:])
	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check minimum node version: %s", err))
		return
	}
	if v >= 16 {
		reporter.AddSuccessfulCheck("Using node version equal or greater than minimum recommended")
	} else {
		reporter.AddError("Not using recommended node version. Update your node to at least version 16")
	}
}

func checkJSAutoInstrumentation(
	reporter *utils.ComponentReporter,
	packageJsonPath string,
) {
	// Check NODE_OPTIONS
	value, err := env.CheckEnvVar(env.NodeOptions)
	if err != nil {
		reporter.AddWarning(fmt.Sprintf("NODE_OPTIONS not set correctly: %s", err))
	} else {
		reporter.AddSuccessfulCheck("NODE_OPTIONS set correctly")
	}

	// Dependencies for auto instrumentation on package.json
	filePath := packageJsonPath + "package.json"
	dat, err := os.ReadFile(filePath)
	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check file %s: %s", filePath, err))
		return
	}

	content := string(dat)
	requiredDeps := []struct {
		name    string
		message string
	}{
		{`"@opentelemetry/auto-instrumentations-node"`, "Dependency @opentelemetry/auto-instrumentations-node missing on package.json. Install the dependency with `npm install @opentelemetry/auto-instrumentations-node`"},
		{`"@opentelemetry/api"`, "Dependency @opentelemetry/api missing on package.json. Install the dependency with `npm install @opentelemetry/auto-instrumentations-node`"},
	}

	for _, dep := range requiredDeps {
		if strings.Contains(content, dep.name) {
			reporter.AddSuccessfulCheck(fmt.Sprintf("Dependency %s added on package.json", strings.Trim(dep.name, `"`)))
		} else {
			reporter.AddError(dep.message)
		}
	}
}

func checkJSCodeBasedInstrumentation(
	reporter *utils.ComponentReporter,
	packageJsonPath string,
	instrumentationFile string,
) {
	// Check NODE_OPTIONS is not set for auto-instrumentation
	if env.IsEnvVarSet(env.NodeOptions) {
		reporter.AddError(`The flag "-manual-instrumentation" was set, but the value of NODE_OPTIONS is set to require auto-instrumentation. Run "unset NODE_OPTIONS" to remove the requirement that can cause a conflict with manual instrumentations`)
	}

	// Check dependencies in package.json
	filePath := packageJsonPath + "package.json"
	packageJsonContent, err := os.ReadFile(filePath)
	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check file %s: %s", filePath, err))
		return
	}

	content := string(packageJsonContent)
	requiredDeps := []struct {
		name    string
		message string
	}{
		{`"@opentelemetry/api"`, "Dependency @opentelemetry/api missing on package.json"},
	}

	for _, dep := range requiredDeps {
		if strings.Contains(content, dep.name) {
			reporter.AddSuccessfulCheck(fmt.Sprintf("Dependency %s added on package.json", strings.Trim(dep.name, `"`)))
		} else {
			reporter.AddError(dep.message)
		}
	}

	// Check for unsupported dependencies
	if strings.Contains(content, `"@opentelemetry/exporter-trace-otlp-proto"`) {
		reporter.AddError(`Dependency @opentelemetry/exporter-trace-otlp-proto added on package.json, which is not supported by Grafana. Switch the dependency to "@opentelemetry/exporter-trace-otlp-http" instead`)
	}

	// Check Exporter
	instrumentationFileContent, err := os.ReadFile(instrumentationFile)
	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check file %s: %s", instrumentationFile, err))
	} else {
		if strings.Contains(string(instrumentationFileContent), "ConsoleSpanExporter") {
			reporter.AddWarning("Instrumentation file is using ConsoleSpanExporter. This exporter is useful during debugging, but replace with OTLPTraceExporter to send to Grafana Cloud")
		}
		if strings.Contains(string(instrumentationFileContent), "ConsoleMetricExporter") {
			reporter.AddWarning("Instrumentation file is using ConsoleMetricExporter. This exporter is useful during debugging, but replace with OTLPMetricExporter to send to Grafana Cloud")
		}
	}
}

func checkSupportedLibraries(reporter *utils.ComponentReporter, commands utils.Commands) {
	supported, err := supportedLibraries()
	if err != nil {
		reporter.AddError(fmt.Sprintf("Error reading supported libraries: %v", err))
		return
	}

	deps := readDependencies(reporter)
	if len(deps) == 0 {
		return
	}

	for _, dep := range deps {
		links := findSupportedLibraries(dep, supported)
		if len(links) > 0 {
			reporter.AddSuccessfulCheck(
				fmt.Sprintf("Found supported library: %s:%s at %s",
					dep.Name, dep.Version, strings.Join(links, ", ")))
		} else if commands.Debug {
			reporter.AddWarning(fmt.Sprintf("Found unsupported library: %s:%s", dep.Name, dep.Version))
		}
	}
}

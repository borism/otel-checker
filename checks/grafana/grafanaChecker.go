package grafana

import (
	"fmt"
	"net/http"
	"otel-checker/checks/beyla"
	"otel-checker/checks/env"
	"otel-checker/checks/utils"
	"slices"
	"strings"
)

func CheckGrafanaSetup(reporter utils.Reporter, grafanaReporter *utils.ComponentReporter, language string, components []string) {
	checkEnvVarsGrafana(reporter, grafanaReporter, language, components)
	checkAuth(grafanaReporter)
}

func checkEnvVarsGrafana(reporter utils.Reporter, grafana *utils.ComponentReporter, language string, components []string) {
	// Check common OpenTelemetry variables
	commonVars := []env.EnvVar{
		env.OtelServiceName,
		env.OtelExporterOTLPProtocol,
		env.OtelExporterOTLPEndpoint,
		env.OtelExporterOTLPHeaders,
		env.OtelMetricsExporter,
		env.OtelTracesExporter,
		env.OtelLogsExporter,
	}

	env.CheckEnvVars(grafana, commonVars...)

	// Check Beyla specific variables if component is enabled
	if slices.Contains(components, "beyla") {
		beyla := reporter.Component("Beyla")
		beylaVars := []env.EnvVar{
			beyla.ServiceName,
			beyla.OpenPort,
			beyla.GrafanaCloudSubmit,
			beyla.GrafanaCloudInstanceID,
			beyla.GrafanaCloudAPIKey,
		}

		env.CheckEnvVars(beyla, beylaVars...)
	}
}

func checkAuth(reporter *utils.ComponentReporter) {
	endpoint := env.GetEnvVar(env.OtelExporterOTLPEndpoint)
	if strings.Contains(endpoint, "localhost") {
		reporter.AddWarning("Credentials not checked, since OTEL_EXPORTER_OTLP_ENDPOINT is using localhost")
		return
	}

	headers := env.GetEnvVar(env.OtelExporterOTLPHeaders)
	if endpoint == "" || headers == "" {
		reporter.AddWarning("Credentials not checked, since both environment variables OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_EXPORTER_OTLP_HEADERS need to be set for this check")
		return
	}

	// Test credentials
	testEndpoint := endpoint + "/v1/metrics"
	req, err := http.NewRequest("POST", testEndpoint, nil)
	if err != nil {
		reporter.AddError(fmt.Sprintf("Error while testing credentials of OTEL_EXPORTER_OTLP_ENDPOINT: %s", err))
		return
	}

	// Extract auth value from headers
	authValue := ""
	for _, h := range strings.SplitN(headers, ",", -1) {
		key, value, _ := strings.Cut(h, "=")
		if key == "Authorization" {
			authValue = value
		}
	}
	req.Header.Set("Authorization", authValue)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		reporter.AddError(fmt.Sprintf("Error while testing credentials of OTEL_EXPORTER_OTLP_ENDPOINT: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		reporter.AddError(fmt.Sprintf("Error while testing credentials of OTEL_EXPORTER_OTLP_ENDPOINT: %s", resp.Status))
	} else {
		reporter.AddSuccessfulCheck("Credentials for OTEL_EXPORTER_OTLP_ENDPOINT are correct")
	}
}

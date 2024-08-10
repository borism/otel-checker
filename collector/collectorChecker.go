package collector

import (
	"fmt"
	"os"
	"otel-checker/utils"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

func CheckCollectorSetup(messages *map[string][]string, language string, configPath string) {
	checkCollectorConfig(messages, configPath)
}

type configFile struct {
	Receivers struct {
		Otlp struct {
			Protocols struct {
				Grpc *string `yaml:"grpc"`
				Http *string `yaml:"http"`
			} `yaml:"protocols"`
		} `yaml:"otlp"`
	} `yaml:"receivers"`
	Exporters struct {
		Otlphttp struct {
			Endpoint string                 `yaml:"endpoint"`
			Auth     map[string]interface{} `yaml:"auth"`
		} `yaml:"otlphttp"`
	} `yaml:"exporters"`
	Service struct {
		Pipelines struct {
			Traces struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"traces"`
			Logs struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"logs"`
			Metrics struct {
				Receivers  []string `yaml:"receivers"`
				Processors []string `yaml:"processors"`
				Exporters  []string `yaml:"exporters"`
			} `yaml:"metrics"`
		} `yaml:"pipelines"`
	} `yaml:"service"`
}

func checkCollectorConfig(messages *map[string][]string, configPath string) {
	filePath := configPath + "config.yaml"
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		utils.AddError(messages, "Collector", fmt.Sprintf("Could not check file %s: %s", filePath, err))
	} else {
		var c configFile
		err = yaml.Unmarshal([]byte(yamlFile), &c)
		if err != nil {
			utils.AddError(messages, "Collector", fmt.Sprintf("Could not parse file %s: %s", filePath, err))
			return
		}

		if c.Receivers.Otlp.Protocols.Http == nil {
			utils.AddWarning(messages, "Collector", "The value of receivers > otlp > protocols > http is nil. Make sure the key exists on your config.yaml")
		}

		match, _ := regexp.MatchString("https:\\/\\/.+\\.grafana\\.net\\/otlp", c.Exporters.Otlphttp.Endpoint)
		if match {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of exporter > otlphttp > endpoint on config.yaml set in the format similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
		} else {
			if strings.Contains(c.Exporters.Otlphttp.Endpoint, "localhost") {
				utils.AddWarning(messages, "Collector", "Value of exporter > otlphttp > endpoint on config.yaml is set to localhost. Update to a Grafana endpoint similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp to be able to send telemetry to your Grafana Cloud instance")
			} else {
				utils.AddError(messages, "Collector", "Value of exporter > otlphttp > endpoint on config.yaml is not set in the format similar to https://otlp-gateway-prod-us-east-0.grafana.net/otlp")
			}
		}

		// Traces
		if slices.Contains(c.Service.Pipelines.Traces.Exporters, "otlphttp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > traces > exporters on config.yaml contains otlphttp")
		} else {
			utils.AddWarning(messages, "Collector", "Value of service > pipelines > traces > exporters on config.yaml does not contain otlphttp")
		}
		if slices.Contains(c.Service.Pipelines.Traces.Receivers, "otlp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > traces > receivers on config.yaml contains otlp")
		} else {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > traces > receivers on config.yaml does not contain otlp")
		}

		// Logs
		if slices.Contains(c.Service.Pipelines.Logs.Exporters, "otlphttp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > logs > exporters on config.yaml contains otlphttp")
		} else {
			utils.AddWarning(messages, "Collector", "Value of service > pipelines > logs > exporters on config.yaml does not contain otlphttp")
		}
		if slices.Contains(c.Service.Pipelines.Logs.Receivers, "otlp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > logs > receivers on config.yaml contains otlp")
		} else {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > logs > receivers on config.yaml does not contain otlp")
		}

		// Metrics
		if slices.Contains(c.Service.Pipelines.Metrics.Exporters, "otlphttp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > metrics > exporters on config.yaml contains otlphttp")
		} else {
			utils.AddWarning(messages, "Collector", "Value of service > pipelines > metrics > exporters on config.yaml does not contain otlphttp")
		}
		if slices.Contains(c.Service.Pipelines.Metrics.Receivers, "otlp") {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > metrics > receivers on config.yaml contains otlp")
		} else {
			utils.AddSuccessfulCheck(messages, "Collector", "Value of service > pipelines > metrics > receivers on config.yaml does not contain otlp")
		}
	}
}

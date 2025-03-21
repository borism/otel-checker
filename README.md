# OTel Me If It's Right

Checker if the implementation of OpenTelemetry instrumentation is correct.

## Usage

Requirement: Golang

## Installation
1. Install the `otel-checker` binary
```
go install github.com/grafana/otel-checker@latest
```
2. You can confirm it was installed with:
```
❯ ls $GOPATH/bin
otel-checker
```

## Flags

The available flags:
```
❯ otel-checker -h
Usage of otel-checker:
  -manual-instrumentation
    	Provide if your application is using manual instrumentation (auto instrumentation as default)
  -collector-config-path string
    	Path to collector's config.yaml file. Required if using Collector and the config file is not in the same location as the otel-checker is being executed from. E.g. "-collector-config-path=src/inst/"
  -components string
    	Instrumentation components to test, separated by ',' (required). Possible values: sdk, collector, beyla, alloy, grafana-cloud
  -debug
        Output debug information
  -instrumentation-file string
    	Name (including path) to instrumentation file. Required if using manual-instrumentation. E.g."-instrumentation-file=src/inst/instrumentation.js"
  -language string
    	Language used for instrumentation (required). Possible values: dotnet, go, java, js, python, ruby, php
  -package-json-path string
    	Path to package.json file. Required if instrumentation is in JavaScript and the file is not in the same location as the otel-checker is being executed from. E.g. "-package-json-path=src/inst/"
  -web-server
        Set if you would like the results served in a web server in addition to console output
```

## Checks

### Common Environment Variables
         
These checks are automatically performed for all languages and components.

- Best practices for setting common environment variables:
  - Service name
  - Exporter protocol 

- Resource attributes checks:
  - Validates the presence of recommended OpenTelemetry resource attributes
  - Checks for the following attributes:
    - `service.name` (via `OTEL_SERVICE_NAME` or in `OTEL_RESOURCE_ATTRIBUTES`)
    - `service.namespace` (e.g., "shop")
    - `deployment.environment.name` (e.g., "production")
    - `service.instance.id` (e.g., "checkout-123")
    - `service.version` (e.g., "1.2")
  - For missing attributes, provides specific recommendations with example values
  - Follows the [OpenTelemetry specification](https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/) for precedence (e.g., `OTEL_SERVICE_NAME` takes precedence over `service.name` in `OTEL_RESOURCE_ATTRIBUTES`)
  - Example warning: `Set OTEL_RESOURCE_ATTRIBUTES="service.namespace=shop": An optional namespace for service.name`

### Grafana Cloud

Use the `-components=grafana-cloud` flag to check the following:

- Endpoints
- Authentication

### SDK

#### JavaScript

Use `-components=sdk -language=js` flag to check the following:

- Node version
- Required dependencies on package.json
- Required environment variables
- Resource detectors
- Dependencies compatible with Grafana Cloud
- Usage of Console Exporter
- Prints which libraries are supported based on the `package.json` in the current directory.

#### Python

Use `-components=sdk -language=python` flag to check the following:

- Prints which libraries are supported:
  - The used libraries are discovered from `requirements.txt` in the current directory.

#### .NET

Use `-components=sdk -language=dotnet` flag to check the following:

- .NET version
- Available instrumentation for .NET libraries and dependencies
- Auto-instrumentation environment variables

**Only .NET 8.0 and higher are supported**

#### Java
   
Use `-components=sdk -language=java` flag to check the following:

- Java version
- Prints which libraries (as discovered from a locally running maven or gradle) are supported:
  - With `-manual-instrumentation`, the libraries for manual instrumentation are printed.
  - Without `-manual-instrumentation`, it will print the libraries supported by the [Java Agent](https://github.com/open-telemetry/opentelemetry-java-instrumentation/).
  - A maven or gradle wrapper will be used if found in the current directory or a parent directory.

#### Go

Use `-components=sdk -language=go` flag to check the following:

- Prints which libraries are supported for manual instrumentation 
  based on the `go.mod` in the current directory.

#### Ruby

Use `-components=sdk -language=ruby` flag to check the following:

- Ruby version
- Bundler installation
- `Gemfile` and `Gemfile.lock` exist
- Required dependencies installed
- Optional auto-instrumentation dependencies installed

#### PHP

Use `-components=sdk -language=php` flag to check the following:

- PHP version
- Composer installation
- `composer.json` and `composer.lock` exist
- Required dependencies in `composer.lock`
- Some auto-instrumentation dependencies installed

### Collector

Use `-components=collector` flag to check the following:

- Config receivers and exporters

### Beyla

Use `-components=beyla` flag to check the following:

- Environment variables

### Alloy
TBD

## Examples

Application with auto-instrumentation
![auto instrumentation exemple](./assets/auto.png)

Application with custom instrumentation using SDKs and Collector
![sdk and collector example](./assets/sdk.png)

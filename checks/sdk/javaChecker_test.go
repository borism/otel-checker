package sdk

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFindSupportedLibrary(t *testing.T) {
	modules, err := supportedLibraries()
	require.NoError(t, err)
	assert.Equal(t, []string{
		"https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/logback/logback-mdc-1.0", "https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/logback/logback-appender-1.0",
	}, findSupportedLibraries(JavaLibrary{
		Group:    "ch.qos.logback",
		Artifact: "logback-classic",
		Version:  "1.5.16",
	}, modules))
}

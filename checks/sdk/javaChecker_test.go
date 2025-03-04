package sdk

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFindSupportedLibrary(t *testing.T) {
	modules, err := supportedLibraries()
	require.NoError(t, err)
	assert.True(t, findSupportedLibrary(JavaLibrary{
		Group:    "ch.qos.logback",
		Artifact: "logback-classic",
		Version:  "1.5.16",
	}, modules))
}

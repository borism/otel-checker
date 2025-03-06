package dotnet

import (
	"otel-checker/checks/env"
)

// .NET specific environment variables
var (
	CoreCLREnableProfiling = env.EnvVar{
		Name:         "CORECLR_ENABLE_PROFILING",
		Required:     true,
		DefaultValue: "1",
		Description:  "Enable .NET profiling",
	}

	CoreCLRProfiler = env.EnvVar{
		Name:         "CORECLR_PROFILER",
		Required:     true,
		DefaultValue: "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
		Description:  ".NET profiler GUID",
	}

	CoreCLRProfilerPath = env.EnvVar{
		Name:        "CORECLR_PROFILER_PATH",
		Required:    true,
		Description: "Path to .NET profiler",
	}

	OtelDotNetAutoHome = env.EnvVar{
		Name:        "OTEL_DOTNET_AUTO_HOME",
		Required:    true,
		Description: "Home directory for .NET auto-instrumentation",
	}
)

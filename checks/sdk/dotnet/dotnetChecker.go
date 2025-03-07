package dotnet

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"otel-checker/checks/env"
	"otel-checker/checks/utils"
)

const minDotNetVersion = 8

func CheckDotNetSetup(reporter *utils.ComponentReporter, commands utils.Commands) {
	checkDotNetVersion(reporter)

	project, err := checkProject(reporter)

	if err != nil {
		return
	}

	reportDotNetSupportedInstrumentations(reporter, project.SDK)

	if commands.ManualInstrumentation {
		checkDotNetCodeBasedInstrumentation(reporter)
	} else {
		checkDotNetAutoInstrumentation(reporter, commands)
	}
}

func checkDotNetVersion(reporter *utils.ComponentReporter) {
	cmd := exec.Command("dotnet", "--version")
	stdout, err := cmd.Output()

	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not check .NET version: %s", err))
		return
	}

	version := strings.TrimSpace(string(stdout))
	versionParts := strings.Split(version, ".")
	if len(versionParts) == 0 {
		reporter.AddError("Could not parse .NET version: version string is empty")
		return
	}
	majorVersion := versionParts[0]
	v, err := strconv.Atoi(majorVersion)

	if err != nil {
		reporter.AddError(fmt.Sprintf("Could not parse .NET version: %s", err))
		return
	}

	if v >= minDotNetVersion {
		reporter.AddSuccessfulCheck(fmt.Sprintf("Using .NET version equal or greater than minimum recommended (%d.0)", minDotNetVersion))
	} else {
		reporter.AddError(fmt.Sprintf("Not using recommended .NET version. Update your .NET SDK to at least version %d.0", minDotNetVersion))
	}
}

func checkDotNetAutoInstrumentation(reporter *utils.ComponentReporter, commands utils.Commands) {
	env.CheckEnvVars(reporter, commands.Language,
		env.EnvVar{
			Name:          "CORECLR_ENABLE_PROFILING",
			RequiredValue: "1",
		},
		env.EnvVar{
			Name:          "CORECLR_PROFILER",
			RequiredValue: "{918728DD-259F-4A6A-AC2B-B85E1B658318}",
		},
		env.EnvVar{
			Name:     "CORECLR_PROFILER_PATH",
			Required: true,
		},
		env.EnvVar{
			Name:     "OTEL_DOTNET_AUTO_HOME",
			Required: true,
		})
}

func checkDotNetCodeBasedInstrumentation(reporter *utils.ComponentReporter) {
}

func readDotNetDependenciesFromCli() (*NuGetPackageList, error) {
	cmd := exec.Command("dotnet", "list", "package", "--format", "json", "--include-transitive")
	stdout, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to run dotnet list package: %w", err)
	}

	var deps NuGetPackageList
	if err := json.Unmarshal(stdout, &deps); err != nil {
		return nil, fmt.Errorf("failed to parse dependencies JSON: %w", err)
	}

	return &deps, nil
}

func findProject() (string, error) {
	var csprojFiles []string

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != "." {
			return filepath.SkipDir
		}
		if filepath.Ext(d.Name()) == ".csproj" {
			csprojFiles = append(csprojFiles, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to search for .csproj files: %w", err)
	}

	switch len(csprojFiles) {
	case 0:
		return "", fmt.Errorf("no .csproj files found in current directory")
	case 1:
		return csprojFiles[0], nil
	default:
		return "", fmt.Errorf("multiple .csproj files found: %s", strings.Join(csprojFiles, ", "))
	}
}

func checkProject(reporter *utils.ComponentReporter) (*CSharpProject, error) {
	project, err := findProject()

	if err != nil {
		reporter.AddError(fmt.Sprintf("Failed to find project file: %s", err))
		return nil, err
	}

	reporter.AddSuccessfulCheck(fmt.Sprintf("Found project file: %s", project))
	content, err := os.ReadFile(project)

	if err != nil {
		reporter.AddError(fmt.Sprintf("Failed to read project file: %s", err))
		return nil, err
	}

	var csProj CSharpProject
	if err := xml.Unmarshal(content, &csProj); err != nil {
		reporter.AddError(fmt.Sprintf("Failed to parse project file: %s", err))
		return nil, err
	}

	return &csProj, nil
}

func reportDotNetSupportedInstrumentations(reporter *utils.ComponentReporter, sdk string) {
	deps, err := readDotNetDependenciesFromCli()

	if err != nil {
		reporter.AddError(fmt.Sprintf("Failed to read dependencies: %s", err))
		return
	}

	instr := ReadAvailableInstrumentations()

	implicit, err := ImplicitPackagesForSdk(sdk)

	if err != nil {
		reporter.AddError(fmt.Sprintf("Unrecognized SDK: %s", sdk))
		return
	}

	if len(implicit) == 0 {
		reporter.AddWarning(fmt.Sprintf("No implicit packages found for SDK: %s", sdk))
	} else {
		for _, pkg := range implicit {
			lib, ok := instr[pkg]

			if !ok {
				continue
			}

			reporter.AddSuccessfulCheck(fmt.Sprintf("Found supported instrumentation for %s: %s", pkg, lib))
		}
	}

	for _, project := range deps.Projects {
		for _, framework := range project.Frameworks {
			for _, pkg := range framework.TopLevelPackages {
				lib, ok := instr[pkg.ID]

				if !ok {
					continue
				}

				reporter.AddSuccessfulCheck(fmt.Sprintf("Found supported instrumentation for %s: %s", pkg.ID, lib))
			}

		}
	}
	if len(deps.Projects) == 0 {
		reporter.AddError("No dependencies found in project")
		return
	}
}

package sdk

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
	"os/exec"
	"otel-checker/checks/utils"
	"slices"
	"strings"
)

var gradleFiles = []string{
	"build.gradle",
	"build.gradle.kts",
}

type JavaLibrary struct {
	Group    string        `json:"groupId"`
	Artifact string        `json:"artifactId"`
	Version  string        `json:"version"`
	Children []JavaLibrary `json:"children"`
}

type SupportedModules map[string]SupportedModule

type SupportedModule struct {
	Instrumentations []Instrumentation `yaml:"instrumentations"`
}

type Instrumentation struct {
	Name           string                `yaml:"name"`
	SrcPath        string                `yaml:"srcPath"`
	Types          []InstrumentationType `yaml:"types"`
	TargetVersions []string              `yaml:"target_versions"`
}

type InstrumentationType string

const (
	Javaagent InstrumentationType = "JAVAAGENT"
	Library   InstrumentationType = "LIBRARY"
)

func CheckJavaSetup(messages *map[string][]string, autoInstrumentation bool, debug bool) {
	checkJavaVersion(messages)
	if autoInstrumentation {
		checkJavaAutoInstrumentation(messages, debug)
	} else {
		checkJavaCodeBasedInstrumentation(messages)
	}
}

func checkJavaVersion(messages *map[string][]string) {
	// check for java 8
}

func checkJavaAutoInstrumentation(messages *map[string][]string, debug bool) {
	supported, err := supportedLibraries()
	if err != nil {
		utils.AddError(messages, "SDK", fmt.Sprintf("Error reading supported libraries: %v", err))
	}

	if utils.FileExists("pom.xml") {
		checkMaven(messages, supported, debug)
	}
	for _, file := range gradleFiles {
		if utils.FileExists(file) {
			checkGradle(file)
		}
	}
}

func checkJavaCodeBasedInstrumentation(messages *map[string][]string) {}

func checkMaven(messages *map[string][]string, supported SupportedModules, debug bool) {
	println("Reading Maven dependencies")

	tool := "mvn"
	if utils.FileExists("mvnw") {
		// todo windows
		tool = "./mvnw"
	}
	// call maven to get dependencies
	cmd := exec.Command(tool, "dependency:tree", "-Dscope=runtime", "-DoutputType=json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.AddError(messages, "SDK", fmt.Sprintf("Error running maven dependency:tree:\n%v\n%s", err, output))
	}
	out := string(output)
	deps := parseMavenDeps(out)
	if len(deps) == 0 {
		utils.AddWarning(messages, "SDK", "No maven dependencies found")
	}
	outputSupportedLibraries(deps, supported, messages, debug)
}

func outputSupportedLibraries(deps []JavaLibrary, supported SupportedModules, messages *map[string][]string, debug bool) {
	for _, dep := range deps {
		links := findSupportedLibraries(dep, supported)
		if len(links) > 0 {
			utils.AddSuccessfulCheck(messages, "SDK",
				fmt.Sprintf("Found supported library: %s:%s:%s at %s",
					dep.Group, dep.Artifact, dep.Version, strings.Join(links, ", ")))
		} else if debug {
			utils.AddWarning(messages, "SDK", fmt.Sprintf("Found unsupported library: %s:%s:%s", dep.Group, dep.Artifact, dep.Version))
		}
		outputSupportedLibraries(dep.Children, supported, messages, false)
	}
}

func findSupportedLibraries(library JavaLibrary, supported SupportedModules) []string {
	var links []string
	for moduleName, module := range supported {
		for _, instrumentation := range module.Instrumentations {
			// todo check type (agent or library)
			for _, version := range instrumentation.TargetVersions {
				// e.g. com.amazonaws:aws-lambda-java-core:[1.0.0,)
				split := strings.Split(version, ":")
				if len(split) != 3 {
					panic(fmt.Sprintf("invalid version range: %s", version))
				}
				versionRange, err := ParseVersionRange(split[2])
				if err != nil {
					panic(fmt.Sprintf("error parsing version range in module %s: %v", moduleName, err))
				}
				if library.Group == split[0] && library.Artifact == split[1] {
					v := FixVersion(library.Version)
					if semver.IsValid(v) {
						// ignore invalid versions from applications
						if versionRange.matches(v) {
							l := fmt.Sprintf("https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/%s", instrumentation.SrcPath)
							if !slices.Contains(links, l) {
								links = append(links, l)
							}
						}
					}
				}
			}
		}
	}
	return links
}

func parseMavenDeps(out string) []JavaLibrary {
	c := ""
	isJson := false
	for l := range strings.Lines(out) {
		if strings.Contains(l, "[INFO] {") {
			isJson = true
		}
		if isJson {
			c += strings.TrimPrefix(l, "[INFO] ")
		}
		if strings.Contains(l, "[INFO] }") {
			isJson = false
		}
	}

	var deps JavaLibrary
	err := json.Unmarshal([]byte(c), &deps)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil
	}

	return []JavaLibrary{deps}
}

func checkGradle(file string) {
	println(fmt.Sprintf("Reading Gradle dependencies from %s", file))
}

// https://github.com/open-telemetry/opentelemetry-java-instrumentation/pull/13449
// see https://cloud-native.slack.com/archives/C014L2KCTE3/p1741003980069869
// CNCF slack channel #otel-java
//
//go:embed supported-java-libraries.yaml
var supportedModules []byte

func supportedLibraries() (SupportedModules, error) {
	modules := SupportedModules{}
	err := yaml.Unmarshal(supportedModules, &modules)
	if err != nil {
		return nil, err
	}
	delete(modules, "internal")
	return modules, nil
}

package sdk

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"otel-checker/checks/utils"
	"strings"
)

var gradleFiles = []string{
	"build.gradle",
	"build.gradle.kts",
}

type JavaLibrary struct {
	Group    string         `json:"groupId"`
	Artifact string         `json:"artifactId"`
	Version  semver.Version `json:"version"`
	Children []JavaLibrary  `json:"children"`
}

type SupportedModules map[string]SupportedModule

type SupportedModule struct {
	Instrumentations []Instrumentation `json:"instrumentations"`
}

type Instrumentation struct {
	Name           string                `json:"name"`
	SrcPath        string                `json:"srcPath"`
	Types          []InstrumentationType `json:"types"`
	TargetVersions []string              `json:"target_versions"`
}

type InstrumentationType string

const (
	Javaagent InstrumentationType = "JAVAAGENT"
	Library   InstrumentationType = "LIBRARY"
)

func CheckJavaSetup(
	messages *map[string][]string,
	autoInstrumentation bool,
) {
	checkJavaVersion(messages)
	if autoInstrumentation {
		checkJavaAutoInstrumentation(messages)
	} else {
		checkJavaCodeBasedInstrumentation(messages)
	}
}

func checkJavaVersion(messages *map[string][]string) {
	// check for java 8
}

func checkJavaAutoInstrumentation(messages *map[string][]string) {
	supported, err := supportedLibraries()
	if err != nil {
		utils.AddError(messages, "SDK", fmt.Sprintf("Error reading supported libraries: %v", err))
	}

	if utils.FileExists("pom.xml") {
		checkMaven(messages, supported)
	}
	for _, file := range gradleFiles {
		if utils.FileExists(file) {
			checkGradle()
		}
	}
}

func checkJavaCodeBasedInstrumentation(messages *map[string][]string) {}

func checkMaven(messages *map[string][]string, supported SupportedModules) {
	println("Checking Maven")

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
	println(out)
	deps := parseMavenDeps(out)
	if len(deps) == 0 {
		utils.AddWarning(messages, "SDK", "No maven dependencies found")
	}
	outputSupportedLibraries(deps, supported, messages)
}

func outputSupportedLibraries(deps []JavaLibrary, supported SupportedModules, messages *map[string][]string) {
	for _, dep := range deps {
		library := findSupportedLibrary(dep, supported)
		if library {
			utils.AddSuccessfulCheck(messages, "SDK", fmt.Sprintf("Found supported library: %s:%s:%s", dep.Group, dep.Artifact, dep.Version))
		}
	}
}

func findSupportedLibrary(library JavaLibrary, supported SupportedModules) bool {
	for _, module := range supported {
		for _, instrumentation := range module.Instrumentations {
			// todo check type (agent or library)
			for _, version := range instrumentation.TargetVersions {
				// e.g. com.amazonaws:aws-lambda-java-core:[1.0.0,)
				split := strings.Split(version, ":")
				if len(split) != 3 {
					panic(fmt.Sprintf("invalid version range: %s", version))
				}
				if library.Group == split[0] && library.Artifact == split[1] {
					versionRange, err := ParseVersionRange(split[2])
					if err != nil {
						panic(err)
					}
					if versionRange.matches(library.Version) {
						return true
					}
				}
			}
		}
	}
	return false
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

	dependencies := []JavaLibrary{deps}
	printDeps(dependencies)

	return dependencies
}

func printDeps(deps []JavaLibrary) {
	for _, d := range deps {
		fmt.Printf("%s:%s:%s\n", d.Group, d.Artifact, d.Version)
		printDeps(d.Children)
	}
}

func checkGradle() {
	println("Checking Gradle")
}

// see https://cloud-native.slack.com/archives/C014L2KCTE3/p1741003980069869
// CNCF slack channel #otel-java
//
//go:embed checks/sdk/supported-java-libraries.yaml
var supportedModules string

func supportedLibraries() (SupportedModules, error) {
	file, err := os.ReadFile("checks/sdk/supported-java-libraries.yaml")
	if err != nil {
		return nil, err
	}
	modules := SupportedModules{}
	err = yaml.Unmarshal(file, &modules)
	if err != nil {
		return nil, err
	}
	return modules, nil
}

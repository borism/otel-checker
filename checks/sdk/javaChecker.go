package sdk

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
	"os/exec"
	"otel-checker/checks/utils"
	"path/filepath"
	"slices"
	"strconv"
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

func (l *JavaLibrary) String() string {
	return fmt.Sprintf("%s:%s:%s", l.Group, l.Artifact, l.Version)
}

type SupportedModules map[string]SupportedModule

type SupportedModule struct {
	Instrumentations []Instrumentation `yaml:"instrumentations"`
}

type Instrumentation struct {
	Name           string                           `yaml:"name"`
	SrcPath        string                           `yaml:"srcPath"`
	TargetVersions map[InstrumentationType][]string `yaml:"target_versions"`
}

type InstrumentationType string

const (
	Javaagent InstrumentationType = "JAVAAGENT"
	Library   InstrumentationType = "LIBRARY"
)

func CheckJavaSetup(reporter *utils.ComponentReporter, autoInstrumentation bool, debug bool) {
	checkJavaVersion(reporter)
	if autoInstrumentation {
		checkJavaAutoInstrumentation(reporter, debug)
	} else {
		checkJavaCodeBasedInstrumentation(reporter, debug)
	}
}

func checkJavaVersion(reporter *utils.ComponentReporter) {
	out := RunCommand(reporter, exec.Command("java", "-version"))
	if out != "" {
		//openjdk version "21.0.2" 2024-01-16 LTS
		line := strings.Split(out, "\n")[0]
		field := strings.Split(line, " ")[2]
		version := strings.Trim(field, "\"")
		major, err := strconv.Atoi(strings.Split(version, ".")[0])
		if err != nil {
			reporter.AddError(fmt.Sprintf("Error parsing Java version %s: %v", out, err))
		}
		if strings.HasPrefix(version, "1.8") {
			major = 8
		}
		if major < 8 {
			reporter.AddError(fmt.Sprintf("Java version %s is not supported. Please use Java 8 or higher", version))
		} else {
			reporter.AddSuccessfulCheck(fmt.Sprintf("Java version %s is supported", version))
		}
	}
}

func checkJavaAutoInstrumentation(reporter *utils.ComponentReporter, debug bool) {
	reportSupportedInstrumentations(reporter, debug, Javaagent)
}

func checkJavaCodeBasedInstrumentation(reporter *utils.ComponentReporter, debug bool) {
	reportSupportedInstrumentations(reporter, debug, Library)
}

func reportSupportedInstrumentations(reporter *utils.ComponentReporter, debug bool, instrumentationType InstrumentationType) {
	supported, err := supportedLibraries()
	if err != nil {
		reporter.AddError(fmt.Sprintf("Error reading supported libraries: %v", err))
	}

	deps := readDependencies(reporter)
	outputSupportedLibraries(deps, supported, reporter, debug, instrumentationType)
}

func readDependencies(reporter *utils.ComponentReporter) []JavaLibrary {
	if utils.FileExists("pom.xml") {
		return checkMaven(reporter)
	}
	for _, file := range gradleFiles {
		if utils.FileExists(file) {
			return checkGradle(file, reporter)
		}
	}
	return nil
}

func checkMaven(reporter *utils.ComponentReporter) []JavaLibrary {
	println("Reading Maven dependencies")

	out := RunCommand(reporter, exec.Command(searchWrapper("mvn", "mvnw"),
		"dependency:tree", "-Dscope=runtime", "-DoutputType=json"))
	if out == "" {
		return []JavaLibrary{}
	}
	deps := parseMavenDeps(out)
	if len(deps) == 0 {
		reporter.AddWarning("No Maven dependencies found")
	}
	return deps
}

func searchWrapper(base string, wrapper string) string {
	tool := getWrapper(wrapper, []string{"."})
	if tool == "" {
		return base
	}
	return tool
}

func getWrapper(wrapper string, level []string) string {
	if len(level) > 10 {
		return ""
	}
	p := filepath.Join(filepath.Join(level...), wrapper)
	if utils.FileExists(p) {
		// the . is needed to run the wrapper in the current directory
		return fmt.Sprintf(".%c%s", filepath.Separator, p)
	}
	return getWrapper(wrapper, append(level, ".."))
}

func outputSupportedLibraries(
	deps []JavaLibrary, supported SupportedModules, reporter *utils.ComponentReporter,
	debug bool, instrumentationType InstrumentationType) {
	for _, dep := range deps {
		links := findSupportedLibraries(dep, supported, instrumentationType)
		if len(links) > 0 {
			reporter.AddSuccessfulCheck(
				fmt.Sprintf("Found supported library: %s:%s:%s at %s",
					dep.Group, dep.Artifact, dep.Version, strings.Join(links, ", ")))
		} else if debug {
			reporter.AddWarning(fmt.Sprintf("Found unsupported library: %s:%s:%s", dep.Group, dep.Artifact, dep.Version))
		}
		outputSupportedLibraries(dep.Children, supported, reporter, false, instrumentationType)
	}
}

func findSupportedLibraries(library JavaLibrary, supported SupportedModules, instrumentationType InstrumentationType) []string {
	var links []string
	for moduleName, module := range supported {
		for _, instrumentation := range module.Instrumentations {
			for _, version := range instrumentation.TargetVersions[instrumentationType] {
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

func checkGradle(file string, reporter *utils.ComponentReporter) []JavaLibrary {
	println("Reading Gradle dependencies")

	out := RunCommand(reporter, exec.Command(searchWrapper("gradle", "gradlew"),
		fmt.Sprintf("--build-file=%s", file), "dependencies", "--configuration=runtimeClasspath"))
	if out == "" {
		return []JavaLibrary{}
	}
	deps := parseGradleDeps(out)
	if len(deps) == 0 {
		reporter.AddWarning("No Gradle dependencies found")
	}
	return deps
}

// https://github.com/open-telemetry/opentelemetry-java-instrumentation/pull/13449
// see https://cloud-native.slack.com/archives/C014L2KCTE3/p1741003980069869
// CNCF slack channel #otel-java
//
//go:embed instrumentation-list.yaml
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

func parseGradleDeps(out string) []JavaLibrary {
	lines := strings.Split(out, "\n")
	var deps []JavaLibrary
	for _, l := range lines {
		if strings.Contains(l, "---") {
			index := strings.Index(l, "---")
			dep := strings.TrimSpace(l[index+4:])
			split := strings.Split(dep, ":")
			if len(split) == 3 {
				d := JavaLibrary{
					Group:    split[0],
					Artifact: split[1],
					Version:  split[2],
				}
				s := d.String()
				if !slices.ContainsFunc(
					deps,
					func(l JavaLibrary) bool {
						return l.String() == s
					}) {
					deps = append(deps, d)
				}
			}
		}
	}

	return deps
}

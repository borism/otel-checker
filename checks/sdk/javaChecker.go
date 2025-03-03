package sdk

import (
	"fmt"
	"os/exec"
	"otel-checker/checks/utils"
	"strings"
)

var gradleFiles = []string{
	"build.gradle",
	"build.gradle.kts",
}

type javaDependency struct {
	group    string
	artifact string
	version  string
}

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
	if utils.FileExists("pom.xml") {
		checkMaven(messages)
	}
	for _, file := range gradleFiles {
		if utils.FileExists(file) {
			checkGradle()
		}
	}
}

func checkJavaCodeBasedInstrumentation(messages *map[string][]string) {}

func checkMaven(messages *map[string][]string) {
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
}

func parseMavenDeps(out string) []javaDependency {
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
	println(c)
	json.
}

func checkGradle() {
	println("Checking Gradle")
}

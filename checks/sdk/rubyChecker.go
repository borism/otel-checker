package sdk

import (
	"os/exec"
	utils "otel-checker/checks/utils"
	"strings"

	"golang.org/x/mod/semver"
)

func CheckRubySetup(
	messages *map[string][]string,
	autoInstrumentation bool,
) {
	checkRubyVersion(messages)
	if autoInstrumentation {
		checkRubyAutoInstrumentation(messages)
	} else {
		checkRubyCodeBasedInstrumentation(messages)
	}
}

func checkRubyVersion(messages *map[string][]string) {
	hasCRuby := checkCRubyVersion(messages)
	hasJRuby := checkJRubyVersion(messages)
	hasTruffleRuby := checkJRubyVersion(messages)

	if hasCRuby || hasJRuby || hasTruffleRuby {
		utils.AddSuccessfulCheck(messages, "SDK", "Ruby setup successful")
	} else {
		utils.AddError(messages, "SDK", "No Ruby found, install CRuby >= 3.0, JRuby >= 9.3.2.0, or TruffleRuby >= 22.1")
	}
}

func checkCRubyVersion(messages *map[string][]string) bool {
	cmd := exec.Command("ruby", "-v")
	stdout, err := cmd.Output()

	if err != nil {
		return false
	}

	if strings.Contains(string(stdout), "ruby 3") {
		utils.AddSuccessfulCheck(messages, "SDK", "Using CRuby version equal or greater than minimum recommended")
		return true
	} else {
		utils.AddError(messages, "SDK", "Not using recommended CRuby version, update to CRuby >= 3.0.")
		return false
	}
}

func checkJRubyVersion(messages *map[string][]string) bool {
	cmd := exec.Command("jruby", "--version")
	stdout, err := cmd.Output()

	if err != nil {
		return false
	}

	version := strings.Fields(string(stdout))[2]

	if semver.Compare(version, "9.3.2.0") >= 0 {
		utils.AddSuccessfulCheck(messages, "SDK", "Using JRuby version equal or greater than minimum recommended")
		return true
	} else {
		utils.AddError(messages, "SDK", "Not using recommended JRuby version, update to JRuby >= 9.3.2.0.")
		return false
	}
}

// TruffleRuby check is not supported yet
func checkTruffleRubyVersion(messages *map[string][]string) bool {
	return false
}

func checkRubyAutoInstrumentation(messages *map[string][]string) {}

func checkRubyCodeBasedInstrumentation(messages *map[string][]string) {}

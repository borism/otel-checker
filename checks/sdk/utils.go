package sdk

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"strings"
)

type VersionRange struct {
	lower          *semver.Version
	upper          *semver.Version
	upperInclusive bool
}

// ParseVersionRange parses versions of the form "[5.0,)", "[2.0,4.0)".
func ParseVersionRange(version string) (VersionRange, error) {
	split := strings.Split(version, ",")
	if len(split) != 2 {
		return VersionRange{}, fmt.Errorf("version has more than one comma: %s", version)
	}
	if !strings.HasPrefix(split[0], "[") {
		return VersionRange{}, fmt.Errorf("version does not start with '[': %s", version)
	}
	upperInclusive := false
	if strings.HasSuffix(version, "]") {
		upperInclusive = true
	} else if strings.HasSuffix(version, ")") {
		upperInclusive = false
	} else {
		return VersionRange{}, fmt.Errorf("version does not end with ']' or ')': %s", version)
	}

	var err error
	var lower, upper *semver.Version
	l := strings.TrimLeft(split[0], "[")
	if l != "" {
		lower, err = semver.NewVersion(l)
		if err != nil {
			return VersionRange{}, err
		}
	}
	u := strings.TrimRight(split[1], ")]")
	if u != "" {
		upper, err = semver.NewVersion(u)
		if err != nil {
			return VersionRange{}, err
		}
	}
	return VersionRange{
		lower,
		upper,
		upperInclusive,
	}, nil
}

func (r *VersionRange) matches(v semver.Version) bool {
	if r.lower != nil && v.LessThan(r.lower) {
		return false
	}
	if r.upper == nil {
		return true
	}
	if r.upperInclusive {
		return !v.GreaterThan(r.upper)
	}
	return v.LessThan(r.upper)
}

func CheckSDKSetup(
	messages *map[string][]string,
	language string,
	autoInstrumentation bool,
	packageJsonPath string,
	instrumentationFile string,
) {
	switch language {
	case "dotnet":
		CheckDotNetSetup(messages, autoInstrumentation)
	case "go":
		CheckGoSetup(messages, autoInstrumentation)
	case "java":
		CheckJavaSetup(messages, autoInstrumentation)
	case "js":
		CheckJSSetup(messages, autoInstrumentation, packageJsonPath, instrumentationFile)
	case "python":
		CheckPythonSetup(messages, autoInstrumentation)
	}
}

package sdk

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"strings"
)

type VersionRange struct {
	lower          *semver.Version
	lowerInclusive bool
	upper          *semver.Version
	upperInclusive bool
}

// ParseVersionRange parses versions of the form "[5.0,)", "[2.0,4.0)".
func ParseVersionRange(version string) (VersionRange, error) {
	split := strings.Split(version, ",")
	if len(split) != 2 {
		return VersionRange{}, fmt.Errorf("version has more than one comma: %s", version)
	}
	lowerInclusive := false
	if strings.HasPrefix(version, "[") {
		lowerInclusive = true
	} else if strings.HasPrefix(version, "(") {
		lowerInclusive = false
	} else {
		return VersionRange{}, fmt.Errorf("version does not start with '[' or '(': %s", version)
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
	l := strings.TrimLeft(split[0], "[(")
	if l != "" {
		lower, err = semver.NewVersion(l)
		if err != nil {
			return VersionRange{}, fmt.Errorf("cannot parse lower version %s: %w", l, err)
		}
	}
	u := strings.TrimRight(split[1], ")]")
	if u != "" {
		upper, err = semver.NewVersion(u)
		if err != nil {
			return VersionRange{}, fmt.Errorf("cannot parse upper version %s: %w", u, err)
		}
	}
	return VersionRange{
		lower,
		lowerInclusive,
		upper,
		upperInclusive,
	}, nil
}

func (r *VersionRange) matches(v semver.Version) bool {
	if r.lower != nil {
		if r.lowerInclusive {
			if v.LessThan(r.lower) {
				return false
			}
		} else {
			if v.LessThanEqual(r.lower) {
				return false
			}
		}
		return false
	}
	if r.upper == nil {
		return true
	}
	if r.upperInclusive {
		return v.GreaterThanEqual(r.upper)
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

package sdk

import (
	"fmt"
	"golang.org/x/mod/semver"
	"strings"
)

type VersionRange struct {
	lower          string
	lowerInclusive bool
	upper          string
	upperInclusive bool
}

var suffixes = []string{
	".Final",
	".RELEASE",
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

	l := FixVersion(strings.TrimLeft(split[0], "[("))
	if l != "" {
		if !semver.IsValid(l) {
			return VersionRange{}, fmt.Errorf("invalid semver: '%s'", l)
		}
	}
	u := FixVersion(strings.TrimRight(split[1], ")]"))
	if u != "" {
		if !semver.IsValid(u) {
			return VersionRange{}, fmt.Errorf("invalid semver: '%s'", u)
		}
	}
	return VersionRange{
		l,
		lowerInclusive,
		u,
		upperInclusive,
	}, nil
}

func FixVersion(version string) string {
	if version == "" {
		return version
	}
	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	split := strings.Split(version, ".")
	if len(split) == 1 {
		version = version + ".0"
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(version, suffix) {
			return strings.TrimSuffix(version, suffix)
		}
	}
	return version
}

func (r *VersionRange) matches(v string) bool {
	if r.lower != "" {
		compare := semver.Compare(v, r.lower)
		if r.lowerInclusive {
			if compare < 0 {
				return false
			}
		} else {
			if compare <= 0 {
				return false
			}
		}
		return false
	}
	if r.upper == "" {
		return true
	}
	compare := semver.Compare(v, r.upper)
	if r.upperInclusive {
		return compare >= 0
	}
	return compare < 0
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

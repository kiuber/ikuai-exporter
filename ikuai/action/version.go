package action

import (
	"strconv"
	"strings"
)

func ParseMajorVersion(version string) int {
	version = strings.TrimSpace(version)
	if version == "" {
		return 0
	}

	parts := strings.SplitN(version, ".", 2)
	if len(parts) == 0 {
		return 0
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	return major
}

func IsV4(version string) bool {
	return ParseMajorVersion(version) >= 4
}

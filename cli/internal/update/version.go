package update

import (
	"fmt"
	"strconv"
	"strings"
)

type semanticVersion struct {
	major      int
	minor      int
	patch      int
	prerelease []string
}

func parseVersion(value string) (semanticVersion, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "v"))
	value = strings.SplitN(value, "+", 2)[0]
	parts := strings.SplitN(value, "-", 2)
	core := strings.Split(parts[0], ".")
	if len(core) != 3 {
		return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
	}

	numbers := make([]int, 3)
	for index, part := range core {
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 {
			return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
		}
		numbers[index] = number
	}

	parsed := semanticVersion{major: numbers[0], minor: numbers[1], patch: numbers[2]}
	if len(parts) == 2 {
		if parts[1] == "" {
			return semanticVersion{}, fmt.Errorf("invalid semantic version %q", value)
		}
		parsed.prerelease = strings.Split(parts[1], ".")
	}
	return parsed, nil
}

func isNewer(latest string, current string) (bool, error) {
	latestVersion, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	currentVersion, err := parseVersion(current)
	if err != nil {
		return false, err
	}

	latestCore := []int{latestVersion.major, latestVersion.minor, latestVersion.patch}
	currentCore := []int{currentVersion.major, currentVersion.minor, currentVersion.patch}
	for index := range latestCore {
		if latestCore[index] != currentCore[index] {
			return latestCore[index] > currentCore[index], nil
		}
	}

	return comparePrerelease(latestVersion.prerelease, currentVersion.prerelease) > 0, nil
}

func comparePrerelease(left []string, right []string) int {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}
	if len(left) == 0 {
		return 1
	}
	if len(right) == 0 {
		return -1
	}

	length := len(left)
	if len(right) < length {
		length = len(right)
	}
	for index := 0; index < length; index++ {
		leftNumber, leftNumeric := numericIdentifier(left[index])
		rightNumber, rightNumeric := numericIdentifier(right[index])
		switch {
		case leftNumeric && rightNumeric && leftNumber != rightNumber:
			if leftNumber > rightNumber {
				return 1
			}
			return -1
		case leftNumeric != rightNumeric:
			if leftNumeric {
				return -1
			}
			return 1
		case left[index] != right[index]:
			if left[index] > right[index] {
				return 1
			}
			return -1
		}
	}

	switch {
	case len(left) > len(right):
		return 1
	case len(left) < len(right):
		return -1
	default:
		return 0
	}
}

func numericIdentifier(value string) (int, bool) {
	if value == "" {
		return 0, false
	}
	number, err := strconv.Atoi(value)
	return number, err == nil
}

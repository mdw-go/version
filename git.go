package version

import (
	"errors"
	"strconv"
	"strings"
)

func ParseGitDescribe(raw string) (number Number, err error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "fatal: No names found, cannot describe anything.") {
		number.Prefix = "v"
		number.Dirty = true
		return number, nil
	}
	if strings.HasPrefix(raw, "v") {
		number.Prefix = "v"
		raw = strings.TrimPrefix(raw, "v")
	}

	fields := strings.Split(raw, "-")
	number.Dirty = len(fields) > 1

	parts := strings.Split(fields[0], ".")
	if len(parts) < 3 {
		return Number{Prefix: "v"}, errors.New("three version fields are required: major.minor.patch")
	}
	number.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return number, err
	}
	number.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return number, err
	}
	number.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return number, err
	}
	return number, nil
}

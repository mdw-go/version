package version

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Number struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
	Dev    string
}

func Parse(raw string) (number Number, err error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "v") {
		number.Prefix = "v"
		raw = strings.TrimPrefix(raw, "v")
	}

	fields := strings.Split(raw, "-dev")
	parts := strings.Split(fields[0], ".")
	if len(parts) < 3 {
		return Number{}, errors.New("three version fields are required: major.minor.patch")
	}
	number.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return Number{}, err
	}
	number.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return Number{}, err
	}
	number.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return Number{}, err
	}
	if len(fields) > 1 {
		number.Dev = "-dev" + fields[1]
	}
	return number, nil
}

func (this Number) String() (result string) {
	return fmt.Sprintf("%s%d.%d.%d%s", this.Prefix, this.Major, this.Minor, this.Patch, this.Dev)
}
func (this Number) IncrementMajor() Number {
	return Number{Prefix: this.Prefix, Major: this.Major + 1}
}
func (this Number) IncrementMinor() Number {
	return Number{Prefix: this.Prefix, Major: this.Major, Minor: this.Minor + 1}
}
func (this Number) IncrementPatch() Number {
	return Number{Prefix: this.Prefix, Major: this.Major, Minor: this.Minor, Patch: this.Patch + 1}
}
func (this Number) IncrementDev(name string) Number {
	return Number{
		Prefix: this.Prefix,
		Major:  this.Major,
		Minor:  this.Minor,
		Patch:  this.Patch,
		Dev:    "-dev-" + name,
	}
}

func Sort(versions []Number) {
	sort.Slice(versions, func(i, j int) bool {
		I, J := versions[i], versions[j]
		if I.Major == J.Major {
			if I.Minor == J.Minor {
				if I.Patch == J.Patch {
					return I.Dev < J.Dev
				}
				return I.Patch < J.Patch
			}
			return I.Minor < J.Minor
		}
		return I.Major < J.Major
	})
}

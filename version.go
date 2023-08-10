package version

import (
	"fmt"
	"strings"
)

type Number struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
	Dirty  bool
}

func (this Number) String() string {
	return fmt.Sprintf("%s%d.%d.%d", this.Prefix, this.Major, this.Minor, this.Patch)
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
func (this Number) Increment(how string) Number {
	switch strings.ToLower(how) {
	case "major":
		return this.IncrementMajor()
	case "minor":
		return this.IncrementMinor()
	default:
		return this.IncrementPatch()
	}
}

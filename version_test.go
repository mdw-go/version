package version

import (
	"testing"

	"github.com/mdw-go/testing/should"
)

func TestParse(t *testing.T) {
	var parsed Number
	var err error

	parsed, err = Parse("")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("1")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("1.2")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("a.0.1")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("1.a.0")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("0.1.a")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = Parse("1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Patch: 0, Dev: ""})

	parsed, err = Parse("v1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v", Major: 1, Minor: 2, Patch: 0, Dev: ""})

	parsed, err = Parse("1.2.0-dev4")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Patch: 0, Dev: "-dev4"})

	parsed, err = Parse("1.2.0-devA")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Patch: 0, Dev: "-devA"})
}

var (
	basic    = parse("1.2.3")
	prefixed = parse("v1.2.3")
	dev      = parse("1.2.3-dev4")
)

func TestIncrement(t *testing.T) {
	should.So(t, basic.IncrementMajor(), should.Equal, parse("2.0.0"))
	should.So(t, basic.IncrementMinor(), should.Equal, parse("1.3.0"))
	should.So(t, basic.IncrementPatch(), should.Equal, parse("1.2.4"))
	should.So(t, basic.IncrementDev("name"), should.Equal, parse("1.2.3-dev-name"))
}
func TestString(t *testing.T) {
	should.So(t, basic.String(), should.Equal, "1.2.3")
	should.So(t, prefixed.String(), should.Equal, "v1.2.3")
	should.So(t, dev.String(), should.Equal, "1.2.3-dev4")
}
func parse(raw string) Number {
	parsed, _ := Parse(raw)
	return parsed
}

func TestSort(t *testing.T) {
	versions := []Number{
		parse("3.2.1"),
		parse("3.0.0"),
		parse("3.2.1-dev0"),
		parse("0.0.0-dev100"),
		parse("3.2.0"),
		parse("2.4.12"),
	}
	Sort(versions)
	should.So(t, versions, should.Equal, []Number{
		parse("0.0.0-dev100"),
		parse("2.4.12"),
		parse("3.0.0"),
		parse("3.2.0"),
		parse("3.2.1"),
		parse("3.2.1-dev0"),
	})
}

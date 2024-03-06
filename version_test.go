package version

import (
	"testing"

	"github.com/mdwhatcott/testing/should"
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
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Patch: 0, Dev: -1})

	parsed, err = Parse("v1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v", Major: 1, Minor: 2, Patch: 0, Dev: -1})

	parsed, err = Parse("1.2.0-dev4")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Patch: 0, Dev: 4})

	parsed, err = Parse("1.2.0-devA")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})
}

var (
	basic    = parse("1.2.3")
	prefixed = parse("v1.2.3")
	dev      = parse("1.2.3-dev4")
)

func TestIncrement(t *testing.T) {
	should.So(t, basic.Increment("mAjOr"), should.Equal, parse("2.0.0"))
	should.So(t, basic.Increment("MiNoR"), should.Equal, parse("1.3.0"))
	should.So(t, basic.Increment("PATCH"), should.Equal, parse("1.2.4"))
	should.So(t, basic.Increment("DEV"), should.Equal, parse("1.2.3-dev0"))
	should.So(t, basic.Increment(""), should.Equal, basic)

	should.So(t, prefixed.Increment("mAjOr"), should.Equal, parse("v2.0.0"))
	should.So(t, prefixed.Increment("MiNoR"), should.Equal, parse("v1.3.0"))
	should.So(t, prefixed.Increment("PATCH"), should.Equal, parse("v1.2.4"))
	should.So(t, prefixed.Increment("DEV"), should.Equal, parse("v1.2.3-dev0"))
	should.So(t, prefixed.Increment(""), should.Equal, prefixed)

	should.So(t, dev.Increment("mAjOr"), should.Equal, parse("2.0.0"))
	should.So(t, dev.Increment("MiNoR"), should.Equal, parse("1.3.0"))
	should.So(t, dev.Increment("PATCH"), should.Equal, parse("1.2.4"))
	should.So(t, dev.Increment("DEV"), should.Equal, parse("1.2.3-dev5"))
	should.So(t, dev.Increment(""), should.Equal, dev)
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
		parse("3.2.1-dev0"),
		parse("0.0.0-dev100"),
		parse("2.4.12"),
	}
	Sort(versions)
	should.So(t, versions, should.Equal, []Number{
		parse("0.0.0-dev100"),
		parse("2.4.12"),
		parse("3.2.1"),
		parse("3.2.1-dev0"),
	})
}

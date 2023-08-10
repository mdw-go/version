package version

import (
	"testing"

	"github.com/mdwhatcott/testing/should"
)

var (
	dirty = Number{Major: 1, Minor: 2, Patch: 3, Dirty: true}
	clean = Number{Major: 1, Minor: 2, Patch: 3, Dirty: false}

	dirtyV = Number{Prefix: "v", Major: 1, Minor: 2, Patch: 3, Dirty: true}
	cleanV = Number{Prefix: "v", Major: 1, Minor: 2, Patch: 3, Dirty: false}
)

func TestVersionIncrementationFixture(t *testing.T) {
	should.So(t, dirty.Increment("mAjOr"), should.Equal, parse("2.0.0"))
	should.So(t, dirty.Increment("MiNoR"), should.Equal, parse("1.3.0"))
	should.So(t, dirty.Increment("PATCH"), should.Equal, parse("1.2.4"))
	should.So(t, dirty.Increment(""), should.Equal, parse("1.2.4"))

	should.So(t, dirtyV.Increment("mAjOr"), should.Equal, parse("v2.0.0"))
	should.So(t, dirtyV.Increment("MiNoR"), should.Equal, parse("v1.3.0"))
	should.So(t, dirtyV.Increment("PATCH"), should.Equal, parse("v1.2.4"))
	should.So(t, dirtyV.Increment(""), should.Equal, parse("v1.2.4"))
}
func TestVersionString(t *testing.T) {
	should.So(t, clean.String(), should.Equal, "1.2.3")
	should.So(t, dirty.String(), should.Equal, "1.2.3")

	should.So(t, cleanV.String(), should.Equal, "v1.2.3")
	should.So(t, dirtyV.String(), should.Equal, "v1.2.3")
}
func parse(raw string) Number {
	parsed, _ := ParseGitDescribe(raw)
	return parsed
}

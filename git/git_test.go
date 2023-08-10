package git

import (
	"testing"

	"github.com/mdwhatcott/testing/should"
	"github.com/mdwhatcott/version"
)

func TestParseGitVersion(t *testing.T) {
	var parsed version.Number
	var err error

	parsed, err = parseGitDescribe("")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v"})

	parsed, err = parseGitDescribe("1")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v"})

	parsed, err = parseGitDescribe("1.2")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v"})

	parsed, err = parseGitDescribe("fatal: No names found, cannot describe anything.")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v", Dirty: true})

	parsed, err = parseGitDescribe("1.a.0")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Major: 1})

	parsed, err = parseGitDescribe("1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Major: 1, Minor: 2})

	parsed, err = parseGitDescribe("v1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v", Major: 1, Minor: 2})

	parsed, err = parseGitDescribe("1.2.0-4-g3201d7a")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Major: 1, Minor: 2, Dirty: true})

	parsed, err = parseGitDescribe("v1.2.0-4-g3201d7a")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, version.Number{Prefix: "v", Major: 1, Minor: 2, Dirty: true})
}

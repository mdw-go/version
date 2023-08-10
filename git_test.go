package version

import (
	"testing"

	"github.com/mdwhatcott/testing/should"
)

func TestParseGitVersion(t *testing.T) {
	var parsed Number
	var err error

	parsed, err = ParseGitDescribe("")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v"})

	parsed, err = ParseGitDescribe("1")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v"})

	parsed, err = ParseGitDescribe("1.2")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v"})

	parsed, err = ParseGitDescribe("fatal: No names found, cannot describe anything.")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v", Dirty: true})

	parsed, err = ParseGitDescribe("a.0.1")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{})

	parsed, err = ParseGitDescribe("1.a.0")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1})

	parsed, err = ParseGitDescribe("0.1.a")
	should.So(t, err, should.NOT.BeNil)
	should.So(t, parsed, should.Equal, Number{Minor: 1})

	parsed, err = ParseGitDescribe("1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2})

	parsed, err = ParseGitDescribe("v1.2.0\n")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v", Major: 1, Minor: 2})

	parsed, err = ParseGitDescribe("1.2.0-4-g3201d7a")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Major: 1, Minor: 2, Dirty: true})

	parsed, err = ParseGitDescribe("v1.2.0-4-g3201d7a")
	should.So(t, err, should.BeNil)
	should.So(t, parsed, should.Equal, Number{Prefix: "v", Major: 1, Minor: 2, Dirty: true})
}

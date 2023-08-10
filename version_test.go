package version

import (
	"testing"

	"github.com/mdwhatcott/testing/should"
)

func TestVersionIncrementationFixture(t *testing.T) {
	should.So(t, New(1, 2, 3, true).IncrementMajor(), should.Equal, New(2, 0, 0, false))
	should.So(t, New(1, 2, 3, true).IncrementMinor(), should.Equal, New(1, 3, 0, false))
	should.So(t, New(1, 2, 3, true).IncrementPatch(), should.Equal, New(1, 2, 4, false))

	should.So(t, New(1, 2, 3, true).Increment("mAjOr"), should.Equal, New(2, 0, 0, false))
	should.So(t, New(1, 2, 3, true).Increment("MiNoR"), should.Equal, New(1, 3, 0, false))
	should.So(t, New(1, 2, 3, true).Increment("PATCH"), should.Equal, New(1, 2, 4, false))
	should.So(t, New(1, 2, 3, true).Increment(""), should.Equal, New(1, 2, 4, false))
}

func TestVersionString(t *testing.T) {
	should.So(t, New(1, 2, 3, false).String(), should.Equal, "1.2.3")
	should.So(t, New(1, 2, 3, true).String(), should.Equal, "1.2.3")
}

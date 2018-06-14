package cmd

import (
	"strings"
	"testing"
)

func Test_wellKnownEvents_trimmed(t *testing.T) {
	assertTrimmed := func(t *testing.T, subject string) {
		if subject != strings.TrimSpace(subject) {
			t.Logf("%q has leading or trailing whitespace", subject)
			t.Fail()
		}
	}

	for k, v := range wellKnownEvents {
		assertTrimmed(t, k)
		assertTrimmed(t, v)
	}
}

package flagx

import (
	"flag"
	"testing"
)

func TestFlagx(t *testing.T) {
	NewString("name", "bar", "baz", Required(), Env("BOO"))
	ParseFlagSet(flag.CommandLine, []string{"-name", "foo"})

	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name == "name" {
			t.Log(f.Usage)
			if f.Usage != "baz (env: BOO) (required) (default: bar)" {
				t.Fail()
			}
		}
	})
}

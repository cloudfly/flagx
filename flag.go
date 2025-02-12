package flagx

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	envPrefix = flag.String("env.prefix", "", "Name prefix of environment variables that interact with flags.")
	flagTypes = map[string]any{}
)

// NewBool creates a new bool flag.
func NewBool(name string, value bool, usage string) *bool {
	b := flag.Bool(name, value, usage+envHelp(name))
	flagTypes[name] = b
	return b
}

// NewString creates a new string flag.
func NewString(name string, value string, usage string) *string {
	s := flag.String(name, value, usage+envHelp(name))
	flagTypes[name] = s
	return s
}

// NewInt creates a new int flag.
func NewInt(name string, value int, usage string) *int {
	i := flag.Int(name, value, usage+envHelp(name))
	flagTypes[name] = i
	return i
}

// NewInt64 creates a new int64 flag.
func NewInt64(name string, value int64, usage string) *int64 {
	i64 := flag.Int64(name, value, usage+envHelp(name))
	flagTypes[name] = i64
	return i64
}

// NewFloat creates a new float64 flag.
func NewFloat(name string, value float64, usage string) *float64 {
	f := flag.Float64(name, value, usage+envHelp(name))
	flagTypes[name] = f
	return f
}

// WriteFlags writes all the explicitly set flags to w.
func WriteFlags(w io.Writer) {
	Visit(func(name, value string) {
		fmt.Fprintf(w, "-%s=%q\n", name, value)
	})
}

// Lookup a flag by name. the second return value is the real flag pointer which is returned by flagx.NewXXX.
// nil, nil will be returned if the flag is not found.
func Lookup(name string) (*flag.Flag, any) {
	return flag.Lookup(name), flagTypes[name]
}

// Visit the flags name and values set in command line
func Visit(fn func(string, string)) {
	flag.Visit(func(f *flag.Flag) {
		lname := strings.ToLower(f.Name)
		value := f.Value.String()
		if IsSecretFlag(lname) {
			value = "secret"
		}
		fn(lname, value)
	})
}

// Visit all the flag name and values, including those not set in command line.
func VisitAll(fn func(string, string)) {
	flag.VisitAll(func(f *flag.Flag) {
		lname := strings.ToLower(f.Name)
		value := f.Value.String()
		if IsSecretFlag(lname) {
			value = "secret"
		}
		fn(lname, value)
	})
}

// Parse parses environment vars(if env.prefix provided) and command-line flags.
//
// Flags set via command-line override flags set via environment vars.
//
// This function must be called instead of flag.Parse() before using any flags in the program.
func Parse() {
	ParseFlagSet(flag.CommandLine, os.Args[1:])
}

// ParseFlagSet parses the given args into the given fs.
func ParseFlagSet(fs *flag.FlagSet, args []string) {
	if err := fs.Parse(args); err != nil {
		log.Fatalf("cannot parse flags %q: %s", args, err)
	}

	// Remember explicitly set command-line flags.
	flagsSet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

	if *envPrefix != "" {
		// Obtain the remaining flag values from environment vars.
		fs.VisitAll(func(f *flag.Flag) {
			if flagsSet[f.Name] {
				// The flag is explicitly set via command-line.
				return
			}
			// Get flag value from environment var.
			fname := getEnvFlagName(f.Name)
			if v := os.Getenv(fname); v != "" {
				if err := fs.Set(f.Name, v); err != nil {
					// Do not use lib/logger here, since it is uninitialized yet.
					log.Fatalf("cannot set flag %s to %q, which is read from env var %q: %s", f.Name, v, fname, err)
				}
			}
		})
	}
}

func getEnvFlagName(s string) string {
	// Substitute dots with underscores, since env var names cannot contain dots.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/311#issuecomment-586354129 for details.
	return strings.ToUpper(*envPrefix + strings.ReplaceAll(s, ".", "_"))
}

func envHelp(s string) string {
	if *envPrefix == "" {
		return ""
	}
	return fmt.Sprintf("(env: %s)", getEnvFlagName(s))
}

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
	flags     = map[string]*flagx{}
)

type flagx struct {
	target   any
	env      string
	required bool
}

func (f *flagx) apply(opts []Option) *flagx {
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *flagx) usage(name string, value any, description string) string {
	usage := description
	if f.env != "" {
		usage += fmt.Sprintf(" (env: %s)", f.env)
	} else {
		usage += fmt.Sprintf(" (env: %s)", getEnvFlagName(name))
	}
	if f.required {
		usage += " (required)"
	}
	/*
		if value != nil {
			usage += fmt.Sprintf(" (default: %v)", value)
		}
	*/
	return usage
}

func newFlag(name string, opts []Option) *flagx {
	x := (&flagx{}).apply(opts)
	flags[name] = x
	return x
}

// NewBool creates a new bool flag.
func NewBool(name string, value bool, usage string, opts ...Option) *bool {
	x := newFlag(name, opts)
	b := flag.Bool(name, value, x.usage(name, value, usage))
	x.target = b
	return b
}

// NewString creates a new string flag.
func NewString(name string, value string, usage string, opts ...Option) *string {
	x := newFlag(name, opts)
	s := flag.String(name, value, x.usage(name, value, usage))
	x.target = s
	return s
}

// NewInt creates a new int flag.
func NewInt(name string, value int, usage string, opts ...Option) *int {
	x := newFlag(name, opts)
	i := flag.Int(name, value, x.usage(name, value, usage))
	x.target = i
	return i
}

// NewInt64 creates a new int64 flag.
func NewInt64(name string, value int64, usage string, opts ...Option) *int64 {
	x := newFlag(name, opts)
	i64 := flag.Int64(name, value, x.usage(name, value, usage))
	x.target = i64
	return i64
}

// NewFloat creates a new float64 flag.
func NewFloat(name string, value float64, usage string, opts ...Option) *float64 {
	x := newFlag(name, opts)
	f := flag.Float64(name, value, x.usage(name, value, usage))
	x.target = f
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
	f, ok := flags[name]
	if !ok {
		return nil, nil
	}
	return flag.Lookup(name), f.target
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
	fs.Usage = func() { Usage(fmt.Sprintf("Usage of %s:", os.Args[0])) }
	if err := fs.Parse(args); err != nil {
		log.Fatalf("cannot parse flags %q: %s", args, err)
	}

	// Remember explicitly set command-line flags.
	flagsSet := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		flagsSet[f.Name] = true
	})

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
		} else if fx, ok := flags[f.Name]; ok && fx.required {
			fmt.Fprintf(os.Stderr, "argument %q is required, run command with --%s or set via %s environment variable\n", f.Name, f.Name, getEnvFlagName(f.Name))
			os.Exit(1)
		}
	})
}

func getEnvFlagName(s string) string {
	if f, ok := flags[s]; ok && f.env != "" {
		return f.env
	}
	// Substitute dots with underscores, since env var names cannot contain dots.
	// See https://github.com/VictoriaMetrics/VictoriaMetrics/issues/311#issuecomment-586354129 for details.
	return strings.ToUpper(*envPrefix + strings.ReplaceAll(s, ".", "_"))
}

// Option for flags
type Option func(*flagx)

// Env customize environment for flag.
func Env(env string) Option {
	return func(f *flagx) { f.env = env }
}

// Required mark the flag MUST BE set via command line or environment
func Required() Option {
	return func(f *flagx) { f.required = true }
}

package flagx

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NewArrayString returns new ArrayString with the given name and description.
func NewArrayString(name, description string) *ArrayString {
	description += "\nSupports an `array` of values separated by comma or specified via multiple flags." + envHelp(name)
	var a ArrayString
	flag.Var(&a, name, description)
	return &a
}

// NewArrayDuration returns new ArrayDuration with the given name and description.
func NewArrayDuration(name, description string) *ArrayDuration {
	description += "\nSupports `array` of values separated by comma or specified via multiple flags." + envHelp(name)
	var a ArrayDuration
	flag.Var(&a, name, description)
	return &a
}

// NewArrayBool returns new ArrayBool with the given name and description.
func NewArrayBool(name, description string) *ArrayBool {
	description += "\nSupports `array` of values separated by comma or specified via multiple flags." + envHelp(name)
	var a ArrayBool
	flag.Var(&a, name, description)
	return &a
}

// NewArrayInt returns new ArrayInt with the given name and description.
func NewArrayInt(name, description string) *ArrayInt {
	description += "\nSupports `array` of values separated by comma or specified via multiple flags." + envHelp(name)
	var a ArrayInt
	flag.Var(&a, name, description)
	return &a
}

// NewArrayBytes returns new ArrayBytes with the given name and description.
func NewArrayBytes(name, description string) *ArrayBytes {
	description += "\nSupports the following optional suffixes for size values: KB, MB, GB, TB, KiB, MiB, GiB, TiB."
	description += "\nSupports `array` of values separated by comma or specified via multiple flags." + envHelp(name)
	var a ArrayBytes
	flag.Var(&a, name, description)
	return &a
}

// ArrayString is a flag that holds an array of strings.
//
// It may be set either by specifying multiple flags with the given name
// passed to NewArray or by joining flag values by comma.
//
// The following example sets equivalent flag array with two items (value1, value2):
//
//	-foo=value1 -foo=value2
//	-foo=value1,value2
//
// Each flag value may contain commas inside single quotes, double quotes, [], () or {} braces.
// For example, -foo=[a,b,c] defines a single command-line flag with `[a,b,c]` value.
//
// Flag values may be quoted. For instance, the following arg creates an array of ("a", "b,c") items:
//
//	-foo='a,"b,c"'
type ArrayString []string

// String implements flag.Value interface
func (a *ArrayString) String() string {
	aEscaped := make([]string, len(*a))
	for i, v := range *a {
		if strings.ContainsAny(v, `,'"{[(`+"\n") {
			v = fmt.Sprintf("%q", v)
		}
		aEscaped[i] = v
	}
	return strings.Join(aEscaped, ",")
}

// Set implements flag.Value interface
func (a *ArrayString) Set(value string) error {
	values := parseArrayValues(value)
	*a = append(*a, values...)
	return nil
}

func parseArrayValues(s string) []string {
	if len(s) == 0 {
		return nil
	}
	var values []string
	for {
		v, tail := getNextArrayValue(s)
		values = append(values, v)
		if len(tail) == 0 {
			return values
		}
		s = tail
		if s[0] == ',' {
			s = s[1:]
		}
	}
}

var closeQuotes = map[byte]byte{
	'"':  '"',
	'\'': '\'',
	'[':  ']',
	'{':  '}',
	'(':  ')',
}

func getNextArrayValue(s string) (string, string) {
	v, tail := getNextArrayValueMaybeQuoted(s)
	if strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		vUnquoted, err := strconv.Unquote(v)
		if err == nil {
			return vUnquoted, tail
		}
		v = v[1 : len(v)-1]
		v = strings.ReplaceAll(v, `\"`, `"`)
		v = strings.ReplaceAll(v, `\\`, `\`)
		return v, tail
	}
	if strings.HasPrefix(v, `'`) && strings.HasSuffix(v, `'`) {
		v = v[1 : len(v)-1]
		v = strings.ReplaceAll(v, `\'`, "'")
		v = strings.ReplaceAll(v, `\\`, `\`)
		return v, tail
	}
	return v, tail
}

func getNextArrayValueMaybeQuoted(s string) (string, string) {
	idx := 0
	for {
		n := strings.IndexAny(s[idx:], `,"'[{(`)
		if n < 0 {
			// The last item
			return s, ""
		}
		idx += n
		ch := s[idx]
		if ch == ',' {
			// The next item
			return s[:idx], s[idx:]
		}
		idx++
		m := indexCloseQuote(s[idx:], closeQuotes[ch])
		idx += m
	}
}

func indexCloseQuote(s string, closeQuote byte) int {
	if closeQuote == '"' || closeQuote == '\'' {
		idx := 0
		for {
			n := strings.IndexByte(s[idx:], closeQuote)
			if n < 0 {
				return 0
			}
			idx += n
			if n := getTrailingBackslashesCount(s[:idx]); n%2 == 1 {
				// The quote is escaped with backslash. Skip it
				idx++
				continue
			}
			return idx + 1
		}
	}
	idx := 0
	for {
		n := strings.IndexAny(s[idx:], `"'[{()}]`)
		if n < 0 {
			return 0
		}
		idx += n
		ch := s[idx]
		if ch == closeQuote {
			return idx + 1
		}
		idx++
		m := indexCloseQuote(s[idx:], closeQuotes[ch])
		if m == 0 {
			return 0
		}
		idx += m
	}
}

func getTrailingBackslashesCount(s string) int {
	n := len(s)
	for n > 0 && s[n-1] == '\\' {
		n--
	}
	return len(s) - n
}

// GetOptionalArg returns optional arg under the given argIdx.
func (a *ArrayString) GetOptionalArg(argIdx int) string {
	x := *a
	if argIdx >= len(x) {
		if len(x) == 1 {
			return x[0]
		}
		return ""
	}
	return x[argIdx]
}

// ArrayBool is a flag that holds an array of booleans values.
//
// Has the same api as ArrayString.
type ArrayBool []bool

// IsBoolFlag  implements flag.IsBoolFlag interface
func (a *ArrayBool) IsBoolFlag() bool { return true }

// String implements flag.Value interface
func (a *ArrayBool) String() string {
	formattedBools := make([]string, len(*a))
	for i, v := range *a {
		formattedBools[i] = strconv.FormatBool(v)
	}
	return strings.Join(formattedBools, ",")
}

// Set implements flag.Value interface
func (a *ArrayBool) Set(value string) error {
	values := parseArrayValues(value)
	for _, v := range values {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		*a = append(*a, b)
	}
	return nil
}

// GetOptionalArg returns optional arg under the given argIdx.
func (a *ArrayBool) GetOptionalArg(argIdx int) bool {
	x := *a
	if argIdx >= len(x) {
		if len(x) == 1 {
			return x[0]
		}
		return false
	}
	return x[argIdx]
}

// ArrayDuration is a flag that holds an array of time.Duration values.
//
// Has the same api as ArrayString.
type ArrayDuration []time.Duration

// String implements flag.Value interface
func (a *ArrayDuration) String() string {
	formattedBools := make([]string, len(*a))
	for i, v := range *a {
		formattedBools[i] = v.String()
	}
	return strings.Join(formattedBools, ",")
}

// Set implements flag.Value interface
func (a *ArrayDuration) Set(value string) error {
	values := parseArrayValues(value)
	for _, v := range values {
		b, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*a = append(*a, b)
	}
	return nil
}

// GetOptionalArgOrDefault returns optional arg under the given argIdx,
// or default value, if argIdx not found.
func (a *ArrayDuration) GetOptionalArgOrDefault(argIdx int, defaultValue time.Duration) time.Duration {
	x := *a
	if argIdx >= len(x) {
		if len(x) == 1 {
			return x[0]
		}
		return defaultValue
	}
	return x[argIdx]
}

// ArrayInt is flag that holds an array of ints.
//
// Has the same api as ArrayString.
type ArrayInt []int

// String implements flag.Value interface
func (a *ArrayInt) String() string {
	x := *a
	formattedInts := make([]string, len(x))
	for i, v := range x {
		formattedInts[i] = strconv.Itoa(v)
	}
	return strings.Join(formattedInts, ",")
}

// Set implements flag.Value interface
func (a *ArrayInt) Set(value string) error {
	values := parseArrayValues(value)
	for _, v := range values {
		n, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		*a = append(*a, n)
	}
	return nil
}

// GetOptionalArgOrDefault returns optional arg under the given argIdx.
func (a *ArrayInt) GetOptionalArgOrDefault(argIdx, defaultValue int) int {
	x := *a
	if argIdx < len(x) {
		return x[argIdx]
	}
	if len(x) == 1 {
		return x[0]
	}
	return defaultValue
}

// ArrayBytes is flag that holds an array of Bytes.
//
// Has the same api as ArrayString.
type ArrayBytes []*Bytes

// String implements flag.Value interface
func (a *ArrayBytes) String() string {
	x := *a
	formattedBytes := make([]string, len(x))
	for i, v := range x {
		formattedBytes[i] = v.String()
	}
	return strings.Join(formattedBytes, ",")
}

// Set implemented flag.Value interface
func (a *ArrayBytes) Set(value string) error {
	values := parseArrayValues(value)
	for _, v := range values {
		var b Bytes
		if err := b.Set(v); err != nil {
			return err
		}
		*a = append(*a, &b)
	}
	return nil
}

// GetOptionalArgOrDefault returns optional arg under the given argIdx.
func (a *ArrayBytes) GetOptionalArgOrDefault(argIdx int, defaultValue int64) int64 {
	x := *a
	if argIdx < len(x) {
		return x[argIdx].N
	}
	if len(x) == 1 {
		return x[0].N
	}
	return defaultValue
}

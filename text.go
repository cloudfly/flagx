package flagx

import (
	"encoding/base64"
	"flag"
)

// Text is a flag that used to hold complex content, such as json.
// the passed value should be base64-encoded string, it will be decoded before assignment.
type Text struct {
	Value string
}

func (t *Text) MarshalText() ([]byte, error) {
	r := base64.StdEncoding.EncodeToString([]byte(t.Value))
	return []byte(r), nil
}

func (t *Text) UnmarshalText(text []byte) error {
	r, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}
	t.Value = string(r)
	return nil
}

// NewText creates a new text flag.
func NewText(name string, value string, description string) *Text {
	t := Text{Value: value}
	description += "\nThe flag value shoud be a base64-encoded string, it will be decoded before assignment" + envHelp(name)
	flag.TextVar(&t, name, &t, description+envHelp(name))
	return &t
}

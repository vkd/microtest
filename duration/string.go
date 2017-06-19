package duration

import (
	"bytes"
	"encoding/json"
	"time"

	"gopkg.in/yaml.v2"
)

type StringDuration time.Duration

var (
	testStringDuration = StringDuration(time.Second)

	_ json.Unmarshaler = &testStringDuration
	_ yaml.Unmarshaler = &testStringDuration
)

func (s *StringDuration) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	d, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*s = StringDuration(d)
	return nil
}

func (s *StringDuration) UnmarshalYAML(f func(interface{}) error) error {
	var str string
	err := f(&str)
	if err != nil {
		return err
	}
	d, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*s = StringDuration(d)
	return nil
}

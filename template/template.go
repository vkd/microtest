package template

import (
	"bytes"
	"text/template"
)

type D map[string]string

func String(s string, d D) (string, error) {
	t := template.New("string").Option("missingkey=error")
	// t = t.Delims("{", "}")

	t, err := t.Parse(s)
	if err != nil {
		return "", err
	}

	var bs bytes.Buffer
	err = t.Execute(&bs, d)
	if err != nil {
		return "", err
	}

	return bs.String(), nil
}

func StringDefault(s string, d D) string {
	out, err := String(s, d)
	if err != nil {
		return s
	}
	return out
}

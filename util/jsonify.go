package util

import (
	"encoding/json"
	"errors"
	"regexp"
)

func UnmarshalString(txt string, v interface{}) (err error) {
	exp := regexp.MustCompile(`(?ms)module.export.*(\{.*\})`)
	s := exp.FindStringSubmatch(txt)
	if s != nil {
		exp = regexp.MustCompile(`([^" ]+) *: *(['"])`)
		b := exp.ReplaceAllString(s[1], `"${1}": ${2}`)
		exp = regexp.MustCompile(`'(.*)'`)
		b = exp.ReplaceAllString(b, `"${1}"`)
		exp = regexp.MustCompile(`\s+([^"]+):([^"]+)`)
		b = exp.ReplaceAllString(b, `"${1}":${2}`)
		err = json.Unmarshal([]byte(b), v)
	} else {
		err = errors.New("Cannot parse input string")
	}
	return
}

func Unmarshal(data []byte, v interface{}) (err error) {
	exp := regexp.MustCompile(`(?s)module.export.*(\{.*\})`)
	s := exp.FindSubmatch(data)
	if s != nil {
		exp = regexp.MustCompile(`([^" ]+) *: *(['"])`)
		b := exp.ReplaceAll(s[1], []byte(`"${1}": ${2}`))
		exp = regexp.MustCompile(`'(.*)'`)
		b = exp.ReplaceAll(b, []byte(`"${1}"`))
		exp = regexp.MustCompile(`\s+([^"]+):([^"]+)`)
		b = exp.ReplaceAll(b, []byte(`"${1}":${2}`))
		err = json.Unmarshal(b, v)
	} else {
		err = errors.New("Cannot parse input string")
	}
	return
}

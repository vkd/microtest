package main

import (
	"fmt"
	"io/ioutil"
	"microtest/cmp"
	"microtest/duration"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Name    string `yaml:"name"`
	Image   string `yaml:"image"`
	Command string `yaml:"command"`

	Port int `yaml:"port"`

	PingRequest *PingRequestConfig `yaml:"ping_request"`
	// Sleep duration.StringDuration `yaml:"sleep"`

	// // MockServices []string `yaml:"mockservices"`

	Mocks MockConfigs `yaml:"mocks"`

	Services []*ServiceConfig `yaml:"services"`

	Tests []*TestConfig `yaml:"tests"`
}

type TestConfig struct {
	Name    string            `yaml:"name"`
	Sleep   int               `yaml:"sleep"`
	Request *RequestConfig    `yaml:"request"`
	Mocks   MockConfigs       `yaml:"mocks"`
	Expect  *ExpectMockConfig `yaml:"expect"`
}

type RequestConfig struct {
	URL     string                  `yaml:"url"`
	Method  string                  `yaml:"method"`
	Headers map[string]string       `yaml:"headers"`
	Timeout duration.StringDuration `yaml:"timeout"`
	Body    string                  `yaml:"body"`
	Expect  *ExpectConfig           `yaml:"expect"`
}

type PingRequestConfig struct {
	RequestConfig `yaml:",inline"`
	Count         int `yaml:"count"`
}

type ExpectConfig struct {
	cmp.Comparator
	Status int    `yaml:"status"`
	Body   string `yaml:"body"`
}

type ExpectMockConfig struct {
	ExpectConfig

	Mocks map[string]*ExpectConfig `yaml:"mocks"`
}

type MockConfigs map[string][]MockConfig

type MockConfig struct {
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
	// Request struct {
	// 	cmp.Comparator
	// 	Body string `json:"body"`
	// } `yaml:"request"`
	Out string `yaml:"body"`

	// index int
}

func (m *MockConfig) String() string {
	return fmt.Sprintf("Method: %q, Url: %q, Out: %q", m.Method, m.URL, m.Out)
}

type ServiceConfig struct {
	Image string    `yaml:"image"`
	Env   EnvConfig `yaml:"env"`
}

type EnvConfig map[string]string

func (e EnvConfig) Slice() []string {
	res := make([]string, 0, len(e))
	for k, v := range e {
		res = append(res, fmt.Sprintf("%s:%s", k, v))
	}
	return res
}

func ReadConfig(path string) (*Config, error) {
	fl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	err = yaml.Unmarshal(fl, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

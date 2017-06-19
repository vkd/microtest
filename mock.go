package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Mock struct {
	IsDebug bool

	conf []MockConfig

	host string

	Requests   []*requestResult
	mxRequests sync.Mutex
}

func (m *MockConfig) Equal(method, u string, body []byte) error {
	if m == nil {
		return nil
	}
	err := m.equalURL(u)
	if err != nil {
		return err
	}
	if m.Method != "" {
		if m.Method != method {
			return fmt.Errorf("Wrong method: %s (expect: %s)", method, m.Method)
		}
	}

	return nil
}

func (m *MockConfig) equalURL(u string) error {
	if m.URL == "" {
		return nil
	}
	mockURL, err := url.Parse(m.URL)
	if err != nil {
		return fmt.Errorf("Error on parse config url (%s): %v", m.URL, err)
	}

	inputURL, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("Error on parse input url (%s): %v", u, err)
	}

	if mockURL.Path != inputURL.Path {
		return fmt.Errorf("Wrong url: %s (expect: %s)", inputURL.Path, mockURL.Path)
	}

	for k, vs := range mockURL.Query() {
		if inputVs, ok := inputURL.Query()[k]; ok {
			if len(vs) > len(inputVs) {
				return fmt.Errorf("Wrong query %q len: %s (expect: %s)", k, strings.Join(inputVs, ","), strings.Join(vs, ","))
			}
			for i, q := range vs {
				if q != inputVs[i] {
					return fmt.Errorf("Wrong query %q on position %d: %s (expect: %s)", k, i, strings.Join(inputVs, ","), strings.Join(vs, ","))
				}
			}
		} else {
			return fmt.Errorf("Not found query key: %s", k)
		}
	}
	return nil
}

func NewMock(conf []MockConfig, host string) *Mock {
	m := &Mock{
		conf:    conf,
		IsDebug: isDebug,

		host: host,
	}
	return m
}

func (m *Mock) CheckExpect(exp *ExpectConfig) error {
	if len(m.Requests) == 0 {
		return fmt.Errorf("mock requests is empty")
	}

	for _, r := range m.Requests {
		err := NewExpect().Check(r, exp)
		return err
	}
	return nil
}

func (m *Mock) handle(method, url string, body io.ReadCloser) (status int, out []byte) {
	status = http.StatusInternalServerError
	out = []byte(fmt.Sprintf(`MICROTEST: MOCK RESPONSE %s:%s%s NOT FOUND`, method, m.host, url))

	defer body.Close()
	bodyBs, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("Error on read body: %v", err)
		return http.StatusOK, []byte("{}")
	}

	req := &requestResult{
		URL:     url,
		RawBody: bodyBs,
	}

	m.mxRequests.Lock()
	m.Requests = append(m.Requests, req)
	m.mxRequests.Unlock()

	if m.IsDebug {
		log.Printf("Start find equal out")
		log.Printf("Count configs: %d", len(m.conf))
		log.Printf("All configs: %v", m.conf)
	}

	for _, c := range m.conf {
		if err := c.Equal(method, url, bodyBs); err != nil {
			if m.IsDebug {
				log.Printf("Error on check equal mock response: %v", err)
			}
			continue
		}
		status = http.StatusOK
		out = []byte(c.Out)
		break
	}

	if m.IsDebug {
		log.Printf("url mocks: %s", url)
		log.Printf("mock host: %s", m.host)
		log.Printf("out: %s", string(out))
	}

	return
}

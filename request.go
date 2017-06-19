package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func sendRequest(ip string, port int, conf *RequestConfig) (*requestResult, error) {
	method := conf.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, fmt.Sprintf("http://%s:%d/%s", ip, port, strings.TrimLeft(conf.URL, "/")), strings.NewReader(conf.Body))
	if err != nil {
		return nil, err
	}

	for hKey, hVal := range conf.Headers {
		req.Header.Set(hKey, hVal)
	}

	timeout := time.Duration(conf.Timeout)
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	client := http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &requestResult{
		URL:     "/" + strings.TrimLeft(conf.URL, "/"),
		Status:  resp.StatusCode,
		RawBody: body,
	}, nil
}

type requestResult struct {
	URL     string
	Status  int
	RawBody []byte
}

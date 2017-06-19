package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultMocksPort = 9001
)

// import (
// 	"context"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"strings"
// 	"sync"
// 	"time"
// )

type Mocks struct {
	Mocks map[string]*Mock
	mx    sync.Mutex

	conf MockConfigs

	IsDebug bool
	Port    int

	srv *http.Server
}

func NewMocks(conf MockConfigs) *Mocks {
	if isDebug {
		log.Printf("Mocks config:")
	}

	m := &Mocks{
		Mocks: map[string]*Mock{},
		conf:  conf,

		Port: DefaultMocksPort,
	}
	return m
}

func (m *Mocks) ResetMocks() {
	m.mx.Lock()
	m.Mocks = map[string]*Mock{}
	m.mx.Unlock()

	m.UpdateConfigs(m.conf)
}

func (m *Mocks) UpdateConfigs(conf MockConfigs) {
	m.mx.Lock()

	for mockName, c := range conf {
		if mm, ok := m.Mocks[mockName]; ok {
			mm.conf = c
		} else {
			m.Mocks[mockName] = NewMock(c, mockName)
		}
	}

	m.mx.Unlock()
}

func (m *Mocks) Run() error {
	if m.srv == nil {
		m.srv = &http.Server{
			Addr: fmt.Sprintf(":%d", m.Port),

			ReadTimeout:       5 * time.Second,
			WriteTimeout:      5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		}
	} else {
		err := m.srv.Shutdown(context.Background())
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", m.handle)

	m.srv.Handler = mux
	go func() {
		if isDebug {
			log.Printf("Mocks listen port: %d", m.Port)
		}
		err := m.srv.ListenAndServe()
		if err != nil {
			log.Printf("Error on listen mocks: %v", err)
		}
	}()
	return nil
}

func (m *Mocks) Stop() error {
	if m.srv != nil {
		return m.srv.Shutdown(context.Background())
	}
	return nil
}

func (m *Mocks) CheckExpect(exp map[string]*ExpectConfig) error {
	for mockName, e := range exp {
		mock := m.getMock(mockName)
		if mock == nil {
			return fmt.Errorf("mock not found: %q", mockName)
		}

		err := mock.CheckExpect(e)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Mocks) getMock(host string) *Mock {
	m.mx.Lock()
	mock := m.Mocks[host]
	if mock == nil {
		log.Printf("Create new mock: %q", host)
		mock = NewMock(m.conf[host], host)
		m.Mocks[host] = mock
	}
	m.mx.Unlock()
	return mock
}

func (m *Mocks) handle(w http.ResponseWriter, r *http.Request) {
	mock := m.getMock(trimHost(r.Host))

	if mock == nil {
		m.defaultHandle(w, r)
	}

	status, out := mock.handle(r.Method, r.URL.Path, r.Body)
	w.WriteHeader(status)
	_, err := w.Write(out)
	if err != nil {
		log.Printf("Error on write response to host %q: %v", r.Host, err)
	}
}

func (m *Mocks) defaultHandle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(`{"error": "mock_not_found", "result": null}`))
	if err != nil {
		log.Printf("Error on write default response to host %q: %v", r.Host, err)
	}
}

func trimHost(host string) string {
	if i := strings.Index(host, ":"); i > 0 {
		return host[:i]
	}
	return host
}

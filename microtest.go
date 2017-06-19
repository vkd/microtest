package main

import (
	"context"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/fsouza/go-dockerclient"
)

const (
	EnvHostWorkdir = "MICROTEST_HOST_WORKDIR"
)

var (
	ContainerWorkdirPath    = "/microtest"
	ContainerMicrotestsPath = path.Join(ContainerWorkdirPath, "microtests")
)

type Microtest struct {
	Conf *Config

	IP string

	mocks         *Mocks
	testedService *TestedService
}

func (m *Microtest) Start(ctx context.Context, dc *docker.Client) (err error) {
	conf := m.Conf

	log.Printf("Start mocks")
	m.mocks = NewMocks(conf.Mocks)
	err = m.mocks.Run()
	if err != nil {
		log.Printf("Error on run mocks")
		return err
	}

	var services []*Service
	var extraHosts []string
	m.testedService = NewTestedService(conf, conf.Port)

	defer func() {
		if m.testedService != nil {
			errStop := m.testedService.Stop()
			if errStop != nil {
				log.Printf("Error on stop %q tested container: %v", conf.Image, errStop)
			}
			if err != nil {
				m.testedService.PrintLogs()
			}
			errRem := m.testedService.Remove()
			if errRem != nil {
				log.Printf("Error on remove %q tested container: %v", conf.Image, errRem)
			}
		}

		for _, s := range services {
			s.Stop(dc)
		}
	}()

	log.Printf("Start services")
	for _, sc := range m.Conf.Services {
		srv := NewService(sc)
		ip, err := srv.Start(dc)
		if err != nil {
			log.Printf("Error on start %q service: %v", srv.Name, err)
			return err
		}
		services = append(services, srv)
		extraHosts = append(extraHosts, fmt.Sprintf("%s: %s", srv.Name, ip))
	}

	log.Printf("Sleep 10 sec")
	time.Sleep(10 * time.Second)

	LogPrintfH1("Start tests: %s", conf.Name)

	err = m.testedService.Run(&RunConfig{
		MocksPort:  m.mocks.Port,
		SelfIP:     m.IP,
		ExtraHosts: extraHosts,
	})
	if err != nil {
		return err
	}

	done := make(chan struct{})

	go func() {
		err = m.startTests(conf)
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		err = ctx.Err()
	}

	return err
}

func (m *Microtest) startTests(conf *Config) error {
	err := m.testedService.PingRequest(conf.PingRequest)
	if err != nil {
		log.Printf("Error on ping request: %v", err)
		return err
	}

	for _, t := range conf.Tests {
		m.mocks.ResetMocks()

		err = m.test(t)
		if err != nil {
			LogPrintfH1("Error on test (%s): %v", t.Name, err)
			return err
		}
	}

	log.Print("\n")
	LogPrintfH1("All tests completed successful")

	return nil
}

func (m *Microtest) test(t *TestConfig) error {
	if isDebug {
		LogPrintfH2("Start test: %s", t.Name)
	} else {
		log.Printf("Start test: %s", t.Name)
	}

	if t.Sleep > 0 {
		if isDebug {
			log.Printf("Sleep on %d sec", t.Sleep)
		}
		time.Sleep(time.Duration(t.Sleep) * time.Second)
	}

	m.mocks.UpdateConfigs(t.Mocks)

	err := m.testedService.Request(t.Request, &t.Expect.ExpectConfig)
	if err != nil {
		return err
	}

	err = m.mocks.CheckExpect(t.Expect.Mocks)
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"microtest/duration"
	"microtest/template"
	"os"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

type TestedService struct {
	conf *Config
	cnt  *docker.Container

	ip   string
	port int
}

func NewTestedService(conf *Config, port int) *TestedService {
	t := &TestedService{
		conf: conf,
		port: port,
	}

	if t.port == 0 {
		t.port = 9000
	}

	return t
}

type RunConfig struct {
	MocksPort  int
	SelfIP     string
	ExtraHosts []string
}

func (t *TestedService) Run(mc *RunConfig) error {
	if t.conf.Image == "" {
		return errors.New("'image' not found in config")
	}

	for mockName := range t.conf.Mocks {
		mc.ExtraHosts = append(mc.ExtraHosts, fmt.Sprintf("%s: %s", mockName, mc.SelfIP))
	}

	// var links []string
	// for mockName := range t.conf.Mocks {
	// 	links = append(links, fmt.Sprintf("%s:%s", mockName, mockName))
	// }

	if isDebug {
		log.Printf("Extra hosts: %v", mc.ExtraHosts)
		// log.Printf("Links: %v", links)
	}

	cnt, err := dc.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: t.conf.Image,
			Cmd: strings.Fields(template.StringDefault(t.conf.Command,
				template.D{
					"workdir": "/builds/localhost",
				})),
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/builds/localhost/microtests:ro", os.Getenv(EnvHostWorkdir)),
			},
			ExtraHosts: mc.ExtraHosts,
			// Links:      mc.Links,
		},
	})
	if err != nil {
		log.Printf("Error on create tested %q container: %v", t.conf.Image, err)
		return err
	}

	err = dc.StartContainer(cnt.ID, &docker.HostConfig{})
	if err != nil {
		log.Printf("Error on start tested %q container: %v", t.conf.Image, err)
		return err
	}

	cnt, err = dc.InspectContainer(cnt.ID)
	if err != nil {
		log.Printf("Error on inspect %q container: %v", t.conf.Image, err)
		return err
	}

	t.cnt = cnt

	t.ip = cnt.NetworkSettings.IPAddress
	if isDebug {
		log.Printf("Tested service (%s) started on ip: %s", t.conf.Image, t.ip)
	}
	return nil
}

func (t *TestedService) Stop() error {
	if t.cnt == nil {
		return nil
	}
	err := dc.StopContainer(t.cnt.ID, 3)
	if err != nil {
		return err
	}
	return nil
}

func (t *TestedService) Remove() error {
	if t.cnt != nil {
		err := dc.RemoveContainer(docker.RemoveContainerOptions{
			ID:    t.cnt.ID,
			Force: true,
		})
		if err != nil {
			log.Printf("Error on remove %q container", t.conf.Image)
			return err
		}
	}
	return nil
}

func (t *TestedService) Request(r *RequestConfig, ex *ExpectConfig) error {
	if r == nil {
		return nil
	}

	res, err := sendRequest(t.ip, t.port, r)
	if err != nil {
		log.Printf("Error on send request to (%s:%d): %v", t.ip, t.port, err)
		return err
	}

	err = NewExpect().Check(res, r.Expect)
	if err != nil {
		log.Printf("Error on check request expect: %v", err)
		return err
	}

	err = NewExpect().Check(res, ex)
	if err != nil {
		log.Printf("Error on check expect: %v", err)
		return err
	}

	return nil
}

func (t *TestedService) PingRequest(r *PingRequestConfig) error {
	if r == nil {
		return nil
	}
	if time.Duration(r.Timeout) == 0 {
		r.Timeout = duration.StringDuration(time.Second)
	}

	if r.Count == 0 {
		r.Count = 3
	}

	if r.URL == "" {
		r.URL = "/ping"
	}

	var sleepSkipFirst func(time.Duration)
	sleepSkipFirst = func(time.Duration) { sleepSkipFirst = time.Sleep }

	var err error
	for i := 0; i < r.Count; i++ {
		sleepSkipFirst(time.Second)

		res, err := sendRequest(t.ip, t.port, &r.RequestConfig)
		if err != nil {
			if isDebug {
				log.Printf("Error on send ping request to (%s:%d): %v", t.ip, t.port, err)
			}
			continue
		}

		err = NewExpect().Check(res, &ExpectConfig{Status: 200})
		if err != nil {
			if isDebug {
				log.Printf("Error on check ping request expect: %v", err)
			}
			continue
		}

		if err == nil {
			return nil
		}
	}
	return err
}

func (t *TestedService) PrintLogs() {
	if t == nil || t.cnt == nil {
		log.Print("Error on print logs: tested service is nil")
		return
	}
	bottom := LogPrintH2Borders("Logs tested service %q", t.conf.Image)

	var bs bytes.Buffer

	err := dc.Logs(docker.LogsOptions{
		Container: t.cnt.ID,
		Stdout:    true,
		Stderr:    true,

		OutputStream: &bs,
		ErrorStream:  &bs,
	})
	log.Print(bs.String())
	bottom()
	if err != nil {
		log.Printf("Error on show logs for %s: %v", t.conf.Image, err)
	}
}

package main

import (
	"bytes"
	"log"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type Service struct {
	Name string

	conf *ServiceConfig

	cntID string
}

func NewService(conf *ServiceConfig) *Service {
	name := conf.Image
	if i := strings.Index(name, ":"); i > 0 {
		name = name[:i]
	}
	s := &Service{
		Name: name,
		conf: conf,
	}
	return s
}

func (s *Service) Start(dc *docker.Client) (ip string, err error) {
	err = dc.PullImage(docker.PullImageOptions{
		Repository: s.conf.Image,
	}, docker.AuthConfiguration{})
	if err != nil {
		log.Printf("Error on pull %q image: %v", s.conf.Image, err)
		return "", err
	}

	cnt, err := dc.CreateContainer(docker.CreateContainerOptions{
		Name: s.Name,
		Config: &docker.Config{
			Image: s.conf.Image,
			Env:   s.conf.Env.Slice(),
		},
	})
	if err != nil {
		log.Printf("Error on create %q service container: %v", s.Name, err)
		return "", err
	}
	s.cntID = cnt.ID

	err = dc.StartContainer(cnt.ID, &docker.HostConfig{})
	if err != nil {
		log.Printf("Error on start %q service container: %v", s.Name, err)
		return "", err
	}

	cnt, err = dc.InspectContainer(cnt.ID)
	if err != nil {
		log.Printf("Error on inspect %q service container: %v", s.Name, err)
		return "", err
	}

	log.Printf("%q service started on: %s", s.Name, cnt.NetworkSettings.IPAddress)

	return cnt.NetworkSettings.IPAddress, nil
}

func (s *Service) Stop(dc *docker.Client) {
	if s.cntID != "" {
		err := dc.StopContainer(s.cntID, 2)
		if err != nil {
			log.Printf("Error on stop %q service: %v", s.Name, err)
		}

		if isDebug {
			s.PrintLogs()
		}

		err = dc.RemoveContainer(docker.RemoveContainerOptions{
			ID:    s.cntID,
			Force: true,
		})
		if err != nil {
			log.Printf("Error on stop %q service: %v", s.Name, err)
		}
	}
}

func (s *Service) PrintLogs() {
	if s == nil || s.cntID == "" {
		log.Print("Error on print logs: service is nil")
		return
	}
	bottom := LogPrintH2Borders("Logs %q service", s.Name)

	var bs bytes.Buffer

	err := dc.Logs(docker.LogsOptions{
		Container: s.cntID,
		Stdout:    true,
		Stderr:    true,

		OutputStream: &bs,
		ErrorStream:  &bs,
	})
	log.Print(bs.String())
	bottom()
	if err != nil {
		log.Printf("Error on show logs for %q service: %v", s.Name, err)
	}
}

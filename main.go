package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"sort"
	"strings"
	"syscall"

	"github.com/fsouza/go-dockerclient"

	"log"
)

const (
	SelfDefaultContainerName = "microtest"
	SelfDefaultImage         = "microtest:go"
)

var (
	isDebug  = false
	testPath = "./microtests"

	dc *docker.Client
)

func parseArgs(args []string) {
	for len(args) > 0 {
		switch args[0] {
		case "--debug", "-debug", "debug":
			isDebug = true
		default:
			testPath = args[0]
		}
		args = args[1:]
	}
}

func notifySignal() <-chan struct{} {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)

	out := make(chan struct{})

	go func() {
		for range c {
			select {
			case out <- struct{}{}:
			default:
			}
		}
	}()

	return out
}

func main() {
	parseArgs(os.Args[1:])

	log.SetFlags(0)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-notifySignal()
		log.Printf("Stop container")
		cancel()
	}()

	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Error on get docker client: %v", err)
	}

	dc = dockerClient

	selfContainer := getSelfContainer()

	if selfContainer == nil {
		err := runContainer(ctx)
		if err != nil {
			if isDebug {
				log.Printf("Error on run container: %v", err)
			}
			os.Exit(1)
		}
		return
	}

	err = startMicrotests(ctx, selfContainer)
	if err != nil {
		log.Fatalf("Error on microtest: %v", err)
		// os.Exit(1)
	}
}

func getSelfContainer() *docker.Container {
	containerName := os.Getenv("HOSTNAME")
	if containerName == "" {
		containerName = SelfDefaultContainerName
	}

	cont, err := dc.InspectContainer(containerName)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			log.Printf("Error on get self container (%T): %v", err, err)
		}
	}
	return cont
}

func runContainer(ctx context.Context) error {
	var contID string
	defer func() {
		if contID != "" {
			err := dc.RemoveContainer(docker.RemoveContainerOptions{
				ID:    contID,
				Force: true,
			})
			if err != nil {
				log.Printf("Error on remove 'microtest' container: %v", err)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		if contID != "" {
			err := dc.StopContainer(contID, 5)
			if err != nil {
				log.Printf("Error on stop microtest container: %v", err)
			}
		}
	}()

	stat, err := os.Stat(testPath)
	if err != nil {
		log.Printf("Error on stat path (%s): %v", testPath, err)
		return err
	}
	hostWorkdir := absPathTests(testPath)
	testFile := ""
	if !stat.IsDir() {
		hostWorkdir, testFile = path.Split(hostWorkdir)
	}

	cmd := []string{path.Join(ContainerMicrotestsPath, testFile)}
	if isDebug {
		cmd = append(cmd, "--debug")
	}

	microCont, err := dc.CreateContainer(docker.CreateContainerOptions{
		Name: SelfDefaultContainerName,
		Config: &docker.Config{
			Image: SelfDefaultImage,
			Cmd:   cmd,
			Env: []string{
				fmt.Sprintf("%s=%s", EnvHostWorkdir, hostWorkdir),
			},
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
				fmt.Sprintf("%s:%s:ro",
					hostWorkdir,
					ContainerMicrotestsPath,
				),
			},
		},
	})
	if err != nil {
		log.Printf("Error on create %s container: %v", SelfDefaultImage, err)
		return err
	}

	contID = microCont.ID

	err = dc.StartContainer(microCont.ID, &docker.HostConfig{
		AutoRemove: true,
	})
	if err != nil {
		log.Printf("Error on start container (%s): %v", SelfDefaultImage, err)
		return err
	}

	log.Printf("Microtest container started\n\n")

	// go func() {
	// 	defer func() {
	// 		err := dc.RemoveContainer(docker.RemoveContainerOptions{
	// 			ID:    microCont.ID,
	// 			Force: true,
	// 		})
	// 		if err != nil {
	// 			log.Printf("Error on remove microtest container: %v", err)
	// 		}
	// 	}()
	// 	<-notifySignal()
	// 	err := dc.StopContainer(microCont.ID, 7)
	// 	if err != nil {
	// 		log.Printf("Error on stop container: %v", err)
	// 	}
	// }()

	err = dc.Logs(docker.LogsOptions{
		Container: microCont.ID,

		OutputStream: os.Stdout,
		ErrorStream:  os.Stdout,

		Follow: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		log.Printf("Error on get log from %s container: %v", SelfDefaultImage, err)
		return err
	}

	inspectCont, err := dc.InspectContainer(microCont.ID)
	if err != nil {
		log.Printf("Error on inspect %s container: %v", SelfDefaultImage, err)
		return err
	}

	// Logs() is locker method
	// when Logs() is over, then container already is stopped
	if inspectCont.State.Running {
		log.Printf("%s container is running", SelfDefaultImage)
		return errors.New("container is running")
	}

	if inspectCont.State.ExitCode != 0 {
		return errors.New("Exit code is not 0")
	}
	return nil
}

func startMicrotests(ctx context.Context, selfContainer *docker.Container) error {
	info, err := os.Stat(testPath)
	if err != nil {
		return err
	}

	if isDebug {
		log.Printf("Env in microtest: %v", os.Environ())
	}

	var tests []string

	if info.IsDir() {
		infos, err := ioutil.ReadDir(testPath)
		if err != nil {
			return err
		}
		for _, i := range infos {
			if !i.IsDir() && (strings.HasSuffix(i.Name(), ".yaml") || strings.HasSuffix(i.Name(), ".yml")) {
				tests = append(tests, path.Join(testPath, i.Name()))
			}
		}
	} else {
		tests = append(tests, testPath)
	}

	sort.Strings(tests)

	if isDebug {
		log.Printf("Tests: %v", tests)
	}

	for _, test := range tests {
		err = startMicrotest(ctx, dc, selfContainer, test)
		if err != nil {
			if isDebug {
				log.Printf("Error on start microtest: %v", err)
			}
			return err
		}
	}

	return nil
}

func startMicrotest(ctx context.Context, dc *docker.Client, selfContainer *docker.Container, configPath string) error {
	conf, err := ReadConfig(configPath)
	if err != nil {
		log.Printf("Error on read config: %v", err)
		return err
	}

	return (&Microtest{
		Conf: conf,
		IP:   selfContainer.NetworkSettings.IPAddress,
	}).Start(ctx, dc)
}

func absPathTests(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}

	if strings.HasPrefix(p, "~") {
		return path.Join(os.Getenv("HOME"), p[1:])
	}
	if strings.HasPrefix(p, "./") {
		return path.Join(os.Getenv("PWD"), p[2:])
	}
	return path.Join(os.Getenv("HOME"), p)
}

// func runTests(conf *Config) {

// 	pwd, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}

// 	cmd, err := template.String(conf.Command, template.D{
// 		"workdir": "/builds/localhost",
// 	})
// 	if err != nil {
// 		panic(err)
// 	}

// 	cont, err := dc.CreateContainer(docker.CreateContainerOptions{
// 		Config: &docker.Config{
// 			Image: conf.Image,
// 			Cmd:   strings.Fields(cmd),
// 		},
// 		HostConfig: &docker.HostConfig{
// 			Binds: []string{
// 				fmt.Sprintf("%s:/builds/localhost:ro", pwd),
// 			},
// 		},
// 	})

// 	defer func() {
// 		if dc != nil {
// 			bs := new(bytes.Buffer)
// 			err := dc.Logs(docker.LogsOptions{
// 				Container:    cont.ID,
// 				OutputStream: bs,
// 				ErrorStream:  bs,
// 				Stdout:       true,
// 				Stderr:       true,
// 			})
// 			if err != nil {
// 				log.Printf("Error on get logs: %v", err)
// 			} else {
// 				log.Printf("---------------")
// 				log.Printf("Container logs:")
// 				log.Printf("---------------")
// 				log.Printf("%s", bs.String())
// 				log.Printf("---------------")
// 			}
// 			// err = dc.KillContainer(docker.KillContainerOptions{ID: cont.ID})
// 			err = dc.RemoveContainer(docker.RemoveContainerOptions{ID: cont.ID, Force: true})
// 			if err != nil {
// 				panic(err)
// 			}
// 		}
// 		r := recover()
// 		if r == nil {
// 			return
// 		}
// 		log.Printf("=====================")
// 		log.Printf("Panic: %v", r)
// 		log.Printf("=====================")
// 		os.Exit(1)
// 	}()

// 	err = dc.StartContainer(cont.ID, &docker.HostConfig{})
// 	if err != nil {
// 		panic(err)
// 	}

// 	time.Sleep(1 * time.Second)

// 	contConf, err := dc.InspectContainer(cont.ID)
// 	if err != nil {
// 		panic(err)
// 	}

// 	ip := contConf.NetworkSettings.IPAddress
// 	log.Printf("Start service on ip: %s", ip)
// 	if ip == "" {
// 		panic("Testing service not started")
// 	}

// 	mocks := NewMocks(conf.Mocks)
// 	mocks.IsDebug = true
// 	mocks.Run()

// 	ts := NewTestedService(nil)
// 	ts.IP = ip
// 	ts.Port = 9000

// 	for _, test := range conf.Tests {
// 		err = ts.Request(test.Request, test.Expect)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	// res, err := sendRequest(ip, 9000, conf.Tests[0].Request)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// log.Printf("Status: %d, Body: %s", res.Status, res.RawBody)
// }

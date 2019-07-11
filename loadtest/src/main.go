package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.sc-corp.net/scaddlive/women-who-go.git/loadtest/pkg/kargo"
)

var (
	hostname string
	replicas int
)

var resultsBuffer = &[]*result{}
var parser LogParser

func init() {
	flag.IntVar(&replicas, "replicas", 1, "Number of replicas")

}

func main() {
	flag.Parse()

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Starting loadtest on %s...", hostname)
	errChan := make(chan error, 10)
	signalChan := make(chan os.Signal, 1)

	var dm *kargo.DeploymentManager

	if kargo.EnableKubernetes {
		link, err := kargo.Upload(kargo.UploadConfig{
			ProjectID:  "staging-glass-pen-358",
			BucketName: "test-binaries",
			ObjectName: "loadtest",
			BuildPath:  "../loadtest/src",
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dm = kargo.New()
		config := kargo.DeploymentConfig{
			Args:      []string{},
			Name:      "loadtest",
			BinaryURL: link,
			Namespace: "default",
			Replicas:  1,
		}
		err = dm.Create(config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		go runScalingLoop(dm, config)

		err = dm.Logs(parser)
		if err != nil {
			fmt.Println("Local logging has been disabled.")
		}

	} else {
		go runMain(errChan, hostname, signalChan)
	}

	renderer := makeChartRenderer(parser)
	go renderer.Start()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			if err != nil {
				fmt.Printf("%s - %s\n", hostname, err)
				os.Exit(1)
			}
		case <-signalChan:
			fmt.Printf("%s - Shutdown signal received, exiting...\n", hostname)
			if kargo.EnableKubernetes {
				err := dm.Delete()
				if err != nil {
					fmt.Printf("%s - %s\n", hostname, err)
					os.Exit(1)
				}
			}
			os.Exit(0)
		}
	}

}

func runScalingLoop(dm *kargo.DeploymentManager, config kargo.DeploymentConfig) {
	scaleTo := replicas
	fmt.Printf("started scaling loop from 1 to %d\n", scaleTo)

	for i := 2; i < scaleTo; i++ {
		time.Sleep(time.Duration(15) * time.Second)
		fmt.Printf("Scaling to %d replicas\n", i)

		err := dm.Scale(config, i)
		if err != nil {
			fmt.Println("Failed to scale")
			fmt.Println(err)
		}
		err = dm.Logs(os.Stdout)
		if err != nil {
			fmt.Println("Failed to update log parser")
		}
	}
}

type loadtestConfig struct {
	endpoint          string
	requestsPerMinute int
	batchSize         int
	numWorkers        int
	reqChannel        chan *http.Request
	stdoutChannel     chan string
	errChan           chan error
	sigChan           chan os.Signal
	hostname          string
	httpTimeoutSecs   int
}

func runMain(errChan chan error, hostname string, sigChan chan os.Signal) {
	ctx := context.Background()
	numWorkers := 10
	config := &loadtestConfig{
		endpoint:          "http://35.232.238.57/",
		requestsPerMinute: 100 * numWorkers,
		batchSize:         numWorkers,
		numWorkers:        numWorkers,
		reqChannel:        make(chan *http.Request, numWorkers),
		stdoutChannel:     make(chan string),
		errChan:           errChan,
		sigChan:           sigChan,
		hostname:          hostname,
		httpTimeoutSecs:   10,
	}

	reqGenerator := makeReqGenerator(config)
	clientMgr := makeClientManager(config)

	go reqGenerator.generate(ctx)
	go clientMgr.startWorkers(ctx, config)

	for msg := range config.stdoutChannel {
		logLine(msg)
	}
}

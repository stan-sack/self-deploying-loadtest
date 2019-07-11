package main

import (
	"context"
	"os"
)

type clientManager struct {
	numWorkers    int
	endpoint      string
	errorChannel  chan error
	sigChannel    chan os.Signal
	hostname      string
	reqPerMinute  int
	batchSize     int
	stdoutChannel chan string
}

func makeClientManager(config *loadtestConfig) *clientManager {
	return &clientManager{
		numWorkers:    config.numWorkers,
		endpoint:      config.endpoint,
		errorChannel:  config.errChan,
		sigChannel:    config.sigChan,
		hostname:      config.hostname,
		reqPerMinute:  config.requestsPerMinute,
		batchSize:     config.batchSize,
		stdoutChannel: config.stdoutChannel,
	}
}

func (cm *clientManager) startWorkers(ctx context.Context, config *loadtestConfig) {

	for i := 0; i < cm.numWorkers; i++ {
		// time.Sleep(500 * time.Millisecond)
		client := cm.createClient(config)
		go client.startWorking(ctx)
	}
}

func (cm *clientManager) createClient(config *loadtestConfig) *client {
	return makeClient(config)
}

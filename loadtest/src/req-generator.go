package main

import (
	"context"
	"net/http"
	"time"
)

type reqGenerator struct {
	requestsPerMinute int
	batchSize         int
	reqChannel        chan *http.Request
	endpoint          string
	stdoutChannel     chan string
}

func makeReqGenerator(config *loadtestConfig) *reqGenerator {
	return &reqGenerator{
		requestsPerMinute: config.requestsPerMinute,
		batchSize:         config.batchSize,
		reqChannel:        config.reqChannel,
		endpoint:          config.endpoint,
		stdoutChannel:     config.stdoutChannel,
	}
}

func (rg *reqGenerator) generate(ctx context.Context) {
	defer close(rg.reqChannel)

	messagesPerSec := int64(float64(rg.requestsPerMinute) / float64(60))
	durationPerMessage := time.Duration(int64(time.Second) * int64(rg.batchSize) / messagesPerSec)

	for {
		deadline := time.Now().Add(durationPerMessage)
		for i := 0; i < rg.batchSize; i++ {
			req, err := http.NewRequest("GET", rg.endpoint, nil)
			if err != nil {
				rg.stdoutChannel <- err.Error()
			} else {
				rg.reqChannel <- req
			}

		}

		time.Sleep(time.Until(deadline))

		// check if the context is done, close out channel and stop
		select {
		case <-ctx.Done():
			return // this will close the "out" channel as well
		default:
		}
	}
}

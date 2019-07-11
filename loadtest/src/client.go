package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type client struct {
	httpClient    *http.Client
	reqChannel    chan *http.Request
	stdoutChannel chan string
}

func makeClient(config *loadtestConfig) *client {
	return &client{
		httpClient: &http.Client{
			Timeout: time.Duration(int(time.Second) * config.httpTimeoutSecs),
		},
		reqChannel:    config.reqChannel,
		stdoutChannel: config.stdoutChannel,
	}
}

func (c *client) startWorking(ctx context.Context) {
	for req := range c.reqChannel {
		startTime := time.Now()
		resp, err := c.httpClient.Do(req)

		if err != nil {
			c.stdoutChannel <- err.Error()
			continue
		} else {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.stdoutChannel <- err.Error()
			}
			bodyString := string(bodyBytes)

			hashDuration, err := strconv.Atoi(bodyString)
			if err != nil {
				c.stdoutChannel <- err.Error()
			}
			durationMillis := int(time.Since(startTime) / time.Millisecond)
			c.stdoutChannel <- encodeResult(&result{
				success:             true,
				hashDurationMillis:  hashDuration,
				totalDurationMillis: durationMillis,
			})
		}

	}

	// check if the context is done, close out channel and stop
	select {
	case <-ctx.Done():
		return // this will close the "out" channel as well
	default:
	}
}
